// Package snowflake provides a very simple Twitter snowflake generator and parser.
// https://github.com/bwmarrin/snowflake
package snowflake

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	// Epoch is set to the twitter snowflake epoch of Nov 04 2010 01:42:54 UTC
	// You may customize this to set a different epoch for your application.
	Epoch = int64(1288834974657)

	// NodeBits Number of bits to use for Node
	// Remember, you have a total 22 bits to share between Node/Step
	NodeBits uint8 = 8

	// StepBits Number of bits to use for Step
	// Remember, you have a total 22 bits to share between Node/Step
	StepBits uint8 = 14

	nodeMax   int64 = -1 ^ (-1 << NodeBits)
	nodeMask        = int64(nodeMax << StepBits)
	stepMask  int64 = -1 ^ (-1 << StepBits)
	timeShift       = uint8(NodeBits + StepBits)
	nodeShift       = uint8(StepBits)
)

// A JSONSyntaxError is returned from UnmarshalJSON if an invalid ID is provided.
type JSONSyntaxError struct{ original []byte }

func (j JSONSyntaxError) Error() string {
	return fmt.Sprintf("invalid snowflake ID %q", string(j.original))
}

// Create a map for decoding Base58.  This speeds up the process tremendously.

// A Node struct holds the basic information needed for a snowflake generator
// node
type Node struct {
	node int64
	ni   *nodeItem
}

type nodeItem struct {
	time int64
	step int64
	temp *nodeItem
}

// An ID is a custom type used for a snowflake ID.  This is used so we can
// attach methods onto the ID.
type ID int64

var gmux sync.Mutex
var gNode *Node

// NewNode returns a new snowflake node that can be used to generate snowflake
// IDs
func NewNode(node uint8) (*Node, error) {
	if gNode == nil {
		gmux.Lock()
		defer gmux.Unlock()
		if gNode != nil {
			return gNode, nil
		}
		// re-calc in case custom NodeBits or StepBits were set
		nodeMax = -1 ^ (-1 << NodeBits)
		nodeMask = nodeMax << StepBits
		stepMask = -1 ^ (-1 << StepBits)
		timeShift = NodeBits + StepBits
		nodeShift = StepBits

		if node < 0 || int64(node) > nodeMax {
			return nil, errors.New("Node number must be between 0 and " + strconv.FormatInt(nodeMax, 10))
		}

		gNode = &Node{
			node: int64(node),
			ni:   &nodeItem{},
		}

	}
	return gNode, nil
}

// Generate creates and returns a unique snowflake ID
func (n *Node) Generate() ID {
	var nowTime int64
	newItem := &nodeItem{time: n.ni.time, step: n.ni.step}
	ok := false
	for {
		newItem.time = n.ni.time
		newItem.step = n.ni.step
		nowTime = time.Now().UnixNano() / 1000000
		if newItem.time == nowTime {
			newItem.step = (newItem.step + 1) & stepMask
			if newItem.step == 0 {
				for nowTime <= n.ni.time {
					nowTime = time.Now().UnixNano() / 1000000
				}
			}
		} else {
			newItem.step = 0
		}
		newItem.time = nowTime
		newItem.temp = newItem
		preNi := n.ni
		ok = atomic.CompareAndSwapUintptr((*uintptr)(unsafe.Pointer(&n.ni)), uintptr(unsafe.Pointer(n.ni)), uintptr(unsafe.Pointer(newItem)))
		if ok {
			preNi.temp = nil
			preNi = nil
			return ID(((nowTime-Epoch)<<timeShift | (n.node << nodeShift) | newItem.step))
		}
	}
}

// Int64 returns an int64 of the snowflake ID
func (f ID) Int64() int64 {
	return int64(f)
}

// String returns a string of the snowflake ID
func (f ID) String() string {
	return strconv.FormatInt(int64(f), 10)
}

// Bytes returns a byte slice of the snowflake ID
func (f ID) Bytes() []byte {
	return []byte(f.String())
}

// IntBytes returns an array of bytes of the snowflake ID, encoded as a
// big endian integer.
func (f ID) IntBytes() [8]byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(f))
	return b
}

// Time returns an int64 unix timestamp of the snowflake ID time
func (f ID) Time() int64 {
	return (int64(f) >> timeShift) + Epoch
}

// Node returns an int64 of the snowflake ID node number
func (f ID) Node() int64 {
	return int64(f) & nodeMask >> nodeShift
}

// Step returns an int64 of the snowflake step (or sequence) number
func (f ID) Step() int64 {
	return int64(f) & stepMask
}

// MarshalJSON returns a json byte array string of the snowflake ID.
func (f ID) MarshalJSON() ([]byte, error) {
	buff := make([]byte, 0, 22)
	buff = append(buff, '"')
	buff = strconv.AppendInt(buff, int64(f), 10)
	buff = append(buff, '"')
	return buff, nil
}

// UnmarshalJSON converts a json byte array of a snowflake ID into an ID type.
func (f *ID) UnmarshalJSON(b []byte) error {
	if len(b) < 3 || b[0] != '"' || b[len(b)-1] != '"' {
		return JSONSyntaxError{b}
	}

	i, err := strconv.ParseInt(string(b[1:len(b)-1]), 10, 64)
	if err != nil {
		return err
	}

	*f = ID(i)
	return nil
}

//Clear gNode
func Clear() {
	if gNode != nil {
		if gNode.ni != nil {
			gNode.ni.temp = nil
		}
		gNode = nil
	}
}
