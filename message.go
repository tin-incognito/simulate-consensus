package main

////MsgHandler ...
//type MsgHandler interface {
//	save() error
//	getByHash(string, map[string]msgGetter) msgGetter
//}
//
////msgGetter ...
//type msgGetter interface {
//
//}

//NormalMsg ...
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

//ViewMsg ...
type ViewMsg struct{
	hash        string
	Type        string // Below type
	View        uint64 // Number of view need to be switch to
	SignerID    int // Msg from who?
	Timestamp   uint64
	prevMsgHash *string
	amount      uint64
	hashSignedMsgs  []string
	isValid bool
}

//func (normalMsg *NormalMsg) getByHash (hash string) msgGetter{
//	return normalMsg
//}
//
////save ...
//func (normalMsg *NormalMsg) save() error{
//	return nil
//}
//
////save ...
//func (viewMsg *ViewMsg) save() error{
//	return nil
//}

