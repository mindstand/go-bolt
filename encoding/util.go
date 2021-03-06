package encoding

import (
	"github.com/mindstand/go-bolt/errors"
	"github.com/mindstand/go-bolt/structures/graph"
)

func SliceInterfaceToString(from []interface{}) ([]string, error) {
	to := make([]string, len(from))
	for idx, item := range from {
		toItem, ok := item.(string)
		if !ok {
			return nil, errors.New("Expected string value. Got %T %+v", item, item)
		}
		to[idx] = toItem
	}
	return to, nil
}

func SliceInterfaceToInt(from []interface{}) ([]int, error) {
	to := make([]int, len(from))
	for idx, item := range from {
		to[idx] = int(item.(int64))
	}
	return to, nil
}

func SliceInterfaceToNode(from []interface{}) ([]graph.Node, error) {
	to := make([]graph.Node, len(from))
	for idx, item := range from {
		toItem, ok := item.(graph.Node)
		if !ok {
			return nil, errors.New("Expected Node value. Got %T %+v", item, item)
		}
		to[idx] = toItem
	}
	return to, nil
}

func SliceInterfaceToRelationship(from []interface{}) ([]graph.Relationship, error) {
	to := make([]graph.Relationship, len(from))
	for idx, item := range from {
		toItem, ok := item.(graph.Relationship)
		if !ok {
			return nil, errors.New("Expected Relationship value. Got %T %+v", item, item)
		}
		to[idx] = toItem
	}
	return to, nil
}

func SliceInterfaceToUnboundRelationship(from []interface{}) ([]graph.UnboundRelationship, error) {
	to := make([]graph.UnboundRelationship, len(from))
	for idx, item := range from {
		toItem, ok := item.(graph.UnboundRelationship)
		if !ok {
			return nil, errors.New("Expected UnboundRelationship value. Got %T %+v", item, item)
		}
		to[idx] = toItem
	}
	return to, nil
}
