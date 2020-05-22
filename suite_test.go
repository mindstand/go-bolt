package goBolt

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/mindstand/go-bolt/bolt_mode"
	"github.com/mindstand/go-bolt/connection"
	"github.com/mindstand/go-bolt/log"
	"github.com/stretchr/testify/suite"
)

const (
	createIndexV4Query   = `create index %s for (t:TestNode) on (t.id)`
	createIndexV1t3Query = `create index on :TestNode(id)`

	dropIndexV4Query   = `drop index %s`
	dropIndexV1t3Query = `drop index on :TestNode(id)`
)

func TestRunner(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	log.SetLevel("trace")
	var connectionString, db string
	var protocolVersion int
	var isCluster bool
	if os.Getenv("TEST_ACTIONS") == "true" {
		connectionString = os.Getenv("CONN_STR")
		connectionString = strings.Replace(connectionString, "\"", "", -1)
		connectionString = strings.TrimSpace(connectionString)
		db = os.Getenv("DB")
		pvs := os.Getenv("PVS")
		pvs64, err := strconv.ParseInt(pvs, 10, 64)
		if err != nil {
			t.Logf("failed reading env vars, %s", err.Error())
			t.FailNow()
		}
		protocolVersion = int(pvs64)

		isCluster, err = strconv.ParseBool(os.Getenv("IS_CLUSTER"))
		if err != nil {
			t.Logf("failed reading env vars, %s", err.Error())
			t.FailNow()
		}
		log.Info(connectionString)
		log.Info(db)
		log.Info(protocolVersion)
		log.Info(isCluster)
	} else {
		connectionString = "bolt+routing://neo4j:changeme@localhost:7687"
		protocolVersion = 3
		isCluster = true
	}

	log.Infof("starting integration test")

	suite.Run(t, &BoltTestSuite{
		protocolVersion:  protocolVersion,
		isCluster:        isCluster,
		connectionString: connectionString,
		db:               db,
	})
}

type BoltTestSuite struct {
	suite.Suite
	protocolVersion  int
	isCluster        bool
	connectionString string
	db               string
	client           IClient
	driverPool       IDriverPool
}

func (b *BoltTestSuite) SetupSuite() {
	client, err := NewClient(WithConnectionString(b.connectionString))
	b.Require().Nil(err)
	b.Require().NotNil(client)

	b.client = client
	b.driverPool, err = client.NewDriverPool(15)
	b.Require().Nil(err)
	b.Require().NotNil(b.driverPool)

	b.isCluster = strings.Contains(b.connectionString, "+routing")

	// create database to work out of
	if b.protocolVersion == 4 {
		b.db = "neo4j"
		//conn, err := b.driverPool.Open(bolt_mode.WriteMode)
		//b.Require().Nil(err)
		//b.Require().NotNil(conn)
		//
		//_, err = conn.ExecWithDb(fmt.Sprintf("create or replace database %s;", b.db), map[string]interface{}{}, "system")
		//b.Require().Nil(err)
		//
		//_, err = conn.ExecWithDb(fmt.Sprintf("start database %s;", b.db), map[string]interface{}{}, "system")
		//b.Require().Nil(err)
		//
		//<-time.After(15 * time.Second)
		//
		//b.Require().Nil(b.driverPool.Reclaim(conn))
	} else {
		b.db = ""
	}
}

func (b *BoltTestSuite) TearDownSuite() {
	// clean anything left
	conn, err := b.driverPool.Open(bolt_mode.WriteMode)
	b.Require().Nil(err)
	b.Require().NotNil(conn)
	_, err = conn.ExecWithDb("match (n) detach delete n", nil, b.db)
	b.Require().Nil(err)
	//if b.db != "" {
	//	_, err = conn.ExecWithDb(fmt.Sprintf("drop database %s;", b.db), nil, "system")
	//	b.Require().Nil(err)
	//}
	b.Require().Nil(b.driverPool.Reclaim(conn))
	b.Require().NotNil(b.driverPool)
	b.Require().Nil(b.driverPool.Close())
}

func (b *BoltTestSuite) TestConnectionRecycleClose() {
	poolSize := 6

	driver, err := b.client.NewDriverPool(poolSize)
	b.Require().Nil(err)
	b.Require().NotNil(driver)

	// should kill all of the pool
	for i := 0; i < poolSize; i++ {
		var mode bolt_mode.AccessMode
		if i%2 == 0 {
			mode = bolt_mode.WriteMode
		} else {
			mode = bolt_mode.ReadMode
		}
		conn, err := driver.Open(mode)
		b.Require().Nil(err)
		b.Require().NotNil(conn)
		b.Require().Nil(conn.Close())
		b.Require().Nil(driver.Reclaim(conn))
	}

	conn, err := driver.Open(bolt_mode.ReadMode)
	b.Require().Nil(err)
	b.Require().NotNil(conn)

	_, _, err = conn.Query("return 1", nil)
	b.Require().Nil(err)
}

func (b *BoltTestSuite) TestConnectionRecycleBrokenConnection() {
	poolSize := 5

	driver, err := b.client.NewDriverPool(poolSize)
	b.Require().Nil(err)
	b.Require().NotNil(driver)

	// should kill all of the pool
	for i := 0; i < poolSize*2; i++ {
		var mode bolt_mode.AccessMode
		if i%2 == 0 {
			mode = bolt_mode.WriteMode
		} else {
			mode = bolt_mode.ReadMode
		}
		conn, err := driver.Open(mode)
		b.Require().Nil(err)
		b.Require().NotNil(conn)
		_, _, err = conn.QueryWithDb("aasdfasdfasdfadfa", nil, b.db)
		b.Require().NotNil(err)
		b.Require().Nil(driver.Reclaim(conn))
	}

	conn, err := driver.Open(bolt_mode.ReadMode)
	b.Require().Nil(err)
	b.Require().NotNil(conn)

	_, _, err = conn.QueryWithDb("return 1", nil, b.db)
	b.Require().Nil(err)
}

func (b *BoltTestSuite) TestSingleDriver() {
	if b.isCluster {
		b.T().Skip("single driver is not compatible with cluster, so skipping test")
		return
	}

	driver, err := b.client.NewDriver()
	b.Require().Nil(err)
	b.Require().NotNil(driver)

	// note mode doesn't really matter here
	conn, err := driver.Open(bolt_mode.WriteMode)
	b.Require().Nil(err)
	b.Require().NotNil(conn)

	// run test
	b.connectionTest(conn, bolt_mode.WriteMode, b.db, "TestSingleDriver")
}

func (b *BoltTestSuite) TestPooledDriver() {
	// we're running on a cluster so we need to test readonly as well
	if strings.Contains(b.connectionString, "+routing") {
		conn, err := b.driverPool.Open(bolt_mode.ReadMode)
		b.Require().Nil(err)
		b.Require().NotNil(conn)
		b.connectionTest(conn, bolt_mode.ReadMode, b.db, "TestPooledDriver")
	}

	conn, err := b.driverPool.Open(bolt_mode.WriteMode)
	b.Require().Nil(err)
	b.Require().NotNil(conn)
	b.connectionTest(conn, bolt_mode.WriteMode, b.db, "TestPooledDriver")
}

// connectionTest runs the connection through tests that should exercise its functionality
// requests mode to see if the rejection works properly. Provide mode write on tests that aren't on casual clusters
func (b *BoltTestSuite) connectionTest(conn connection.IConnection, mode bolt_mode.AccessMode, db, testFrom string) {
	b.Require().NotNil(conn)
	// test basic query
	all, m, err := conn.QueryWithDb("return 1;", nil, db)
	b.Require().Nil(err)
	b.Require().NotNil(all)
	b.Require().NotNil(m)
	b.Require().Len(all, 1)
	b.Require().Len(all[0], 1)
	b.Require().Equal([][]interface{}{{
		int64(1),
	}}[0][0], all[0][0])

	// test basic exec
	res, err := conn.ExecWithDb("create (t:TestNode{id:$id}) return t", map[string]interface{}{
		"id": testFrom,
	}, db)
	if mode == bolt_mode.WriteMode {
		// test behavior if its allowed to do writes
		b.Require().Nil(err)
		b.Require().NotNil(res)
		nodesCr, ok := res.GetNodesCreated()
		b.Require().True(ok)
		b.Require().Equal(int64(1), nodesCr)
	} else {
		// test behavior if not allowed to do writes
		b.Require().NotNil(err)
		b.Require().Nil(res)
	}

	qid := fmt.Sprintf("%s-%v", testFrom, 1)

	// test create query
	if mode == bolt_mode.WriteMode {
		data, _, err := conn.QueryWithDb("create (:TestNode{id:$id})", map[string]interface{}{
			"id": qid,
		}, db)
		b.Require().Nil(err)
		b.Require().NotNil(data)
	} else {
		// test behavior if not allowed to do writes
		rows, _, err := conn.QueryWithDb("create (:TestNode{id:$id})", map[string]interface{}{
			"id": qid,
		}, db)
		b.Require().NotNil(err)
		b.Require().Nil(rows)
	}

	// after this point we can quit if its a readonly connection
	if mode == bolt_mode.ReadMode {
		return
	}

	// test delete exec
	res, err = conn.ExecWithDb("match (t:TestNode{id:$id}) delete t", map[string]interface{}{
		"id": testFrom,
	}, db)
	b.Require().Nil(err)
	b.Require().NotNil(res)
	nodesCr, ok := res.GetNodesDeleted()
	b.Require().True(ok)
	b.Require().Equal(int64(1), nodesCr)

	// test query
	data, _, err := conn.QueryWithDb("match (n) where n.id=$id return n", map[string]interface{}{
		"id": qid,
	}, db)
	b.Require().Nil(err)
	b.Require().NotNil(data)

	// test delete query
	data, _, err = conn.QueryWithDb("match (t:TestNode{id:$id}) delete t", map[string]interface{}{
		"id": qid,
	}, db)
	b.Require().Nil(err)
	b.Require().NotNil(data)

	// setup index stuff
	var indexCreateQuery, indexDeleteQuery string

	if b.protocolVersion == 4 {
		indexCreateQuery = fmt.Sprintf(createIndexV4Query, testFrom+"_index")
		indexDeleteQuery = fmt.Sprintf(dropIndexV4Query, testFrom+"_index")
	} else {
		indexCreateQuery = createIndexV1t3Query
		indexDeleteQuery = dropIndexV1t3Query
	}

	// test create index
	_, err = conn.ExecWithDb(indexCreateQuery, nil, db)
	b.Require().Nil(err)

	_, err = conn.ExecWithDb(indexDeleteQuery, nil, db)
	b.Require().Nil(err)

	// test create/read in tx
	tx, err := conn.BeginWithDatabase(db)
	b.Require().Nil(err)
	b.Require().NotNil(tx)

	qid = fmt.Sprintf("%s-%v", testFrom, 2)

	// test create query
	res, err = tx.ExecWithDb("create (:TestNode{id:$id})", map[string]interface{}{
		"id": qid,
	}, db)
	b.Require().Nil(err)
	b.Require().NotNil(res)
	nodesCr, ok = res.GetNodesCreated()
	b.Require().True(ok)
	b.Require().Equal(int64(1), nodesCr)

	data, _, err = tx.QueryWithDb("match (n) where n.id=$id return n", map[string]interface{}{
		"id": qid,
	}, db)
	b.Require().Nil(err)
	b.Require().NotNil(data)

	b.Require().Nil(tx.Commit())

	// test rollback
	tx, err = conn.BeginWithDatabase(db)
	b.Require().Nil(err)
	b.Require().NotNil(tx)

	res, err = tx.ExecWithDb("create (:TestNode{id:$id})", map[string]interface{}{
		"id": qid,
	}, db)
	b.Require().Nil(err)
	b.Require().NotNil(res)
	nodesCr, ok = res.GetNodesCreated()
	b.Require().True(ok)
	b.Require().Equal(int64(1), nodesCr)

	b.Require().Nil(tx.Rollback())
}
