package chain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateChain(t *testing.T) {
	var c = Chain{}
	var PanicGetBlock assert.PanicTestFunc = func() {
		c.GetBlock(1)
	}

	var assert = assert.New(t)

	assert.Equal(1, c.GetBuildingBlockNum(), "incorrect head block number")

	assert.Panics(PanicGetBlock, "did not fail to get a block")

	c.ProduceBlock()

	assert.Equal(2, c.GetBuildingBlockNum(), "incorrect head block number")
	var block = c.GetBlock(1)
	assert.Equal(1, block.Num, "incorrect retrieved block number")
	assert.Equal(0, len(block.Transactions), "incorrect transactions count")

	var _, err = c.GetTableData("test")
	assert.NotEqual(nil, err, "did not return an error for a non-existing value")

	c.PushTransaction(PushTransactionData{
		Data: "test",
		DBOps: &[]DatabaseOperation{
			{
				Key:    "hello",
				Value:  "world",
				Delete: false,
			},
		},
	})

	c.ProduceBlock()
	block = c.GetBlock(2)

	assert.Equal(1, len(block.Transactions), "incorrect transactions count after pushing transaction")
	assert.Equal("test", block.Transactions[0].Data, "incorrect transaction data")
	val, err := c.GetTableData("hello")
	assert.Equal(nil, err, "returned an error for existing table value")
	assert.Equal("world", val, "incorrect information stored in the table")

	c.PushTransaction(PushTransactionData{
		Data: "test",
		DBOps: &[]DatabaseOperation{
			{
				Key:    "hello",
				Delete: true,
			},
		},
	})

	c.ProduceBlock()

	_, err = c.GetTableData("hello")
	assert.NotEqual(nil, err, "did not return an error for a value that no longer exists")
}
