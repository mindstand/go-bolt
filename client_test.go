package goBolt

import (
	"log"
	"testing"
)

func TestClient(t *testing.T) {
	client, err := NewClient(WithBasicAuth("neo4j", "changeme"), WithHostPort("0.0.0.0", 7687))
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	driver, err := client.NewDriver()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	conn, err := driver.Open(ReadWriteMode)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	rows, err := conn.QueryNeo("match (n) return n", nil)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	log.Println(rows.All())

	err = rows.Close()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}
