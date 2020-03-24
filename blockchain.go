package main

import (
	"github.com/tin-incognito/simulate-consensus/utils"
	"time"
)

//Block ...
type Block struct{
	Index uint64
	Hash string
	Timestamp uint64
	Height uint64
	Producer string
	PrevHash *string
	PrevBlock *Block
	IsValid bool
}

//Chain ...
type Chain struct{
	LatestBlock *Block
	Height uint64
	ValidatorsAmount uint64
	view uint64
	seqNumber uint64
}

//SeqNumber ...
func (chain *Chain) SeqNumber() uint64{
	return chain.seqNumber
}

//View ...
func (chain *Chain) View() uint64{
	return chain.view
}

//CreateBlock ...
func (chain Chain) CreateBlock() (*Block, error){
	var prevHash *string

	if chain.LatestBlock != nil {
		prevHash = &chain.LatestBlock.Hash
	}

	hash := utils.GenerateHashV1()

	res := &Block{
		Index: 		  chain.Height + 1,
		Hash:         hash,
		PrevHash:     prevHash,
		Height:       chain.Height + 1,
		Timestamp:    uint64(time.Now().Unix()),
		PrevBlock:    chain.LatestBlock,
	}

	return res, nil
}

//ValidateBlock ...
func (chain Chain) ValidateBlock(block *Block) (bool, error){
	//if block.VoteAmount >=
	block.IsValid = true
	return true, nil
}

//InsertBlock ...
func (chain *Chain) InsertBlock(block *Block) (bool, error){
	if !block.IsValid {
		return false, nil
	}

	block.PrevBlock = chain.LatestBlock
	var prevHash *string
	if chain.LatestBlock != nil {
		prevHash = &chain.LatestBlock.Hash
	}
	block.PrevHash = prevHash

	chain.LatestBlock = block
	chain.Height++

	return true, nil
}
