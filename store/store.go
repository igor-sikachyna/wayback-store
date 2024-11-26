package store

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"wayback-store/chain"
)

type BlockHint struct {
	Block int
	Hint  int
}

type TableData struct {
	Hash     string
	Current  BlockHint
	Previous BlockHint
}

type KeyValue struct {
	Key   string
	Value string
}

type TransactionArguments struct {
	Left  *BlockHint
	Right *BlockHint
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
	block  int
	hint   int
}

type NodeValuePair struct {
	node  *Node
	value string
}

type Store struct {
	chain    *chain.Chain
	root     *Node
	db       map[string]NodeValuePair
	nextHint int
}

func MakeStore(chain *chain.Chain) (result Store) {
	result.chain = chain
	result.db = make(map[string]NodeValuePair)
	result.root = &Node{}

	// Rebuild the state based on the chain data

	return result
}

func hash(data string) (result string) {
	var hasher = sha1.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (s *Store) getNextHint() int {
	s.nextHint++
	return s.nextHint
}

func (s *Store) rehashUpNext(node *Node, transaction *[]TransactionArguments) {
	if node == nil {
		return
	}

	if node.data != nil {
		// The "correct" way it to hash the sum of hashes, for simplicity we use a delimiter not allowed in the key
		node.hash = hash(node.data.Key + "|" + node.data.Value)
	} else {
		node.hash = hash(node.left.hash + node.right.hash)
	}

	// In practice you won't know the block to assign at this stage
	// You will need to push the transaction and wait for the transaction to appear in a block
	// The block number where the transaction appeared should be added to the store and no modification is possible until this moment
	// For simplicity we will pretend like we already know the block number
	node.block = s.chain.GetBuildingBlockNum()
	node.hint = s.getNextHint()

	if node.data != nil {
		*transaction = append(*transaction, TransactionArguments{
			Data: node.data,
		})
	} else {
		*transaction = append(*transaction, TransactionArguments{
			Left: &BlockHint{
				Block: node.left.block,
				Hint:  node.left.hint,
			},
			Right: &BlockHint{
				Block: node.right.block,
				Hint:  node.right.hint,
			},
		})
	}

	s.rehashUpNext(node.parent, transaction)
}

func (s *Store) rehashUp(node *Node) (result []TransactionArguments) {
	s.rehashUpNext(node, &result)
	return
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
		return node.right, s.rehashUp(node.right)
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
		var transaction []TransactionArguments
		existingValue.node, transaction = s.insertAtRoot(key, value)
		fmt.Println(transaction)
	}

	existingValue.value = value
	s.db[key] = existingValue
}

func (s *Store) Get(key string) string {
	return s.db[key].value
}
