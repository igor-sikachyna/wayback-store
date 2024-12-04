package store

import (
	"testing"
	"wayback-store/chain"

	"github.com/stretchr/testify/assert"
)

func TestCreateStore(t *testing.T) {
	var assert = assert.New(t)

	var c = &chain.Chain{}
	var s, err = MakeStore(c)

	assert.Equal(nil, err, "failed to create a store")

	s.Write("a", "1")
	assert.Equal(1, s.root.size, "incorrect root node size")
	assert.Equal(1, s.root.block, "incorrect root node block")
	assert.Equal("a", s.root.data.Key, "incorrect root node key")
	assert.Equal("1", s.root.data.Value, "incorrect root node value")
	// sha256 of a|1
	assert.Equal("df4504ce92500fe8a20fb37090e36afa9053480c3e8dde4f4ae3fab4f5d8b1d7", s.root.hash, "incorrect root node hash")

	assert.Equal("1", s.Get("a"), "incorrect table value returned")
	assert.Equal("", s.Get("invalid"), "did not fail to get a non-existing value")
}

func TestStoreTree(t *testing.T) {
	var assert = assert.New(t)

	var c = &chain.Chain{}
	var s, _ = MakeStore(c)

	s.Write("a", "1")
	s.Write("b", "2")
	s.Write("c", "3")
	s.Write("d", "4")

	// The tree should undergo the following transitions:
	// a+1
	//
	//  |-root-|
	// a+1    b+2
	//
	//     |----- root ----|
	//  |--x--|         |--y--|
	// a+1   c+3       b+2   d+4

	assert.Equal(4, s.root.size, "incorrect root node size")
	assert.Equal(2, s.root.left.size, "incorrect x node size")
	assert.Equal(2, s.root.left.size, "incorrect y node size")
	assert.Equal(1, s.root.left.left.size, "incorrect a node size")
	assert.Equal(1, s.root.left.right.size, "incorrect c node size")
	assert.Equal(1, s.root.right.left.size, "incorrect b node size")
	assert.Equal(1, s.root.right.right.size, "incorrect d node size")

	assert.Equal("a", s.root.left.left.data.Key, "incorrect node key 1")
	assert.Equal("b", s.root.right.left.data.Key, "incorrect node key 2")
	assert.Equal("c", s.root.left.right.data.Key, "incorrect node key 3")
	assert.Equal("d", s.root.right.right.data.Key, "incorrect node key 4")
}
