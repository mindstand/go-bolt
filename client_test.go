package goBolt

import (
	"github.com/mindstand/go-bolt/bolt_mode"
	"github.com/mindstand/go-bolt/log"
	"testing"
)

func TestClient(t *testing.T) {
	log.SetLevel("trace")
	log.Info("opening client")
	client, err := NewClient(WithBasicAuth("neo4j", "changeme"), WithHostPort("0.0.0.0", 7687), WithRouting())
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	log.Infof("opening driver")
	driver, err := client.NewDriverPool(55)
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

	log.Infof("executing query")
	rows, err := conn.Query("match (n) return n", nil)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	log.Infof("showing rows")
	all, m, err := rows.All()
	log.Tracef("rows: %v, %v, %v", all, m, err)

	log.Trace("closing rows")
	err = rows.Close()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	//req := require.New(t)
	//
	//conn, err = driver.Open(bolt_mode.WriteMode)
	//req.Nil(err)
	//req.NotNil(conn)
	//
	//tx, err := conn.Begin()
	//req.Nil(err)
	//req.NotNil(tx)
	//
	//res, err := tx.Exec("merge (:TestNode{num:$num})-[:TEST_EDGE]->(:TestNode{num:$num1})", map[string]interface{}{
	//	"num": 1,
	//	"num1": 2,
	//})
	//req.Nil(err)
	//numCr, ok := res.GetNodesCreated()
	//log.Info(numCr, ok)
	//req.Nil(tx.Commit())
	//
	//log.Trace("closing connection")
	//err = conn.Close()
	//if err != nil {
	//	t.Log(err)
	//	t.FailNow()
	//}

	log.Tracef("closing driver")
	err = driver.Close()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}
