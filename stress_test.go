package goBolt

import (
	"fmt"
	"github.com/mindstand/go-bolt/bolt_mode"
	"github.com/mindstand/go-bolt/log"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestStress(t *testing.T) {
	log.SetLevel("info")
	req := require.New(t)

	client, err := NewClient(WithBasicAuth("neo4j", "changeme"), WithHostPort("0.0.0.0", 7687), WithRouting())
	req.Nil(err)
	req.NotNil(client)

	log.Infof("opening driver")
	driver, err := client.NewDriverPool(20)
	req.Nil(err)
	req.NotNil(driver)

	defer driver.Close()

	numThreads := 10
	var wg sync.WaitGroup
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go driverCycle(driver, req, i, &wg)
	}

	wg.Wait()
}

func driverCycle(driver IDriverPool, req *require.Assertions, iter int, wg *sync.WaitGroup) {
	log.Info("starting")
	read := false
	cycle := 0
	for cycle != 15000{
		if read {
			log.Infof("reading %v", iter)
			conn, err := driver.Open(bolt_mode.ReadMode)
			req.Nil(err)
			req.NotNil(conn)

			rows, err := conn.Query("match (n) return n limit 50", nil)
			req.Nil(err)
			req.NotNil(rows)
			res, meta, err := rows.All()
			req.Nil(err)
			req.NotNil(res)
			req.NotNil(meta)
			req.Nil(rows.Close())
			req.Nil(driver.Reclaim(conn))
			log.Infof("done reading %v", iter)
		} else {
			log.Infof("writing %v", iter)
			conn, err := driver.Open(bolt_mode.WriteMode)
			req.Nil(err)
			req.NotNil(conn)

			tx, err := conn.Begin()
			req.Nil(err)
			req.NotNil(tx)

			res, err := tx.Exec("merge (:TestNode{num:$num})", map[string]interface{}{
				"num": fmt.Sprintf("%v-%v", iter, cycle),
			})
			req.Nil(err)
			req.NotNil(res)
			req.Nil(tx.Commit())

			req.Nil(driver.Reclaim(conn))
			log.Infof("done writing %v", iter)
		}

		cycle++
		read = !read
	}
	log.Infof("completed [%v] cycles [%v]", cycle, iter)
	wg.Done()
}