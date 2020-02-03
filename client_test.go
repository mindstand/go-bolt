package goBolt

import (
	"github.com/mindstand/go-bolt/bolt_mode"
	"github.com/mindstand/go-bolt/log"
	"testing"
)

func TestClient(t *testing.T) {
	log.SetLevel("trace")
	log.Info("opening client")
	client, err := NewClient(WithBasicAuth("neo4j", "TZU6xiVZLbe5L5UmZaU5"), WithHostPort("0.0.0.0", 7687))
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	log.Infof("opening driver")
	driver, err := client.NewDriver()
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
	rows, err := conn.Query("create (:Fuck)", nil)
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

	log.Trace("closing connection")
	err = conn.Close()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}
