package store

import (
	"wayback-store/chain"
)

type TableData struct {
	Hash              string
	LastBlock         int
	Hint              int
	PreviousLastBlock int
	PreviousHint      int
}

type KeyValue struct {
	Key   string
	Value string
}

type TransactionArguments struct {
	Hash  string
	Size  int
	Left  *TableData
	Right *TableData
	Data  *KeyValue
}

func (t TransactionArguments) GetHash() string {
	return "abc"
}

type Node struct {
	hash   string
	size   int
	parent *Node
	left   *Node
	right  *Node
	data   *KeyValue
}

type NodeValuePair struct {
	node  *Node
	value string
}

type Store struct {
	chain *chain.Chain
	root  *Node
	db    map[string]NodeValuePair
}

func MakeStore(chain *chain.Chain) (result Store) {
	result.chain = chain
	result.db = make(map[string]NodeValuePair)
	result.root = &Node{}

	// Rebuild the state based on the chain data

	return result
}

func (s *Store) rehashUp(node *Node) []TransactionArguments {
	return []TransactionArguments{}
}

func (s *Store) insertNode(node *Node, key string, value string) (modified *Node, transaction []TransactionArguments) {
	// Always increment the size for bookkeeping and to let the tree be balanced
	node.size++

	if node.data == nil && node.left == nil && node.right == nil {
		// Special case: root node starts without data and without left or right nodes
		node.data = &KeyValue{Key: key, Value: value}
		return node, s.rehashUp(node)
	} else if node.data != nil {
		// Node contains some data, need to split it into 2 and write the new key-value into the right node
		node.left = &Node{size: 1, parent: node, data: node.data}
		node.right = &Node{size: 1, parent: node, data: &KeyValue{Key: key, Value: value}}
		node.data = nil

		// Can either keep track of the path taken or use the parent field to traverse the tree and build the transaction
		// Here, the simplest solution is to reuse the rehashUp()
		return node.right, s.rehashUp(node)
	} else {
		// Node already contains left and right nodes, select the one that has the smallest size or the left one by default
		if node.left.size > node.right.size {
			return s.insertNode(node.right, key, value)
		}
		return s.insertNode(node.left, key, value)
	}
}

func (s *Store) insertAtRoot(key string, value string) (modified *Node, transaction []TransactionArguments) {
	return s.insertNode(s.root, key, value)
}

func (s *Store) Write(key string, value string) {
	//s.chain.PushTransaction(chain.PushTransactionData{DBOp: {Scope: }})

	var existingValue, exists = s.db[key]
	if exists {
		// Need to find the existing node and modify it
		// Can be optimized to hash the key into a number and have a secondary binary search tree or AVL tree
		// For now just store the node pointer in the key-value table and keep track of the parent node of each node

	} else {
		// Insert the value into the least-filled branch of the tree
		existingValue.node, _ = s.insertAtRoot(key, value)
	}

	existingValue.value = value
	s.db[key] = existingValue
}

func (s *Store) Get(key string) string {
	return s.db[key].value
}
