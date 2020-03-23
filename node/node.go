package node

import (
	"github.com/tin-incognito/simulate-consensus/consensus"
	"github.com/tin-incognito/simulate-consensus/utils"
	"time"
)


//Node ...
//Each node will using a consensus engine
//For joining in network
type Node struct{
	address string `json:"address"`
	index int `json:"index"`
	timestamp int64
	miningKeys      string
	privateKey      string
	isEnableMining    bool
	consensusEngine *consensus.Engine
	ViewNumber uint64
	CurrSeqNumber uint64
	CurrHeadOfChain uint64
	PhaseStatus int
	//BlocksLog
	//MessagesLog
}

//createNode for initiating node in network
//
func (node *Node) CreateNode(index int) (*Node, error){

	res := &Node{
		address:         utils.GenerateHashV1(),
		index:           index,
		timestamp:       time.Now().Unix(),
		miningKeys:      utils.GenerateKey(),
		privateKey:      utils.GenerateKey(),
		isEnableMining:  true,
		consensusEngine: nil,
	}
	return res, nil
}

//start node in network
//
func (node *Node) Start() error{
	go node.consensusEngine.Start()
	return nil
}