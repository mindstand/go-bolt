[![Go Report Card](https://goreportcard.com/badge/github.com/mindstand/go-bolt)](https://goreportcard.com/report/github.com/mindstand/go-bolt)
[![GoDoc](https://godoc.org/github.com/mindstand/go-bolt?status.svg)](https://pkg.go.dev/github.com/mindstand/go-bolt)
![GoBolt Test Suite](https://github.com/mindstand/go-bolt/workflows/GoBolt%20Test%20Suite/badge.svg)
# Go-Bolt - GoLang Bolt Driver

Implements Neo4j Bolt Protocol Versions 1-4

```
go get -u github.com/mindstand/go-bolt
```

### (Disclaimer) This repository is still a major work in progress

## Features
- Supports bolt protocol versions 1-4
- Supports multi db in bolt protocol v4
- Connection Pooling
- `bolt+routing` for casual clusters
- TLS support

## Current todo's
#### (Issues will be updated)
- Documentation across entire repository
- Unit/integration testing across the entire repository for all protocol versions
- Support for neo4j bookmarks

## Long term goals
- Cypher checks preflight
- Benchmark Testing

## Thanks to:
- [@johnnadratowski](https://github.com/johnnadratowski): for your [original implementation](https://github.com/johnnadratowski/golang-neo4j-bolt-driver) of the bolt driver in go. We used the design as a basis for this driver!

## Current example
This will be changed, this is the main integration test at the moment

```go
client, err := NewClient(WithBasicAuth("neo4j", "changme"), WithHostPort("0.0.0.0", 7687))
if err != nil {
    panic(err)
}

driver, err := client.NewDriver()
if err != nil {
    panic(err)
}

conn, err := driver.Open(bolt_mode.WriteMode)
if err != nil {
    panic(err)
}

all, m, err := conn.Query("create (:TestNode{uuid:$id})", map[string]interface{}{
   "id": "random_id",
})
log.Tracef("rows: %v, %v, %v", all, m, err)

err = conn.Close()
if err != nil {
    panic(err)
}
```
