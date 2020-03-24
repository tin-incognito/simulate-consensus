package main

import (
	"github.com/tin-incognito/simulate-consensus/utils"
	"log"
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
	consensusEngine *Engine
	ViewNumber uint64
	CurrSeqNumber uint64
	CurrHeadOfChain uint64
	PhaseStatus int
	//BlocksLog
	//MessagesLog
}

//createNode for initiating node in network
//
func (node *Node) createNode(index int) (*Node, error){

	res := &Node{
		address:         utils.GenerateHashV1(),
		index:           index,
		timestamp:       time.Now().Unix(),
		miningKeys:      utils.GenerateKey(),
		privateKey:      utils.GenerateKey(),
		isEnableMining:  true,
		consensusEngine: NewEngine(),
	}
	return res, nil
}

//start node in network
//
func (node *Node) start() error{
	go node.consensusEngine.start()
	return nil
}

//Pool ...
//Pool will stores the nodes in network
type Pool struct{
	Nodes map[int]*Node
}

//createPool for storing nodes
func (pool *Pool) createPool(amountOfNode int) (*Pool, error){
	res := &Pool{}
	res.Nodes = make(map[int]*Node)

	for i := 0; i < amountOfNode; i++{
		n := &Node{}
		var err error
		n, err = n.createNode(i)
		if err != nil {
			log.Println(err)
			continue
		}
		res.Nodes[i] = n
	}

	return res, nil
}

func (pool *Pool) initValidators() error{
	for _, element := range pool.Nodes{
		element.consensusEngine.BFTProcess.CurrNode = element
		element.consensusEngine.BFTProcess.initValidators(pool.Nodes)
	}
	return nil
}

//start ..
func (pool *Pool) start() error {

	for _, node := range pool.Nodes{
		node.start()
	}
	return nil
}

func (pool *Pool) simulate(chain *Chain) error{
	
	for i, node := range pool.Nodes{
		block, err := chain.CreateBlock()
		if err != nil {
			log.Println(err)
			return err
		}
		msg := PbftMsg{
			PbftMsgInfo: PbftMsgInfo{
				Type:     SEALREQUEST,	
				View:     0,
				SeqNum:   0,
				SignerID: i,
			},
			BlockID: &block.Index,
		}

		log.Println("Node", i)
		log.Println("msg:", msg)

		log.Println(node.consensusEngine)
		log.Println(node.consensusEngine.BFTProcess)
		log.Println(node.consensusEngine.BFTProcess.BroadcastMsgCh)

		node.consensusEngine.BFTProcess.BroadcastMsgCh <- msg
	}
	return nil
}