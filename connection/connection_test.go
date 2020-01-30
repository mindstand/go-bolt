package connection

import (
	"github.com/mindstand/go-bolt/log"
	"testing"
)

func TestConnection(t *testing.T) {
	log.SetLevel("trace")
	connStr := "bolt://neo4j:changeme@localhost:7687"

	conn, err := CreateBoltConn(connStr)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	rows, err := conn.Query("return 1;", nil)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	err = rows.Close()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	for i := 0; i < 4; i++ {
		t.Log("starting tx")
		tx, err := conn.Begin()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}

		res, err := tx.Exec("create (:TestNode{uuid:$val})", map[string]interface{}{
			"val": i,
		})
		if err != nil {
			t.Log(err)
			t.FailNow()
		}

		t.Log(res.Metadata())
		t.Log("committing tx")

		err = tx.Commit()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
	}

	err = conn.Close()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}