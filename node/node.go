package node

import "github.com/tin-incognito/simulate-consensus/consensus"

//Node ...
type Node struct{
	Address string `json:"address"`
	Index int `json:"index"`
	startUpTime int64
	miningKeys      string
	privateKey      string
	isEnableMining    bool
	consensusEngine *consensus.Engine
}

//NodePool ...
type Pool struct{
	Nodes map[int]*Node
}