package main

//BlockHandler ...
type BlockHandler interface {

}

//ChainHandler ...
type ChainHandler interface {
	CreateBlock() (*Block, error)
	ValidateBlock(block *Block) (bool, error)
	InsertBlock(block *Block) error
}

