package routing

import (
	"fmt"
	"github.com/mindstand/go-bolt/bolt_mode"
	"github.com/mindstand/go-bolt/connection"
	"github.com/mindstand/go-bolt/errors"
	"github.com/mindstand/go-bolt/log"
	"strings"
)

//use this in testing https://github.com/graphaware/neo4j-casual-cluster-quickstart/blob/master/docker-compose.yml
//refer here https://github.com/cleishm/libneo4j-client/issues/26#issuecomment-388623799

const clusterOverview = "call dbms.cluster.overview()"
const neoV4SystemDb = "system"

type neoNodeType int

const (
	Leader neoNodeType = iota
	Follower
	ReadReplica
)

type neoNodeConfig struct {
	Id         string
	Addresses  []string
	Database   string
	Groups     []string
	BoltString string
	Action     bolt_mode.AccessMode
	Type       neoNodeType
}

type boltRoutingHandler struct {
	Leaders      []neoNodeConfig
	Followers    []neoNodeConfig
	ReadReplicas []neoNodeConfig
}

func (c *boltRoutingHandler) getReadConnectionStrings() []string {
	var connStrs []string

	if c.ReadReplicas != nil && len(c.ReadReplicas) != 0 {
		for _, node := range c.ReadReplicas {
			connStrs = append(connStrs, node.BoltString)
		}
	}

	if c.Followers != nil && len(c.Followers) != 0 {
		for _, node := range c.Followers {
			connStrs = append(connStrs, node.BoltString)
		}
	}

	return connStrs
}

func (c *boltRoutingHandler) getWriteConnectionStrings() []string {
	var connStrs []string

	if c.Leaders != nil && len(c.Leaders) != 0 {
		for _, node := range c.Leaders {
			connStrs = append(connStrs, node.BoltString)
		}
	}

	return connStrs
}

func (c *boltRoutingHandler) getReadOnlyConnectionString() []string {
	var connStrs []string

	if c.Followers != nil && len(c.Followers) != 0 {
		for _, node := range c.Followers {
			connStrs = append(connStrs, node.BoltString)
		}
	}

	if c.ReadReplicas != nil && len(c.ReadReplicas) != 0 {
		for _, node := range c.ReadReplicas {
			connStrs = append(connStrs, node.BoltString)
		}
	}

	return connStrs
}

func (c *boltRoutingHandler) refreshClusterInfo(conn connection.IConnection) error {
	if conn == nil {
		return errors.New("bolt connection can not be nil")
	}

	var rowResp connection.IRows
	var err error

	if conn.GetProtocolVersionNumber() == 4 {
		rowResp, err = conn.QueryWithDb(clusterOverview, nil, neoV4SystemDb)
	} else {
		rowResp, err = conn.Query(clusterOverview, nil)
	}

	if err != nil {
		return err
	}

	defer rowResp.Close()

	rows, _, err := rowResp.All()
	if err != nil {
		return err
	}

	var leaders []neoNodeConfig
	var followers []neoNodeConfig
	var readReplicas []neoNodeConfig

	for _, row := range rows {
		id, addresses, role, groups, database, err := c.parseRow(row)
		if err != nil {
			return err
		}

		var boltStr string

		for _, str := range addresses {
			if strings.Contains(str, "bolt") {
				boltStr = str
			}
		}

		if boltStr == "" {
			log.Trace("skipping id [%v] due to not having a bolt connection string", id)
			continue
		}

		nodeType, action := c.infoFromRoleString(role)

		node := neoNodeConfig{
			Id:         id,
			BoltString: boltStr,
			Addresses:  addresses,
			Database:   database,
			Groups:     groups,
			Action:     action,
			Type:       nodeType,
		}

		switch nodeType {
		case Leader:
			leaders = append(leaders, node)
			break
		case Follower:
			followers = append(followers, node)
			break
		case ReadReplica:
			readReplicas = append(readReplicas, node)
			break
		default:
			return fmt.Errorf("unknown node type [%v]", nodeType)
		}
	}

	c.Leaders = leaders
	c.Followers = followers
	c.ReadReplicas = readReplicas

	return nil
}

/*
	[0]   [1]   [2]   [3]     [4]
	id  address role groups database
*/
func (c *boltRoutingHandler) parseRow(row []interface{}) (id string, addresses []string, role string, groups []string, database string, err error) {
	//validate length is correct
	rowLen := len(row)
	if rowLen != 5 {
		return "", nil, "", nil, "", fmt.Errorf("invalid number of rows for query `%s`. %v != 5", clusterOverview, rowLen)
	}

	var ok bool

	id, ok = row[0].(string)
	if !ok {
		return "", nil, "", nil, "", errors.New("unable to parse id into string")
	}

	addresses, err = c.convertInterfaceToStringArr(row[1])
	if err != nil {
		return "", nil, "", nil, "", errors.New("unable to parse addresses into []string")
	}

	role, ok = row[2].(string)
	if !ok {
		return "", nil, "", nil, "", errors.New("unable to parse role into string")
	}

	groups, err = c.convertInterfaceToStringArr(row[3])
	if err != nil {
		return "", nil, "", nil, "", errors.New("unable to parse groups into []string")
	}

	database, ok = row[4].(string)
	if !ok {
		return "", nil, "", nil, "", errors.New("unable to parse database into string")
	}
	return id, addresses, role, groups, database, nil
}

func (c *boltRoutingHandler) convertInterfaceToStringArr(i interface{}) ([]string, error) {
	if i == nil {
		return nil, errors.New("iarr cannot be nil")
	}

	iarr, ok := i.([]interface{})
	if !ok {
		return nil, errors.New("unable to cast to []interface{}")
	}

	if len(iarr) == 0 {
		return []string{}, nil
	}

	arr := make([]string, len(iarr), cap(iarr))

	for k, v := range iarr {
		arr[k], ok = v.(string)
		if !ok {
			return nil, errors.New("unable to parse interface{} to string")
		}
	}

	return arr, nil
}

func (c *boltRoutingHandler) infoFromRoleString(s string) (neoNodeType, bolt_mode.AccessMode) {
	switch strings.ToLower(s) {
	case "leader":
		return Leader, bolt_mode.WriteMode
	case "follower":
		return Follower, bolt_mode.WriteMode
	case "read_replica":
		return ReadReplica, bolt_mode.ReadMode
	default:
		return -1, -1
	}
}
