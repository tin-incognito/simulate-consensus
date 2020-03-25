package main

import (
	"github.com/tin-incognito/simulate-consensus/common"
	"github.com/tin-incognito/simulate-consensus/utils"
	"log"
	"sync"
	"time"
)

//Actor
type Actor struct{
	PrePrepareMsgCh chan NormalMsg
	preMutex sync.Mutex
	PrepareMsgCh    chan NormalMsg
	prepareMutex sync.Mutex
	CommitMsgCh chan NormalMsg
	commitMutex sync.Mutex
	FinishMsgCh chan NormalMsg
	finishMutex sync.Mutex
	ViewChangeMessageCh chan ViewMsg
	BroadcastMsgCh chan struct{}
	isStarted      bool
	StopCh         chan struct{}
	CurrNode 	   *Node
	ProposalNode *Node
	Validators map[int]*Node
	chainHandler ChainHandler
	BFTMsgLogs map[string]*NormalMsg
	ViewChangeMsgLogs map[string]*ViewMsg
	//sendMutex sync.Mutex
	testCh chan string
}

func NewActor() *Actor{

	res := &Actor{
		PrePrepareMsgCh:     make (chan NormalMsg),
		PrepareMsgCh:        make (chan NormalMsg),
		CommitMsgCh:         make (chan NormalMsg),
		FinishMsgCh:    make (chan NormalMsg),
		ViewChangeMessageCh: make (chan ViewMsg),
		BroadcastMsgCh:      make (chan struct{}),
		isStarted:           true,
		CurrNode:            nil,
		ProposalNode: nil,
		Validators: make(map[int]*Node),
		StopCh: make(chan struct{}),
		chainHandler: &Chain{},
		BFTMsgLogs: make(map[string]*NormalMsg),
		ViewChangeMsgLogs: make(map[string]*ViewMsg),
		testCh: make(chan string),
	}

	return res
}

//Name return name of actor to user
func (actor Actor) Name() (string,error){
	return common.SawToothConsensus, nil
}

//Start ...
func (actor Actor) start() error{

	actor.isStarted = true

	go func(){
		//ticker := time.Tick(300 * time.Millisecond)

		select {
		case msg := <- actor.testCh:
			log.Println(msg)
		case <- actor.StopCh:
			log.Println(0)
			return
		case broadcastMsg := <- actor.BroadcastMsgCh:

			// This is Pre prepare phase

			// Start idle timeout here

			log.Println("broadcastMsg:", broadcastMsg)
			block, err := actor.chainHandler.CreateBlock()
			if err != nil {
				log.Println(err)
				return
			}

			msg := NormalMsg{
				hash: utils.GenerateHashV1(),
				Type:      PREPREPARE,
				View:      actor.chainHandler.View(),
				SeqNum:    actor.chainHandler.SeqNumber(),
				SignerID:  actor.CurrNode.index,
				Timestamp: uint64(time.Now().Unix()),
				BlockID:   &block.Index,
				prevMsgHash: nil,
			}

			msg.block = block

			for _, member := range actor.Validators{
				//TODO: Research for more effective broadcast way
				// Or may be broadcast by go routine
				log.Println(-1)
				member.consensusEngine.BFTProcess.PrePrepareMsgCh <- msg
				log.Println(1)
			}

		case prePrepareMsg := <- actor.PrePrepareMsgCh:

			// This is still pre prepare phase

			//actor.prepareMutex.Lock()

			log.Println(prePrepareMsg)

			// Check if prepare msg if valid
			if !(prePrepareMsg.SignerID == actor.ProposalNode.index){
				return
			}

			//Save it to somewhere else for every node (actor of consensus engine)
			if actor.BFTMsgLogs[prePrepareMsg.hash] == nil{
				actor.BFTMsgLogs[prePrepareMsg.hash] = new(NormalMsg)
				*actor.BFTMsgLogs[prePrepareMsg.hash] = prePrepareMsg
			} else {
				actor.BFTMsgLogs[prePrepareMsg.hash].Amount++
			}

			// Reset idle timeout here

			// Move to prepare phase

			// Start commit timeout here

			log.Println("actor.CurrNode.index:", actor.CurrNode.index)

			// Node (not primary node) send prepare msg to other nodes
			if len(actor.PrePrepareMsgCh) == 0{
				if !actor.CurrNode.IsProposer{
					msg := NormalMsg{
						hash: 	   utils.GenerateHashV1(),
						Type:      PREPARE,
						View:      actor.chainHandler.View(),
						SeqNum:    actor.chainHandler.SeqNumber(),
						SignerID:  actor.CurrNode.index,
						Timestamp: uint64(time.Now().Unix()),
						BlockID:   prePrepareMsg.BlockID,
						block: prePrepareMsg.block,
						prevMsgHash: &prePrepareMsg.hash,
					}

					//log.Println("actor.CurrNode", actor.CurrNode)

					for _, member := range actor.Validators{
						log.Println(member.IsProposer)
						log.Println("member.index:", member.index)
						log.Println("Send to prepare msg channel")
						log.Println("msg:", msg)
						log.Println(0)
						member.consensusEngine.BFTProcess.PrepareMsgCh <- msg
						log.Println(2)
					}
				}
			}

			//actor.prepareMutex.Unlock()

		case prepareMsg := <- actor.PrepareMsgCh:

			// This is still preparing phase

			log.Println(3)
			log.Println(prepareMsg)


			if prepareMsg.prevMsgHash == nil {
				//TODO: Switch to view change mode
				return
			}

			if actor.BFTMsgLogs[*prepareMsg.prevMsgHash] == nil{
				//TODO: Switch to view change mode
				return
			}

			//Save it to somewhere else for every node (actor of consensus engine)
			if actor.BFTMsgLogs[prepareMsg.hash] == nil{
				actor.BFTMsgLogs[prepareMsg.hash] = new(NormalMsg)
				*actor.BFTMsgLogs[prepareMsg.hash] = prepareMsg
			} else {
				actor.BFTMsgLogs[prepareMsg.hash].Amount++
			}


			//TODO: Checking for > 2*f + 1

			// Node (not primary node) send prepare msg to other nodes
			//TODO: Optimize by once a node has 2f + 1 switch to commit phase
			if len(actor.PrePrepareMsgCh) == 0{
				if actor.BFTMsgLogs[prepareMsg.hash].Amount <= (2 * f + 1){
					//TODO: Switch to view change mode
					return
				}

				//Move to committing phase
				msg := NormalMsg{
					hash: 	   utils.GenerateHashV1(),
					Type:      COMMIT,
					View:      actor.chainHandler.View(),
					SeqNum:    actor.chainHandler.SeqNumber(),
					SignerID:  actor.CurrNode.index,
					Timestamp: uint64(time.Now().Unix()),
					BlockID:   prepareMsg.BlockID,
					block: prepareMsg.block,
					prevMsgHash: &prepareMsg.hash,
				}

				//log.Println("actor.CurrNode", actor.CurrNode)

				for _, member := range actor.Validators{
					log.Println("member.index:", member.index)
					log.Println("Send to prepare msg channel")
					log.Println("msg:", msg)
					log.Println(0)
					member.consensusEngine.BFTProcess.CommitMsgCh <- msg
					log.Println(2)
				}
			}

		case commitMsg := <- actor.CommitMsgCh:
			//This is still commiting phase

			// This is still preparing phase

			log.Println(3)
			log.Println(commitMsg)


			if commitMsg.prevMsgHash == nil {
				//TODO: Switch to view change mode
				return
			}

			if actor.BFTMsgLogs[*commitMsg.prevMsgHash] == nil{
				//TODO: Switch to view change mode
				return
			}

			//Save it to somewhere else for every node (actor of consensus engine)
			if actor.BFTMsgLogs[commitMsg.hash] == nil{
				actor.BFTMsgLogs[commitMsg.hash] = new(NormalMsg)
				*actor.BFTMsgLogs[commitMsg.hash] = commitMsg
			} else {
				actor.BFTMsgLogs[commitMsg.hash].Amount++
			}


			//TODO: Checking for > 2*f + 1

			// Node (not primary node) send prepare msg to other nodes
			//TODO: Optimize by once a node has 2f + 1 switch to commit phase
			if len(actor.PrePrepareMsgCh) == 0{
				if actor.BFTMsgLogs[commitMsg.hash].Amount <= (2 * f + 1){
					//TODO: Switch to view change mode
					return
				}

				//Move to finishing phase

				//TODO: Stop commit timeout here

				if actor.CurrNode.IsProposer{
					//Update current chain
					check, err := actor.chainHandler.ValidateBlock(commitMsg.block)
					if err != nil || !check {
						//TODO: Switch to view change mode
						return
					}
					check, err = actor.chainHandler.InsertBlock(commitMsg.block)
					if err != nil || !check {
						//TODO: Switch to view change mode
						return
					}
					//Update status of proposer
					//actor.CurrNode.IsProposer = false
					err = actor.chainHandler.IncreaseSeqNum()
					if err != nil {
						return
					}
					//Return to prepreparing phase
					//
				}
			}

			//For starting a view change mode:
			// - The idle timeout expires
			// - The commit timeout expires
			// - The view change timeout expires
			// - Multiple PrePrepare messages are received for the same view and sequence number but different blocks
			// - A Prepare message is received from the primary
			// - f + 1 ViewChange messages are received for the same view
		case viewChangeMsg := <- actor.ViewChangeMessageCh:

			//There are 3 types of viewChange msg:
			// - Viewchange msg
			// - NewView msg
			// -

			//Process of view changing mode
			// - One node will update its mode to viewChanging(v), where v is the node is changing to
			// - Stop both the idle and commit time out
			// - Stop the view change timeout if it's been started
			// - Broadcast a ViewChange message for new view
			// -

			log.Println(4)
			log.Println(viewChangeMsg)
		//default:
		//	log.Println("Error in sending message")

		//case <-ticker:
		//	log.Println(5)
		//	return
		}
	}()


	return nil
}

func (actor *Actor) initValidators(m map[int]*Node) {
	for i, element := range m{
		if actor.CurrNode.index != i{
			actor.Validators[i] = element
		}
	}
}

// Engine ...
type Engine struct{
	BFTProcess *Actor
}

func NewEngine() *Engine {
	engine := &Engine{
		BFTProcess: NewActor(),
	}

	return engine
}

//start ...
func (engine *Engine) start() error{
	err := engine.BFTProcess.start()
	return err
}

