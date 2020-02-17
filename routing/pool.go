package routing

import (
	"fmt"
	"github.com/mindstand/go-bolt/bolt_mode"
	"github.com/mindstand/go-bolt/connection"
	"github.com/mindstand/go-bolt/errors"
	"github.com/mindstand/go-bolt/log"
	"strings"
	"sync"
	"time"
)

type routingPool struct {
	userPassPart  string
	tlsInfo       string
	leaderConnStr string
	// for general access
	mutex sync.Mutex
	// for updating stuff
	updateMutex sync.Mutex

	numConns int

	refreshInterval time.Duration
	routingHandler  *boltRoutingHandler

	readStrings []string
	readI       int
	readMutex   sync.Mutex

	rwStrings []string
	rwI       int
	rwMutex   sync.RWMutex

	connStrLookup map[string][]*connectionPoolWrapper

	borrowedConns map[string]*connectionPoolWrapper
	heldConns     map[string]*connectionPoolWrapper

	writeStack *connectionStack
	readStack  *connectionStack

	exitRefreshChan chan bool
	exitListenChan  chan bool
	running         bool
}

func NewRoutingPool(leaderConnStr string, numConns int, refreshInterval time.Duration, authPart, tlsInfo string) (IRoutingPool, error) {
	if leaderConnStr == "" {
		return nil, errors.New("leaderConnStr can not be nil")
	}

	if strings.Contains(leaderConnStr, "+routing") {
		leaderConnStr = strings.Replace(leaderConnStr, "+routing", "", -1)
	}

	if numConns == 0 || numConns == 1 {
		return nil, fmt.Errorf("number of connections must be greater than 1, provided [%v]", numConns)
	}

	return &routingPool{
		userPassPart:    authPart,
		tlsInfo:         tlsInfo,
		leaderConnStr:   leaderConnStr,
		mutex:           sync.Mutex{},
		updateMutex:     sync.Mutex{},
		numConns:        numConns,
		refreshInterval: refreshInterval,
		routingHandler: &boltRoutingHandler{
			Leaders:      []neoNodeConfig{},
			Followers:    []neoNodeConfig{},
			ReadReplicas: []neoNodeConfig{},
		},
		readStrings:     []string{},
		readI:           0,
		readMutex:       sync.Mutex{},
		rwStrings:       []string{},
		rwI:             0,
		rwMutex:         sync.RWMutex{},
		connStrLookup:   map[string][]*connectionPoolWrapper{},
		borrowedConns:   map[string]*connectionPoolWrapper{},
		heldConns:       map[string]*connectionPoolWrapper{},
		writeStack:      newStack(),
		readStack:       newStack(),
		exitRefreshChan: make(chan bool, 1),
		exitListenChan:  make(chan bool, 1),
		running:         false,
	}, nil
}

func (r *routingPool) addAuthInfoToConnStr(connStr string) string {

	if r.userPassPart == "" && r.tlsInfo == "" {
		return connStr
	}

	hostPort := strings.Replace(connStr, "bolt://", "", -1)
	return fmt.Sprintf("bolt://%s@%s%s", r.userPassPart, hostPort, r.tlsInfo)
}

func (r *routingPool) Start() error {
	if r.running {
		return errors.New("pool is already running")
	}

	writeConns := 0
	readConns := 0
	if r.numConns%2 == 0 {
		writeConns = r.numConns / 2
		readConns = r.numConns / 2
	} else {
		// write would have one more if the number of connections is odd
		writeConns = ((r.numConns - 1) / 2) + 1
		readConns = (r.numConns - 1) / 2
	}

	conn, err := connection.CreateBoltConn(r.leaderConnStr)
	if err != nil {
		return err
	}

	defer conn.Close()

	err = r.routingHandler.refreshClusterInfo(conn)
	if err != nil {
		return err
	}

	r.rwStrings = r.routingHandler.getWriteConnectionStrings()
	r.readStrings = r.routingHandler.getReadConnectionStrings()

	// create write conns
	for i := 0; i < writeConns; i++ {
		connStr, err := r.nextWriteConnectionString()
		if err != nil {
			return err
		}

		connStr = r.addAuthInfoToConnStr(connStr)

		writeConn, err := r.newConnection(bolt_mode.WriteMode, connStr)
		if err != nil {
			return err
		}

		if _, ok := r.connStrLookup[connStr]; !ok {
			r.connStrLookup[connStr] = []*connectionPoolWrapper{writeConn}
		} else {
			r.connStrLookup[connStr] = append(r.connStrLookup[connStr], writeConn)
		}

		err = r.writeStack.Push(writeConn)
		if err != nil {
			return err
		}
	}

	// create read conns
	for i := 0; i < readConns; i++ {
		connStr, err := r.nextReadConnectionString()
		if err != nil {
			return err
		}

		connStr = r.addAuthInfoToConnStr(connStr)

		readConn, err := r.newConnection(bolt_mode.ReadMode, connStr)
		if err != nil {
			return err
		}

		if _, ok := r.connStrLookup[connStr]; !ok {
			r.connStrLookup[connStr] = []*connectionPoolWrapper{readConn}
		} else {
			r.connStrLookup[connStr] = append(r.connStrLookup[connStr], readConn)
		}

		err = r.readStack.Push(readConn)
		if err != nil {
			return err
		}
	}

	// start refresh watcher
	go r.refreshWatcher()
	// start delete listener
	go r.deletedConnectionsHandler()

	r.running = true
	return nil
}

func (r *routingPool) Stop() error {
	if !r.running {
		return errors.New("routing pool not running")
	}

	r.exitRefreshChan <- true
	r.exitListenChan <- true

	// close all connections
	for _, conn := range r.borrowedConns {
		log.Tracef("closing %v", conn)
		if conn.Connection != nil {
			conn.Connection.Close()
		}
	}
	for _, conn := range r.heldConns {
		log.Tracef("closing %v", conn)
		if conn.Connection != nil {
			conn.Connection.Close()
		}
	}

	return nil
}

func (r *routingPool) refreshWatcher() {
	for {
		timer := time.After(r.refreshInterval)

		select {
		case <-r.exitRefreshChan:
			return
		case <-timer:
			var wg sync.WaitGroup
			r.updateMutex.Lock()
			wg.Add(1)
			go r.refreshConnections(&wg)
			wg.Wait()
			r.updateMutex.Unlock()
		}
	}
}

// uses routing handler to check if any connections should be dumped
func (r *routingPool) refreshConnections(wg *sync.WaitGroup) {
	defer wg.Done()

	conn, err := connection.CreateBoltConn(r.leaderConnStr)
	if err != nil {
		log.Error(err)
		return
	}

	defer conn.Close()

	err = r.routingHandler.refreshClusterInfo(conn)
	if err != nil {
		log.Error(err)
		return
	}

	newWriteStrs := r.routingHandler.getWriteConnectionStrings()
	newReadStrs := r.routingHandler.getReadConnectionStrings()

	var deadStrs, actualWrites, actualReads []string

	// check for dead strings
	for _, curStr := range r.rwStrings {
		// string is not found, so it must be dead
		if !stringSliceContains(newWriteStrs, curStr) {
			deadStrs = append(deadStrs, curStr)
		} else {
			actualWrites = append(actualWrites, curStr)
		}
	}

	for _, curStr := range r.readStrings {
		// string is not found, so it must be dead
		if !stringSliceContains(newReadStrs, curStr) {
			deadStrs = append(deadStrs, curStr)
		} else {
			actualReads = append(actualReads, curStr)
		}
	}

	// add anything new
	for _, newStr := range newWriteStrs {
		if !stringSliceContains(actualWrites, newStr) {
			actualWrites = append(actualWrites, newStr)
		}
	}

	for _, newStr := range newReadStrs {
		if !stringSliceContains(actualReads, newStr) {
			actualReads = append(actualReads, newStr)
		}
	}

	// mark them for deletion, system will rebalance elsewhere
	if len(deadStrs) != 0 {
		for _, str := range deadStrs {
			conns, ok := r.connStrLookup[str]
			if !ok || conns == nil || len(conns) == 0 {
				continue
			}

			for _, markedConn := range conns {
				if markedConn == nil {
					continue
				}

				markedConn.markForDeletion = true
			}
		}

		// tell the stacks to delete it
		r.writeStack.PruneMarkedConnections()
		r.readStack.PruneMarkedConnections()
	}
}

func (r *routingPool) nextReadConnectionString() (string, error) {
	r.readMutex.Lock()
	defer r.readMutex.Unlock()

	if r.readStrings == nil || len(r.readStrings) == 0 {
		return "", errors.New("no connection strings")
	}

	// increment r.readI
	r.readI++

	//
	if r.readI == len(r.readStrings) {
		r.readI = 0
	}

	return r.readStrings[r.readI], nil
}

func (r *routingPool) nextWriteConnectionString() (string, error) {
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()

	if r.rwStrings == nil || len(r.rwStrings) == 0 {
		return "", errors.New("no connection strings")
	}

	// increment r.readI
	r.rwI++

	//
	if r.rwI == len(r.rwStrings) {
		r.rwI = 0
	}

	return r.rwStrings[r.rwI], nil
}

func (r *routingPool) makeConnId(connType bolt_mode.AccessMode) string {
	return fmt.Sprintf("%v-%s", connType, stringWithCharset(50, charset))
}

func (r *routingPool) newConnection(connType bolt_mode.AccessMode, connStr string) (*connectionPoolWrapper, error) {
	conn, err := connection.CreateBoltConn(connStr)
	if err != nil {
		return nil, err
	}

	conn.SetConnectionId(r.makeConnId(connType))

	connWrap := &connectionPoolWrapper{
		Connection:      conn,
		ConnStr:         connStr,
		ConnType:        connType,
		borrowed:        false,
		numBorrows:      0,
		markForDeletion: false,
	}

	r.heldConns[conn.GetConnectionId()] = connWrap
	return connWrap, nil
}

// listen for bad connections
// todo better balancing
func (r *routingPool) deletedConnectionsHandler() {
	for {
		select {
		case rConn := <-r.readStack.connRemovedDelegate:
			err := rConn.Connection.Close()
			if err != nil {
				log.Error(err.Error())
				break
			}

			// remove connection from lookup table
			if _, ok := r.heldConns[rConn.Connection.GetConnectionId()]; ok {
				delete(r.heldConns, rConn.Connection.GetConnectionId())
			}

			// remove from connstr lookup table
			if _, ok := r.connStrLookup[rConn.ConnStr]; ok {
				// find the obj and remove it
				for i, conn := range r.connStrLookup[rConn.ConnStr] {
					if conn == nil {
						continue
					}

					// check if this is the one we're looking for
					if conn.Connection.GetConnectionId() == rConn.Connection.GetConnectionId() {
						// remove it from the slice
						copy(r.connStrLookup[rConn.ConnStr][i:], r.connStrLookup[rConn.ConnStr][i+1:])
						r.connStrLookup[rConn.ConnStr][len(r.connStrLookup[rConn.ConnStr])-1] = nil
						r.connStrLookup[rConn.ConnStr] = r.connStrLookup[rConn.ConnStr][:len(r.connStrLookup[rConn.ConnStr])-1]
						break
					}
				}
			}

			str, err := r.nextReadConnectionString()
			if err != nil {
				log.Error(err.Error())
				break
			}

			conn, err := r.newConnection(bolt_mode.ReadMode, str)
			if err != nil {
				log.Error(err.Error())
				break
			}

			err = r.readStack.Push(conn)
			if err != nil {
				log.Error(err.Error())
			}
			break
		case wConn := <-r.writeStack.connRemovedDelegate:
			err := wConn.Connection.Close()
			if err != nil {
				log.Error(err.Error())
				break
			}

			// remove connection from lookup table
			if _, ok := r.heldConns[wConn.Connection.GetConnectionId()]; ok {
				delete(r.heldConns, wConn.Connection.GetConnectionId())
			}

			// remove from connstr lookup table
			if _, ok := r.connStrLookup[wConn.ConnStr]; ok {
				// find the obj and remove it
				for i, conn := range r.connStrLookup[wConn.ConnStr] {
					if conn == nil {
						continue
					}

					// check if this is the one we're looking for
					if conn.Connection.GetConnectionId() == wConn.Connection.GetConnectionId() {
						// remove it from the slice
						copy(r.connStrLookup[wConn.ConnStr][i:], r.connStrLookup[wConn.ConnStr][i+1:])
						r.connStrLookup[wConn.ConnStr][len(r.connStrLookup[wConn.ConnStr])-1] = nil
						r.connStrLookup[wConn.ConnStr] = r.connStrLookup[wConn.ConnStr][:len(r.connStrLookup[wConn.ConnStr])-1]
						break
					}
				}
			}

			str, err := r.nextWriteConnectionString()
			if err != nil {
				log.Error(err.Error())
				break
			}

			conn, err := r.newConnection(bolt_mode.WriteMode, str)
			if err != nil {
				log.Error(err.Error())
				break
			}

			err = r.writeStack.Push(conn)
			if err != nil {
				log.Error(err.Error())
			}
			break
		case <-r.exitListenChan:
			return
		}
	}
}

func (r *routingPool) BorrowRConnection() (connection.IConnection, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.running {
		return nil, errors.New("pool is not running")
	}

	conn, err := r.readStack.Pop()
	if err != nil {
		return nil, errors.New("error retrieving write connection")
	}

	connId := conn.Connection.GetConnectionId()

	if _, ok := r.heldConns[connId]; ok {
		delete(r.heldConns, connId)
		r.borrowedConns[connId] = conn
	}

	return conn.Connection, nil
}

func (r *routingPool) BorrowRWConnection() (connection.IConnection, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.running {
		return nil, errors.New("pool is not running")
	}

	conn, err := r.writeStack.Pop()
	if err != nil {
		return nil, err
	}

	connId := conn.Connection.GetConnectionId()

	if _, ok := r.heldConns[connId]; ok {
		delete(r.heldConns, connId)
		r.borrowedConns[connId] = conn
	}

	return conn.Connection, nil
}

func (r *routingPool) Reclaim(conn connection.IConnection) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.running {
		return errors.New("pool is not running")
	}

	connId := conn.GetConnectionId()

	connWrap, ok := r.borrowedConns[connId]
	if !ok {
		err := conn.Close()
		if err != nil {
			return fmt.Errorf("connection not found with id [%s]. Also had error closing connection, %w", connId, err)
		}
		return fmt.Errorf("connection not found with id [%s]", connId)
	}

	// remove from borrowed and replace in held
	delete(r.borrowedConns, connId)
	r.heldConns[connId] = connWrap

	if connWrap.ConnType == bolt_mode.WriteMode {
		return r.writeStack.Push(connWrap)
	} else {
		return r.readStack.Push(connWrap)
	}
}
