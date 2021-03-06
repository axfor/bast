// Package snowflake provides a very simple Twitter snowflake generator and parser.
// https://github.com/bwmarrin/snowflake
package snowflake

import (
	"errors"
	"strconv"
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	// Epoch is set to the twitter snowflake epoch of Nov 04 2010 01:42:54 UTC in milliseconds
	// You may customize this to set a different epoch for your application.
	Epoch = int64(1288834974657)

	// NodeBits Number of bits to use for Node
	// Remember, you have a total 22 bits to share between Node/Step
	NodeBits uint8 = 8

	// StepBits Number of bits to use for Step
	// Remember, you have a total 22 bits to share between Node/Step
	StepBits uint8 = 14

	nodeMax   = int64(-1 ^ (-1 << NodeBits))
	nodeMask  = int64(nodeMax << StepBits)
	stepMask  = int64(-1 ^ (-1 << StepBits))
	timeShift = uint8(NodeBits + StepBits)
	nodeShift = uint8(StepBits)
)

// A Node struct holds the basic information needed for a snowflake generator
// node
type Node struct {
	nodeVal int64
	node    int64
	ni      unsafe.Pointer
	epoch   time.Time
}

type nodeItem struct {
	time int64
	step int64
}

// An ID is a custom type used for a snowflake ID.  This is used so we can
// attach methods onto the ID.
type ID int64

//var gmux sync.Mutex
//var gNode *Node

// NewNode returns a new snowflake node that can be used to generate snowflake
func NewNode(node uint8) (*Node, error) {
	if node < 0 || int64(node) > nodeMax {
		return nil, errors.New("Node number must be between 0 and " + strconv.FormatInt(nodeMax, 10))
	}
	n := nodeItem{}
	var curTime = time.Now()
	ne := &Node{
		node: int64(node),
		ni:   unsafe.Pointer(&n),
		// add time.Duration to curTime to make sure we use the monotonic clock if available
		epoch: curTime.Add(time.Unix(Epoch/1000, (Epoch%1000)*1000000).Sub(curTime)),
	}
	ne.nodeVal = ne.node << nodeShift
	return ne, nil
}

// Generate creates and returns a unique snowflake ID
func (n *Node) Generate() ID {
	return ID(n.GenerateWithInt64())
}

// ID creates and returns a unique snowflake ID
func (n *Node) ID() int64 {
	return n.GenerateWithInt64()
}

// GenerateWithInt64 creates and returns a unique snowflake ID
func (n *Node) GenerateWithInt64() int64 {
	var now, stop int64
	ok := false
	newItem := &nodeItem{}
	var ni *nodeItem
	var p unsafe.Pointer
	for {
		p = atomic.LoadPointer(&n.ni)
		ni = (*nodeItem)(p)
		now = time.Since(n.epoch).Nanoseconds() / 1000000
		if ni.time == now {
			newItem.step = (ni.step + 1) & stepMask
			if newItem.step == 0 {
				for now <= ni.time {
					now = time.Since(n.epoch).Nanoseconds() / 1000000
				}
			}
		} else {
			newItem.step = 0
		}
		newItem.time = now
		stop = newItem.step
		ok = atomic.CompareAndSwapPointer(&n.ni, p, unsafe.Pointer(newItem))
		if ok {
			return int64(now<<timeShift | n.nodeVal | stop)
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
