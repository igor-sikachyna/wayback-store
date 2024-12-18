package chain

import "errors"

type DatabaseOperation struct {
	Key    string
	Value  string
	Delete bool
}

type Transaction struct {
	Data string
}

type PushTransactionData struct {
	Data  string
	DBOps *[]DatabaseOperation
}

type Block struct {
	Num int
	// Simplify the model to have just one "action" per transaction
	Transactions []Transaction
}

type Chain struct {
	headBlockNum  int
	blocks        []Block
	buildingBlock []Transaction
	data          map[string]string
}

func (c *Chain) GetBlock(blockNumber int) Block {
	return c.blocks[blockNumber-1]
}

func (c *Chain) GetBuildingBlockNum() int {
	return c.headBlockNum + 1
}

func (c *Chain) ProduceBlock() {
	c.headBlockNum++
	var block = Block{Num: c.headBlockNum, Transactions: c.buildingBlock}
	c.blocks = append(c.blocks, block)
	c.buildingBlock = []Transaction{}
}

func (c *Chain) PushTransaction(trx PushTransactionData) {
	if c.data == nil {
		c.data = make(map[string]string)
	}

	if trx.DBOps != nil {
		var operations = (*trx.DBOps)
		for _, operation := range operations {
			if operation.Delete {
				delete(c.data, operation.Key)
			} else {
				c.data[operation.Key] = operation.Value
			}
		}
	}
	c.buildingBlock = append(c.buildingBlock, Transaction{Data: trx.Data})
}

func (c *Chain) GetTableData(key string) (string, error) {
	value, ok := c.data[key]
	if ok {
		return value, nil
	}
	return "", errors.New("not found")
}
