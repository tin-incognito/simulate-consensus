package sawtooth

import "encoding/json"

//MessageBFT ...
type MessageBFT struct{
	SignerID string
	NodeID    string
	Type      string
	Content   []byte
	Timestamp int64
	//TimeSlot  int64
}

//BFTPropose ...
type BFTPropose struct{
	Block json.RawMessage
}

//BFTVote ...
type BFTVote struct {
	Block json.RawMessage
}