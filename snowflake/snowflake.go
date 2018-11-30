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

// A NodeX struct holds the basic information needed for a snowflake generator
// node
type NodeX struct {
	node int64
	ni   *nodeItem
}

type nodeItem struct {
	time int64
	step int64
	temp *nodeItem
}

// A Node struct holds the basic information needed for a snowflake generator
// node
type Node struct {
	mu   sync.Mutex
	time int64
	node int64
	step int64
}

// An ID is a custom type used for a snowflake ID.  This is used so we can
// attach methods onto the ID.
type ID int64

var gmu sync.Mutex
var gmux sync.Mutex
var gNodeX *NodeX
var gNode *Node

// NewNodeX returns a new snowflake node that can be used to generate snowflake
// IDs
func NewNodeX(node int64) (*NodeX, error) {
	if gNodeX == nil {
		gmux.Lock()
		defer gmux.Unlock()
		if gNodeX != nil {
			return gNodeX, nil
		}
		// re-calc in case custom NodeBits or StepBits were set
		nodeMax = -1 ^ (-1 << NodeBits)
		nodeMask = nodeMax << StepBits
		stepMask = -1 ^ (-1 << StepBits)
		timeShift = NodeBits + StepBits
		nodeShift = StepBits

		if node < 0 || node > nodeMax {
			return nil, errors.New("Node number must be between 0 and " + strconv.FormatInt(nodeMax, 10))
		}

		gNodeX = &NodeX{
			// time: 0,
			node: node,
			ni:   &nodeItem{},
			// step: 0,
		}

	}
	return gNodeX, nil
}

// NewNode returns a new snowflake node that can be used to generate snowflake
// IDs
func NewNode(node int64) (*Node, error) {
	if gNode == nil {
		gmu.Lock()
		defer gmu.Unlock()
		if gNode != nil {
			return gNode, nil
		}
		// re-calc in case custom NodeBits or StepBits were set
		nodeMax = -1 ^ (-1 << NodeBits)
		nodeMask = nodeMax << StepBits
		stepMask = -1 ^ (-1 << StepBits)
		timeShift = NodeBits + StepBits
		nodeShift = StepBits

		if node < 0 || node > nodeMax {

			return nil, errors.New("Node number must be between 0 and " + strconv.FormatInt(nodeMax, 10))
		}

		gNode = &Node{
			// time: 0,
			node: node,
			// n:    &nodeItem{time: 0, step: 0},
			// step: 0,
		}
	}
	return gNode, nil
}

// GenerateX creates and returns a unique snowflake ID
func (n *NodeX) GenerateX() ID {
	var nowTime int64
	newItem := &nodeItem{time: n.ni.time, step: n.ni.step}
	ok := false
	for {
		// fmt.Printf("time=%d\n", n.ni.time)
		nowTime = time.Now().UnixNano() / 1000000
		newItem.time = n.ni.time
		newItem.step = n.ni.step
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
			// fmt.Printf("time2=%d\n", n.ni.time)
			// r :=
			return ID(((nowTime-Epoch)<<timeShift | (n.node << nodeShift) | (n.ni.step)))
		}
		// else {
		// 	// fmt.Printf("false,")
		// 	return ID(0)
		// }
	}
}

// Generate creates and returns a unique snowflake ID
func (n *Node) Generate() ID {

	n.mu.Lock()
	defer n.mu.Unlock()
	now := time.Now().UnixNano() / 1000000

	if n.time == now {
		n.step = (n.step + 1) & stepMask

		if n.step == 0 {
			for now <= n.time {
				now = time.Now().UnixNano() / 1000000
			}
		}
	} else {
		n.step = 0
	}

	n.time = now

	r := ID((now-Epoch)<<timeShift | (n.node << nodeShift) | (n.step))

	return r
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

//Clear 清空
func Clear() {
	if gNodeX != nil {
		if gNodeX.ni != nil {
			gNodeX.ni.temp = nil
		}
		gNodeX = nil
	}
}
