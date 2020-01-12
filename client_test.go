package goBolt

import (
	ll "github.com/mindstand/go-bolt/log"
	"log"
	"testing"
)

func TestClient(t *testing.T) {
	ll.SetLevel("trace")
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

func TestClientV4(t *testing.T) {
	ll.SetLevel("trace")
	client, err := NewClient(WithBasicAuth("neo4j", "changeme"), WithHostPort("0.0.0.0", 7687), WithVersion("4"))
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	driver, err := client.NewDriverV4()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	conn, err := driver.Open("system", ReadWriteMode)
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
