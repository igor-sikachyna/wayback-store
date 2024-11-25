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
	hash  string
	size  int
	left  *Node
	right *Node
	data  *KeyValue
}

type Store struct {
	chain *chain.Chain
	root  *Node
	db    map[string]string
}

func MakeStore(chain *chain.Chain) (result Store) {
	result.chain = chain
	result.db = make(map[string]string)

	// Rebuild the state based on the chain data

	return result
}

func (s *Store) Write(key string, value string) {
	//s.chain.PushTransaction(chain.PushTransactionData{DBOp: {Scope: }})

	s.db[key] = value
}

func (s *Store) Get(key string) string {
	return s.db[key]
}
