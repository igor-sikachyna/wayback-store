package chain

type DatabaseOperation struct {
	Key    string
	Value  string
	Delete bool
}

type Transaction struct {
	Data string
}

type PushTransactionData struct {
	Data string
	DBOp *DatabaseOperation
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
	if trx.DBOp != nil {
		var operation = (*trx.DBOp)
		if operation.Delete {
			delete(c.data, operation.Key)
		} else {
			c.data[operation.Key] = operation.Value
		}
	}
	c.buildingBlock = append(c.buildingBlock, Transaction{Data: trx.Data})
}
