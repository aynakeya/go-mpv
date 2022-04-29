package mpv

/*
#include <stdlib.h>
#include <mpv/client.h>
#include <mpv_helper.h>

*/
import "C"
import "unsafe"

type Node struct {
	Value  interface{}
	Format Format
}

func newNode(node *C.mpv_node) Node {
	switch Format(node.format) {
	case FORMAT_NONE:
		return Node{nil, FORMAT_NONE}
	case FORMAT_STRING:
		return Node{C.GoString(C.get_node_string(node)), FORMAT_STRING}
	case FORMAT_FLAG:
		return Node{C.get_node_flag(node) == 1, FORMAT_FLAG}
	case FORMAT_INT64:
		return Node{int64(C.get_node_int64(node)), FORMAT_INT64}
	case FORMAT_DOUBLE:
		return Node{float64(C.get_node_double(node)), FORMAT_DOUBLE}
	case FORMAT_NODE_MAP:
		return Node{
			Format: FORMAT_NODE_MAP,
			Value:  newNodeMap(C.get_node_list(node)),
		}
	case FORMAT_NODE_ARRAY:
		return Node{
			Format: FORMAT_NODE_ARRAY,
			Value:  newNodeArray(C.get_node_list(node)),
		}
	case FORMAT_BYTE_ARRAY:
		panic("Not implement yet")
	default:
		panic("no such format")
	}
}

// CNode create a mpv_node in Go memory space, with pointer to C memory space
// must free after use
func (n Node) CNode() *C.mpv_node {
	var cnode = C.create_node(C.mpv_format(n.Format))
	switch n.Format {
	case FORMAT_NONE:
	case FORMAT_STRING:
		C.set_node_string(cnode, C.CString(n.Value.(string)))
	case FORMAT_FLAG:
		C.set_node_flag(cnode, boolToCInt(n.Value.(bool)))
	case FORMAT_INT64:
		C.set_node_int64(cnode, C.int64_t(n.Value.(int64)))
	case FORMAT_DOUBLE:
		C.set_node_double(cnode, C.double(n.Value.(float64)))
	case FORMAT_NODE_MAP:
		C.set_node_list(cnode, newCNodeMap(n.Value.(map[string]Node)))
	case FORMAT_NODE_ARRAY:
		C.set_node_list(cnode, newCNodeArray(n.Value.([]Node)))
	case FORMAT_BYTE_ARRAY:
		panic("Not implement yet")
	default:
		panic("no such format")
	}
	return cnode
}

// newNodeArray
func newNodeArray(nl *C.mpv_node_list) []Node {
	nodes := make([]Node, int(nl.num))
	for i := 0; i < int(nl.num); i++ {
		nodes[i] = newNode(C.get_node_list_element(nl, C.int(i)))
	}
	return nodes
}

// newNodeMap
func newNodeMap(nl *C.mpv_node_list) map[string]Node {
	nodes := make(map[string]Node)
	for i := 0; i < int(nl.num); i++ {
		nodes[C.GoString(C.get_node_list_key(nl, C.int(i)))] = newNode(C.get_node_list_element(nl, C.int(i)))
	}
	return nodes
}

func newCNodeArray(nodes []Node) *C.mpv_node_list {
	cNodeList := C.create_node_list(C.mpv_format(FORMAT_NODE_ARRAY), C.int(len(nodes)))
	for i, node := range nodes {
		cnode := node.CNode()
		C.set_node_list_element(cNodeList, C.int(i), cnode)
		C.free(unsafe.Pointer(cnode))
	}
	return cNodeList
}

func newCNodeMap(nodes map[string]Node) *C.mpv_node_list {
	cNodeList := C.create_node_list(C.mpv_format(FORMAT_NODE_MAP), C.int(len(nodes)))
	i := 0
	for key, node := range nodes {
		ckey := C.CString(key)
		cnode := node.CNode()
		C.set_node_list_element(cNodeList, C.int(i), cnode)
		C.set_node_list_key(cNodeList, C.int(i), ckey)
		C.free(unsafe.Pointer(cnode))
		C.free(unsafe.Pointer(ckey))
		i++
	}
	return cNodeList
}
