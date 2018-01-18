package gnmi

import (
	"fmt"
	"strconv"

	log "github.com/golang/glog"
	"github.com/openconfig/goyang/pkg/yang"

	pb "github.com/openconfig/gnmi/proto/gnmi"
)

// getChildNode gets a node's child with corresponding schema specified by path
// element. If not found and createIfNotExist is set as true, an empty node is
// created and returned.
func getChildNode(node map[string]interface{}, schema *yang.Entry, elem *pb.PathElem, createIfNotExist bool) (interface{}, *yang.Entry) {
	var nextSchema *yang.Entry
	var ok bool

	if nextSchema, ok = schema.Dir[elem.Name]; !ok {
		return nil, nil
	}

	var nextNode interface{}
	if elem.GetKey() == nil {
		if nextNode, ok = node[elem.Name]; !ok {
			if createIfNotExist {
				node[elem.Name] = make(map[string]interface{})
				nextNode = node[elem.Name]
			}
		}
		return nextNode, nextSchema
	}

	nextNode = getKeyedListEntry(node, elem, createIfNotExist)
	return nextNode, nextSchema
}

// getKeyedListEntry finds the keyed list entry in node by the name and key of
// path elem. If entry is not found and createIfNotExist is true, an empty entry
// will be created (the list will be created if necessary).
func getKeyedListEntry(node map[string]interface{}, elem *pb.PathElem, createIfNotExist bool) map[string]interface{} {
	curNode, ok := node[elem.Name]
	if !ok {
		if !createIfNotExist {
			return nil
		}

		// Create a keyed list as node child and initialize an entry.
		m := make(map[string]interface{})
		for k, v := range elem.Key {
			m[k] = v
			if vAsNum, err := strconv.ParseFloat(v, 64); err == nil {
				m[k] = vAsNum
			}
		}
		node[elem.Name] = []interface{}{m}
		return m
	}

	// Search entry in keyed list.
	keyedList, ok := curNode.([]interface{})
	if !ok {
		return nil
	}
	for _, n := range keyedList {
		m, ok := n.(map[string]interface{})
		if !ok {
			log.Errorf("wrong keyed list entry type: %T", n)
			return nil
		}
		keyMatching := true
		// must be exactly match
		for k, v := range elem.Key {
			attrVal, ok := m[k]
			if !ok {
				return nil
			}
			if v != fmt.Sprintf("%v", attrVal) {
				keyMatching = false
				break
			}
		}
		if keyMatching {
			return m
		}
	}
	if !createIfNotExist {
		return nil
	}

	// Create an entry in keyed list.
	m := make(map[string]interface{})
	for k, v := range elem.Key {
		m[k] = v
		if vAsNum, err := strconv.ParseFloat(v, 64); err == nil {
			m[k] = vAsNum
		}
	}
	node[elem.Name] = append(keyedList, m)
	return m
}
