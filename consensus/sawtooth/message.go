package sawtooth

import "encoding/json"


type PbftMsgInfo struct{
	Type string
	View uint64
	SeqNum uint64
	SignerID []byte
}

//MessageBFT ...
type MessageBFT struct{
	SignerID string
	NodeID    string
	Type      string
	Content   []byte
	Timestamp int64
	SeqNum uint64
	View uint64

	//TimeSlot  int64
}

//type

//BFTPropose ...
type BFTPropose struct{
	Block json.RawMessage
}

//BFTVote ...
type BFTVote struct {
	Block json.RawMessage
}


