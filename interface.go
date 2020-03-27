package main

//BlockHandler ...
type BlockHandler interface {

}

//ChainHandler ...
type ChainHandler interface {
	CreateBlock() (*Block, error)
	ValidateBlock(block *Block) (bool, error)
	InsertBlock(block *Block) (bool, error)
	SeqNumber() uint64
	View() uint64
	ValidatorsAmount() uint64
	IncreaseSeqNum() error
	IncreaseView() error
	print()
	latestBlock() *Block
}

