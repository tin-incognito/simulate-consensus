package main

type NormalMsg struct{
	hash string
	Type string // Below type
	View uint64 // Number of view in network
	SeqNum uint64 // Sequence number = block.Height + 1
	SignerID int // Msg from who?
	Timestamp uint64
	BlockID *uint64
	block *Block
	Amount uint64
	prevMsgHash *string
}

type ViewMsg struct{

}