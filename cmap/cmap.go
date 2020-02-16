//Copyright 2018 The axx Authors. All rights reserved.

package cmap

import "github.com/easierway/concurrent_map"

// CMap is a thread safe map collection with better performance.
// The backend map entries are separated into the different partitions.
// Threads can access the different partitions safely without lock.
//
// inwrap for concurrent_map.ConcurrentMap
type CMap struct {
	innerMap *concurrent_map.ConcurrentMap
}

// Partitionable is the interface which should be implemented by key type.
// It is to define how to partition the entries.
//
// inwrap for concurrent_map.Partitionable
type Partitionable interface {
	// Value is raw value of the key
	Value() interface{}

	// PartitionKey is used for getting the partition to store the entry with the key.
	// E.g. the key's hash could be used as its PartitionKey
	// The partition for the key is partitions[(PartitionKey % m.numOfBlockets)]
	//
	// 1 Why not provide the default hash function for partition?
	// Ans: As you known, the partition solution would impact the performance significantly.
	// The proper partition solution balances the access to the different partitions and
	// avoid of the hot partition. The access mode highly relates to your business.
	// So, the better partition solution would just be designed according to your business.
	PartitionKey() int64
}

// New is to create a ConcurrentMap with the setting number of the partitions
//
// inwrap for concurrent_map.CreateConcurrentMap
func New(numOfPartitions int) *CMap {
	if numOfPartitions <= 0 {
		numOfPartitions = 99
	}
	return &CMap{
		innerMap: concurrent_map.CreateConcurrentMap(numOfPartitions),
	}
}

// Get is to get the value by the key
//
// inwrap for concurrent_map.CreateConcurrentMap
func (m *CMap) Get(key Partitionable) (interface{}, bool) {
	return m.innerMap.Get(key)
}

// Set is to store the KV entry to the map
//
// inwrap for concurrent_map.CreateConcurrentMap
func (m *CMap) Set(key Partitionable, v interface{}) {
	m.innerMap.Set(key, v)
}

// Del is to delete the entries by the key
//
// inwrap for concurrent_map.CreateConcurrentMap
func (m *CMap) Del(key Partitionable) {
	m.innerMap.Del(key)
}

// StrKey is to convert a string to StringKey
func StrKey(key string) *concurrent_map.StringKey {
	return concurrent_map.StrKey(key)
}

// I64Key is to convert a int64 to Int64Key
func I64Key(key int64) *concurrent_map.Int64Key {
	return concurrent_map.I64Key(key)
}

// IntKey is to convert a int to Int64Key
func IntKey(key int) *concurrent_map.Int64Key {
	return concurrent_map.I64Key(int64(key))
}
