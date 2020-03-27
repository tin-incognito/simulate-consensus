package main

import (
	"github.com/tin-incognito/simulate-consensus/utils"
	"log"
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
	validatorsAmount uint64
	view uint64
	seqNumber uint64
}

//print ...
func (chain *Chain) print(){
	log.Println("==========================================================")
	log.Println("chain height:", chain.Height)
	log.Println("latest block:", chain.LatestBlock)
	log.Println("view:", chain.view)
	log.Println("seqnumber:", chain.seqNumber)
	log.Println("==========================================================")
}

func (chain *Chain) latestBlock() *Block{
	return chain.LatestBlock
}

func (chain *Chain) IncreaseView() error{
	return nil
}

func (chain *Chain) IncreaseSeqNum() error{
	chain.seqNumber++
	return nil
}

func (chain *Chain) ValidatorsAmount() uint64{
	return chain.validatorsAmount
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
		Index: 		  chain.Height,
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
