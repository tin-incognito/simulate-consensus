package main

import (
	"github.com/tin-incognito/simulate-consensus/utils"
	"math/rand"
	"time"
)

//sendPrepareViewChangeMsg ...
func sendPrepareViewchangeMsg(nodes map[int]*Node){

	for _, element := range nodes{
		indexProposer := element.consensusEngine.BFTProcess.ProposalNode.index
		nodes[indexProposer].IsProposer = false
		break
	}

	rdNum := rand.Intn(int(n) - 0) + 0

	msg := ViewMsg{
		Type: PREPAREVIEWCHANGE,
		Timestamp: uint64(time.Now().Unix()),
		hash: utils.GenerateHashV1(),
		View: nodes[rdNum].consensusEngine.BFTProcess.chainHandler.View(),
		SignerID: nodes[rdNum].consensusEngine.BFTProcess.CurrNode.index,
		prevMsgHash: nil,
	}

	nodes[rdNum].consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
}
