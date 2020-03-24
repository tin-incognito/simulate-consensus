package main

//PbftMsgInfo ...
type PbftMsgInfo struct{
	Type string // Below type
	View uint64 // Number of view in network
	SeqNum uint64 // sequence number = block.Height + 1
	SignerID int // Msg from who?
}

//PbftMsg ...
type PbftMsg struct{
	PbftMsgInfo
	BlockID *uint64 // Msg for which block?
}

const (
	VIEWCHANGE = "viewchange"
	PREPREPARE = "preprepare"
	PREPARE = "prepare"
	COMMIT = "commit"
	SEALREQUEST = "sealrequest"
)