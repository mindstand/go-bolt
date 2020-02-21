package goBolt

import (
	"github.com/mindstand/go-bolt/bolt_mode"
	"github.com/mindstand/go-bolt/log"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestClient(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	log.SetLevel("trace")
	log.Info("opening client")
	req := require.New(t)
	client, err := NewClient(WithBasicAuth("neo4j", "changeme"), WithHostPort("0.0.0.0", 7687))
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	log.Infof("opening driver")
	driver, err := client.NewDriverPool(2)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	log.Info("opening connection")
	conn, err := driver.Open(bolt_mode.WriteMode)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	//_, err = conn.ExecWithDb("create  (:TestNode{id:'asdfasd'})", nil, "neo4j")
	//req.Nil(err)

	rows, err := conn.QueryWithDb("match (n{id:'asdfasd'}) return n", nil, "neo4j")
	req.Nil(err)
	req.NotNil(rows)
	all, meta, err := rows.All()
	req.NotNil(all)
	req.NotNil(meta)
	req.Nil(err)
	req.Nil(rows.Close())
	//
	//log.Infof("executing query")
	//rows, err := conn.Exec("call db.indexes", nil)
	//if err != nil {
	//	t.Log(err)
	//	t.FailNow()
	//}
	//
	//log.Infof("starting transaction")
	//
	//tx, err := conn.Begin()
	//req.Nil(err)
	//req.NotNil(tx)
	//
	//log.Infof("dropping index")
	//_, err = tx.Exec("drop index on :TestNode(firstname)", nil)
	//req.Nil(err)
	//
	//log.Infof("committing tx")
	//req.Nil(tx.Commit())
	//
	//log.Infof("showing rows, %v", rows)
	////all, m, err := rows.All()
	////log.Infof("rows: %v, %v, %v", all, m, err)
	////
	////log.Trace("closing rows")
	////err = rows.Close()
	////if err != nil {
	////	t.Log(err)
	////	t.FailNow()
	////}
	//
	//req.Nil(driver.Reclaim(conn))
	//
	////conn, err = driver.Open(bolt_mode.WriteMode)
	////req.Nil(err)
	////req.NotNil(conn)
	////
	////tx, err := conn.Begin()
	////req.Nil(err)
	////req.NotNil(tx)
	////
	////res, err := tx.Exec("merge (:TestNode{num:$num})-[:TEST_EDGE]->(:TestNode{num:$num1})", map[string]interface{}{
	////	"num": 1,
	////	"num1": 2,
	////})
	////req.Nil(err)
	////numCr, ok := res.GetNodesCreated()
	////log.Info(numCr, ok)
	////req.Nil(tx.Commit())
	////
	////log.Trace("closing connection")
	////err = conn.Close()
	////if err != nil {
	////	t.Log(err)
	////	t.FailNow()
	////}

	log.Tracef("closing driver")
	err = driver.Close()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}
