package blockchain

//Block ...
type Block struct{
	Hash string
	PrevHash *string
	Height uint64
	Version int8
	CPendingTxs <-chan Transaction
	CRemovedTxs <-chan Transaction
	PedingTxs map[string]Transaction
	FinalizedTxs map[string]Transaction
	Timestamp uint64
	PrevBlock *Block
	IsValid bool
	VoteAmount uint64
}

