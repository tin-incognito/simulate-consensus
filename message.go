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
	prepareExpire bool
	commitExpire bool
}

type ViewMsg struct{
	hash        string
	Type        string // Below type
	View        uint64 // Number of view need to be switch to
	SignerID    int // Msg from who?
	Timestamp   uint64
	prevMsgHash *string
	amount      uint64
	singedMsgs  []*ViewMsg
	owner int
	isValid bool
	backViewChangeExpire bool
	backVerifyNewViewExpire bool
	backNewView bool
}

//FaultyMsg ...
type FaultyMsg struct{
	hash string
	Type string
	signerID int
	timestamp uint64
	prevMsgHash *string
	note string
	singedMsg map[int]string
}