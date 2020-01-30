package goBolt

import (
	"context"
	"errors"
	"fmt"
	"github.com/mindstand/go-bolt/connection"
	"github.com/mindstand/go-bolt/encoding/encoding_v1"
	"github.com/mindstand/go-bolt/log"
	"math/rand"
	"time"
)

// sprintByteHex returns a formatted string of the byte array in hexadecimal
// with a nicely formatted human-readable output
func sprintByteHex(b []byte) string {
	output := ""
	for i, b := range b {
		output += fmt.Sprintf("%x", b)
		if (i+1)%16 == 0 {
			output += "\n\n"
		} else if (i+1)%4 == 0 {
			output += "  "
		} else {
			output += " "
		}
	}
	output += fmt.Sprintf("\n%x\n%s\n", b, string(b))

	return output
}

// driverArgsToMap turns internalDriver.Value list into a parameter map
// for neo4j parameters
func driverArgsToMap(args []Value) (map[string]interface{}, error) {
	output := map[string]interface{}{}
	for _, arg := range args {
		argBytes, ok := arg.([]byte)
		if !ok {
			return nil, errors.New("You must pass only a gob encoded map to the Exec/Query args")
		}

		m, err := encoding_v1.Unmarshal(argBytes)
		if err != nil {
			return nil, err
		}

		for k, v := range m.(map[string]interface{}) {
			output[k] = v
		}

	}

	return output, nil
}

func connectionNilOrClosed(conn IConnection) bool {
	if conn.getConnection() == nil { //nil check before attempting read
		return true
	}
	err := conn.getConnection().SetReadDeadline(time.Now())
	if err != nil {
		log.Error("Bad Connection state detected", err) //the error caught here could be a io.EOF or a timeout, either way we want to log the error & return true
		return true
	}

	zero := make([]byte, 0)

	_, err = conn.getConnection().Read(zero) //read zero bytes to validate connection is still alive
	if err != nil {
		log.Error("Bad Connection state detected", err) //the error caught here could be a io.EOF or a timeout, either way we want to log the error & return true
		return true
	}

	//check if there were any connection errors
	if conn.getConnErr() != nil {
		return true
	}

	return false
}

func getPoolFunc(connStrs []string, readonly bool) func(ctx context.Context) (interface{}, error) {
	return func(ctx context.Context) (interface{}, error) {
		if len(connStrs) == 0 {
			return nil, errors.New("no connection strings provided")
		}

		var i int

		if len(connStrs) == 1 {
			i = 0
		} else {
			i = rand.Intn(len(connStrs))
		}

		conn, err := connection.createConnection(connStrs[i])
		if err != nil {
			return nil, err
		}
		//set whether or not it is readonly
		conn.readOnly = readonly

		return conn, nil
	}
}
