package main

import (
	"github.com/tin-incognito/simulate-consensus/utils"
	"log"
	"sync"
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
	CurrSeqNumber uint64
	CurrHeadOfChain uint64
	IsProposer bool
	Mode string
	View uint64
	PrimaryNode *Node
}

//updateAfterNormalMode ...
func (node *Node) updateAfterNormalMode() error{
	node.CurrHeadOfChain = node.consensusEngine.BFTProcess.chainHandler.Height()
	node.CurrSeqNumber = node.consensusEngine.BFTProcess.chainHandler.SeqNumber()
	return nil
}

//updateAfterViewChangeMode ...
func (node *Node) updateAfterViewChangeMode() error{
	return  nil
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
		CurrSeqNumber: 0,
		CurrHeadOfChain: 0,
		View: 0,
		Mode: NormalMode,
		IsProposer: false,
	}
	return res, nil
}

//start node in network
//
func (node *Node) start() error{
	node.consensusEngine.start()
	return nil
}

//Pool ...
//Pool will stores the nodes in network
type Pool struct{
	Nodes map[int]*Node
	n []*Node
}

func initNodes(n int) {
	nodes = make(map[int]*Node)

	for i := 0; i < n; i++{
		n := &Node{}
		var err error
		n, err = n.createNode(i)
		if err != nil {
			log.Println(err)
			continue
		}
		nodes[i] = n
	}

	for _, element := range nodes{
		element.consensusEngine.BFTProcess.CurrNode = element
		for i, e := range nodes{
			element.consensusEngine.BFTProcess.Validators[i] = e
		}
	}
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
		pool.n = append(pool.n, n)
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

func start(){
	for _, node := range nodes{
		go node.start()
	}
}

var simulateMutex sync.Mutex

func simulate(){
	log.Println("Start simulating")

	//engine := NewEngine()

	nodes[0].IsProposer = true

	var wg sync.WaitGroup

	for _, node := range nodes{

		wg.Add(1)

		go func(node *Node){
			defer wg.Done()

			node.Mode = NormalMode

			node.PrimaryNode = nodes[0]

		}(node)

		wg.Wait()

	}

	go func(){
		nodes[0].consensusEngine.BFTProcess.BroadcastMsgCh <- true

	}()

}

//start ..
func (pool *Pool) start() error {

	for _, node := range pool.Nodes{
		node.start()
	}
	return nil
}

func (pool *Pool) simulate() error{

	log.Println("Start simulating")

	pool.Nodes[0].IsProposer = true

	for index := range pool.Nodes{
		element := pool.Nodes[index]

		if element.consensusEngine.BFTProcess.ProposalNode == nil {
			element.consensusEngine.BFTProcess.ProposalNode = pool.Nodes[0]
		} else {
			*element.consensusEngine.BFTProcess.ProposalNode = *pool.Nodes[0]
		}
	}

	pool.Nodes[0].consensusEngine.BFTProcess.BroadcastMsgCh <- true

	//log.Println("pool.Nodes[0].consensusEngine.BFTProcess:", pool.Nodes[0].consensusEngine.BFTProcess)

	return nil
}