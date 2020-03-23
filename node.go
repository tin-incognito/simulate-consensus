package main

import "github.com/tin-incognito/simulate-consensus/consensus"

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
}

//createNode for initiating node in network
//
func (node *Node) createNode() (*Node, error){
	return nil, nil
}

//start node in network
//
func (node *Node) start() error{

}