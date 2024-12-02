package store

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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
	Hint  int
	Left  *BlockHint
	Right *BlockHint
	Data  *KeyValue
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

func (s *Store) restoreTree(node *Node, block int, hint int) (err error) {
	// Can be optimized by caching blocks instead of fetching them each time
	var blockData = s.chain.GetBlock(block)
	// Find the matching hint
	var found bool = false
	for _, trx := range blockData.Transactions {
		var transaction TransactionArguments
		json.Unmarshal([]byte(trx.Data), &transaction)
		if transaction.Hint == hint {
			found = true
			node.block = block
			node.hint = hint

			if transaction.Left != nil {
				// Right should always exist if there is Left, for simplicity won't check it
				node.left = &Node{parent: node}
				node.right = &Node{parent: node}
				err = s.restoreTree(node.left, transaction.Left.Block, transaction.Left.Hint)
				if err != nil {
					return err
				}
				err = s.restoreTree(node.right, transaction.Right.Block, transaction.Right.Hint)
				if err != nil {
					return err
				}
			} else {
				// In all other cases there should be Data
				// Make a data copy instead of reference
				node.data = &KeyValue{Key: transaction.Data.Key, Value: transaction.Data.Value}
				s.db[transaction.Data.Key] = NodeValuePair{node: node, value: transaction.Data.Value}
				// Will calculate redundant hashes, can be optimized by post-order traversing the tree
				s.rehashUp(node)
			}
		}
	}

	// TODO: Recalculate the sizes

	if !found {
		return errors.New("not found")
	}
	return
}

func MakeStore(chain *chain.Chain) (result Store, err error) {
	result.chain = chain
	result.db = make(map[string]NodeValuePair)
	result.root = &Node{}

	// Rebuild the state based on the chain data
	blockString, tempErr := chain.GetTableData("block")
	if tempErr != nil {
		return
	}
	block, err := strconv.Atoi(blockString)
	if err != nil {
		return
	}
	hintString, tempErr := chain.GetTableData("hint")
	if tempErr != nil {
		return
	}
	hint, err := strconv.Atoi(hintString)
	if err != nil {
		return
	}

	result.nextHint = hint + 1
	err = result.restoreTree(result.root, block, hint)

	return
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

func prependTransaction(trx []TransactionArguments, action TransactionArguments) []TransactionArguments {
	trx = append(trx, TransactionArguments{})
	copy(trx[1:], trx)
	trx[0] = action
	return trx
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
		*transaction = prependTransaction(*transaction, TransactionArguments{
			Hint: node.hint,
			Data: node.data,
		})
	} else {
		*transaction = prependTransaction(*transaction, TransactionArguments{
			Hint: node.hint,
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
		node.left = &Node{size: 1, parent: node, data: node.data, block: node.block, hint: node.hint}
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

func (s *Store) formatHeadTransaction(transaction TransactionArguments) (result chain.PushTransactionData) {
	data, err := json.Marshal(transaction)
	if err != nil {
		fmt.Println(err)
		return
	}
	return chain.PushTransactionData{
		Data: string(data),
		DBOps: &[]chain.DatabaseOperation{{
			Key:   "hint",
			Value: strconv.Itoa(transaction.Hint),
		}, {
			Key:   "block",
			Value: strconv.Itoa(s.chain.GetBuildingBlockNum()),
		}},
	}
}

func (s *Store) formatBodyTransaction(transaction TransactionArguments) (result chain.PushTransactionData) {
	data, err := json.Marshal(transaction)
	if err != nil {
		fmt.Println(err)
		return
	}
	return chain.PushTransactionData{
		Data:  string(data),
		DBOps: nil,
	}
}

func (s *Store) Write(key string, value string) {
	//s.chain.PushTransaction(chain.PushTransactionData{DBOp: {Scope: }})

	var existingValue, exists = s.db[key]
	var transaction []TransactionArguments
	if exists {
		// Need to find the existing node and modify it
		// Can be optimized to hash the key into a number and have a secondary binary search tree or AVL tree
		// For now just store the node pointer in the key-value table and keep track of the parent node of each node
		existingValue.value = value
		transaction = s.rehashUp(existingValue.node)
	} else {
		// Insert the value into the least-filled branch of the tree
		existingValue.node, transaction = s.insertAtRoot(key, value)
	}
	fmt.Println(transaction)

	existingValue.value = value
	s.db[key] = existingValue

	// Push transaction
	for i, trx := range transaction {
		if i == 0 {
			s.chain.PushTransaction(s.formatHeadTransaction(trx))
		} else {
			s.chain.PushTransaction(s.formatBodyTransaction(trx))
		}
	}
}

func (s *Store) Get(key string) string {
	return s.db[key].value
}
