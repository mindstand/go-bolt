package routing

import (
	"errors"
	"fmt"
	"github.com/mindstand/go-bolt/bolt_mode"
	"github.com/mindstand/go-bolt/connection"
	"github.com/mindstand/go-bolt/log"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type routingPool struct {
	userPassPart  string
	tlsInfo       string
	leaderConnStr string

	// access mutex
	mutex sync.RWMutex

	// running
	running *int32
	// is updating
	updating *int32

	// configuration
	totalConns      int
	minWriteIdle    int
	maxWriteIdle    int
	minReadIdle     int
	maxReadIdle     int
	refreshInterval time.Duration
	routingHandler  *boltRoutingHandler

	// lookups
	connStrLookup map[string][]*connectionPoolWrapper
	borrowedConns map[string]*connectionPoolWrapper
	heldConns     map[string]*connectionPoolWrapper

	// write stuff
	rwStrings  []string
	rwIndex    int
	writeQueue *Queue

	// read stuff
	readStrings []string
	readIndex   int
	readQueue   *Queue
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

	run := int32(0)
	update := int32(0)

	return &routingPool{
		userPassPart:    authPart,
		tlsInfo:         tlsInfo,
		leaderConnStr:   leaderConnStr,
		mutex:           sync.RWMutex{},
		running:         &run,
		updating:        &update,
		totalConns:      numConns,
		minWriteIdle:    0,
		maxWriteIdle:    0,
		minReadIdle:     0,
		maxReadIdle:     0,
		refreshInterval: refreshInterval,
		routingHandler: &boltRoutingHandler{
			Leaders:      []neoNodeConfig{},
			Followers:    []neoNodeConfig{},
			ReadReplicas: []neoNodeConfig{},
		},
		connStrLookup: map[string][]*connectionPoolWrapper{},
		borrowedConns: map[string]*connectionPoolWrapper{},
		heldConns:     map[string]*connectionPoolWrapper{},
		rwStrings:     []string{},
		rwIndex:       0,
		writeQueue:    NewQueue(),
		readStrings:   []string{},
		readIndex:     0,
		readQueue:     NewQueue(),
	}, nil
}

func (r *routingPool) removeConn(conn *connectionPoolWrapper) error {
	if conn == nil || conn.Connection == nil {
		return errors.New("invalid connection, can not be nil")
	}

	if !conn.Connection.ValidateOpen() {
		err := conn.Connection.Close()
		if err != nil {
			return err
		}
	}

	delete(r.heldConns, conn.Connection.GetConnectionId())
	return nil
}

func (r *routingPool) addWriteConn() error {
	// validate we have connection strings
	if r.rwStrings == nil || len(r.rwStrings) == 0 {
		return errors.New("no connection strings")
	}

	r.rwIndex++

	// make sure its not an overflow
	if r.rwIndex == len(r.rwStrings) {
		r.rwIndex = 0
	}

	// create the bolt connection
	connWrap, err := r.newConnection(bolt_mode.ReadMode, r.rwStrings[r.rwIndex])
	if err != nil {
		return err
	}

	// add it to the queue
	r.writeQueue.Enqueue(connWrap)

	return nil
}

func (r *routingPool) addReadConn() error {
	// validate we have connection strings
	if r.readStrings == nil || len(r.readStrings) == 0 {
		return errors.New("no connection strings")
	}

	r.readIndex++

	// make sure its not an overflow
	if r.readIndex == len(r.readStrings) {
		r.readIndex = 0
	}

	// create the bolt connection
	connWrap, err := r.newConnection(bolt_mode.ReadMode, r.readStrings[r.readIndex])
	if err != nil {
		return err
	}

	// add it to the queue
	r.readQueue.Enqueue(connWrap)

	return nil
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

func (r *routingPool) refreshConnections() {
	conn, err := connection.CreateBoltConn(r.leaderConnStr)
	if err != nil {
		log.Error(err)
		return
	}

	defer conn.Close()

	err = r.routingHandler.refreshClusterInfo(conn)

	newWriteStrs := r.routingHandler.getWriteConnectionStrings()
	newReadStrs := r.routingHandler.getReadOnlyConnectionString()

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
		newStr = r.addAuthInfoToConnStr(newStr)
		if !stringSliceContains(actualWrites, newStr) {
			actualWrites = append(actualWrites, newStr)
		}
	}

	for _, newStr := range newReadStrs {
		newStr = r.addAuthInfoToConnStr(newStr)
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
	}

	r.rwStrings = actualWrites
	r.readStrings = actualReads
}

func (r *routingPool) prune() {
	reads := []*connectionPoolWrapper{}
	writes := []*connectionPoolWrapper{}
	for _, connWrap := range r.readQueue.items {
		if connWrap.markForDeletion {
			err := connWrap.Connection.Close()
			if err != nil {
				log.Error(err)
				continue
			}
		} else {
			reads = append(reads, connWrap)
		}
	}

	for _, connWrap := range r.writeQueue.items {
		if connWrap.markForDeletion {
			err := r.removeConn(connWrap)
			if err != nil {
				log.Error(err)
				continue
			}
		} else {
			writes = append(writes, connWrap)
		}
	}

	r.writeQueue = NewQueueFromSlice(writes...)
	r.readQueue = NewQueueFromSlice(reads...)
}

func (r *routingPool) balance() {
	if r.readQueue.Size() < r.minReadIdle {
		for r.readQueue.Size() == r.minReadIdle {
			err := r.addReadConn()
			if err != nil {
				log.Error(err)
				continue
			}
		}
	}

	if r.writeQueue.Size() < r.minWriteIdle {
		for r.writeQueue.Size() == r.minWriteIdle {
			err := r.addReadConn()
			if err != nil {
				log.Error(err)
				continue
			}
		}
	}

	if r.writeQueue.Size() > r.maxWriteIdle {
		for r.writeQueue.Size() == r.maxWriteIdle {
			err := r.removeConn(r.writeQueue.Dequeue())
			if err != nil {
				log.Error(err)
				continue
			}
		}
	}

	if r.readQueue.Size() > r.maxReadIdle {
		for r.readQueue.Size() == r.maxReadIdle {
			err := r.removeConn(r.readQueue.Dequeue())
			if err != nil {
				log.Error(err)
				continue
			}
		}
	}
}

func (r *routingPool) refreshHandler() {
	for r.isRunning() {
		// block routine until interval is up
		<-time.After(r.refreshInterval)
		r.mutex.Lock()
		// refresh connections and mark dead conns
		r.refreshConnections()
		// remove dead conns
		r.prune()
		// balance pool to params
		r.balance()
		r.mutex.Unlock()
	}
}

func (r *routingPool) Start() error {
	if r.isRunning() {
		return errors.New("pool already running")
	}

	r.setRunning(true)

	if r.totalConns%2 == 0 {
		r.maxWriteIdle = r.totalConns / 2
		r.maxReadIdle = r.totalConns / 2
	} else {
		// write would have one more if the number of connections is odd
		r.maxWriteIdle = ((r.totalConns - 1) / 2) + 1
		r.maxReadIdle = (r.totalConns - 1) / 2
	}

	//conn, err := connection.CreateBoltConn(r.leaderConnStr)
	//if err != nil {
	//	return err
	//}
	//
	//defer conn.Close()

	// refresh connections and mark dead conns
	r.refreshConnections()

	for i := 0; i < r.maxWriteIdle; i++ {
		err := r.addWriteConn()
		if err != nil {
			return err
		}
	}

	for i := 0; i < r.maxReadIdle; i++ {
		err := r.addReadConn()
		if err != nil {
			return err
		}
	}

	go r.refreshHandler()

	return nil
}

func (r *routingPool) Stop() error {
	if !r.isRunning() {
		return errors.New("routingPool is not running")
	}

	r.setRunning(false)

	for r.readQueue.Size() != 0 {
		err := r.removeConn(r.readQueue.Dequeue())
		if err != nil {
			return err
		}
	}

	for r.writeQueue.Size() != 0 {
		err := r.removeConn(r.writeQueue.Dequeue())
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *routingPool) internalBorrow(queue *Queue) (*connectionPoolWrapper, error) {
	if queue == nil {
		return nil, errors.New("queue can not be nil")
	}

	conn := queue.Dequeue()
	if conn == nil {
		return nil, fmt.Errorf("stale queue")
	}

	if conn.markForDeletion {
		err := r.removeConn(conn)
		if err != nil {
			return nil, err
		}
		return r.internalBorrow(queue)
	}

	if !conn.Connection.ValidateOpen() {
		err := r.removeConn(conn)
		if err != nil {
			return nil, err
		}
		return r.internalBorrow(queue)
	}

	delete(r.heldConns, conn.Connection.GetConnectionId())
	r.borrowedConns[conn.Connection.GetConnectionId()] = conn

	return conn, nil
}

func (r *routingPool) BorrowRConnection() (connection.IConnection, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	connWrap, err := r.internalBorrow(r.readQueue)
	if err != nil {
		return nil, err
	}

	return connWrap.Connection, nil
}

func (r *routingPool) BorrowRWConnection() (connection.IConnection, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	connWrap, err := r.internalBorrow(r.writeQueue)
	if err != nil {
		return nil, err
	}

	return connWrap.Connection, nil
}

func (r *routingPool) isRunning() bool {
	return atomic.LoadInt32(r.running) == 1
}

func (r *routingPool) setRunning(b bool) {
	if b {
		atomic.StoreInt32(r.running, 1)
	} else {
		atomic.StoreInt32(r.running, 0)
	}
}

func (r *routingPool) isUpdating() bool {
	return atomic.LoadInt32(r.updating) == 1
}

func (r *routingPool) setUpdating(b bool) {
	if b {
		atomic.StoreInt32(r.updating, 1)
	} else {
		atomic.StoreInt32(r.updating, 0)
	}
}

func (r *routingPool) Reclaim(conn connection.IConnection) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.isRunning() {
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

	// check if the connection is dead, if it is discard and make a new one
	if !connWrap.Connection.ValidateOpen() {
		err := r.removeConn(connWrap)
		if err != nil {
			if connWrap.ConnType == bolt_mode.WriteMode {
				return r.addWriteConn()
			} else {
				return r.addReadConn()
			}
		}
	}

	if connWrap.ConnType == bolt_mode.WriteMode {
		r.writeQueue.Enqueue(connWrap)
	} else {
		r.readQueue.Enqueue(connWrap)
	}

	return nil
}

func (r *routingPool) addAuthInfoToConnStr(connStr string) string {

	if r.userPassPart == "" && r.tlsInfo == "" {
		return connStr
	}

	hostPort := strings.Replace(connStr, "bolt://", "", -1)
	return fmt.Sprintf("bolt://%s@%s%s", r.userPassPart, hostPort, r.tlsInfo)
}
