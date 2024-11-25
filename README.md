# wayback-store

A simple POC to show how to store key-value data in blockchain transaction data (calldata on Ethereum) while having very little data stored in key-value store (storage on Ethereum)

2 types of transactions are introduced:
- Modify the on-chain table to store the reference to the root of the Merkle tree (current block number)
- Do nothing, but add the references to the left and right nodes of the Merkle tree to transaction arguments

When writing into the table:
- Add the new value locally to the least filled branch of the tree (or modify an existing leaf if the key already exists)
- Recalculate the hashes of the modified nodes locally
- Build and push a transaction that includes the data for all modified nodes + a transaction to modify the on-chain table state

To read the key-value table:
- Read the on-chain data to get the root of the tree
- Traverse the linked blocks to rebuild the key-value table

To get an older iteration of the key-value table state:
- Read the on-chain data to get the reference to a previous state of the tree root
- Traverse the references until you find the tree root state you are interested in
- Traverse the linked blocks to rebuild the key-value table

On EOS-based chains this approach allows you to have a persistent data storage with write bandwidth limited by your stake and no read bandwidth limitations.

On Ethereum you have to pay gas for each modification, but it still may theoretically be cost-efficient:

- Storage costs 20000 + 2176 Gas per 32 bytes = 693 gas per byte
- Calldata costs 16 gas per byte
- Assume the minimum overhead of (int32 (block number) + int16 (hint)) * 2 + int256 (hash) + ~100 (overhead of calling the contract) = 144 bytes per transaction
- Assume data size of 256 bytes per key (value + key size combined since the key is used for convenience and is not used to balance the tree)
- Maximum depth of the Merkle tree before storage becomes more cost-efficient: 693 * 256 == (144 * x + 256) * 16 -> x = 75 (maximum 2^74 key-value pairs)
- For 1024 values (depth 10, x = 11) you will reduce the gas spending by a factor of 6