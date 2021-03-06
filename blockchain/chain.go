package blockchain

import (
	"github.com/tin-incognito/simulate-consensus/utils"
	"time"
)

//Chain
type Chain struct {
	LatestBlock *Block
	Height uint64
	ValidatorsAmount uint64
}

//CreateBlock ...
func (chain Chain) CreateBlock(finalizedTxs, pendingTxs map[string]Transaction) (*Block, error){
	var prevHash *string

	if chain.LatestBlock != nil {
		prevHash = &chain.LatestBlock.Hash
	}

	hash := utils.GenerateHashV1()

	res := &Block{
		Hash:         hash,
		PrevHash:     prevHash,
		Height:       chain.Height + 1,
		Version:      1,
		CPendingTxs:  nil,
		CRemovedTxs:  nil,
		PedingTxs:    pendingTxs,
		FinalizedTxs: finalizedTxs,
		Timestamp:    uint64(time.Now().Unix()),
		PrevBlock:    chain.LatestBlock,
		IsValid: 	  false,
		VoteAmount:   0,
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