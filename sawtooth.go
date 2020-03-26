package main

import (
	"github.com/tin-incognito/simulate-consensus/common"
	"github.com/tin-incognito/simulate-consensus/utils"
	"log"
	"math/rand"
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
	ViewChangeMsgCh chan ViewMsg
	BackViewChangeMsgCh chan ViewMsg
	BroadcastMsgCh chan bool
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
		PrePrepareMsgCh:     make (chan NormalMsg, 100),
		PrepareMsgCh:        make (chan NormalMsg, 100),
		CommitMsgCh:         make (chan NormalMsg, 100),
		FinishMsgCh:    make (chan NormalMsg, 100),
		ViewChangeMsgCh: make (chan ViewMsg, 100),
		BackViewChangeMsgCh: make (chan ViewMsg, 100),
		BroadcastMsgCh:      make (chan bool, 100),
		isStarted:           true,
		CurrNode:            nil,
		ProposalNode: nil,
		Validators: make(map[int]*Node),
		StopCh: make(chan struct{}),
		chainHandler: &Chain{},
		BFTMsgLogs: make(map[string]*NormalMsg),
		ViewChangeMsgLogs: make(map[string]*ViewMsg),
		testCh: make(chan string, 100),
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
		for {
			select {
			case msg := <- actor.testCh:
				log.Println("[Test]", msg)
			case <- actor.StopCh:
				log.Println(0)
				return
			case _ = <- actor.BroadcastMsgCh:

				// This is Pre prepare phase

				// Start idle timeout here

				//log.Println("broadcastMsg:", broadcastMsg)

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

				log.Println("[Broadcast]", msg.hash)

				msg.block = block

				// From one node to n - 1 nodes
				// So the hash of message need to be similar

				for _, member := range actor.Validators{
					//TODO: Research for more effective broadcast way
					// Or may be broadcast by go routine
					member.consensusEngine.BFTProcess.PrePrepareMsgCh <- msg
				}

			case prePrepareMsg := <- actor.PrePrepareMsgCh:

				// This is still pre prepare phase

				//actor.prepareMutex.Lock()

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				log.Println("[Pre Prepare] currActor.CurrNode.index:", currActor.CurrNode.index)
				log.Println("[Pre Prepare] prePrepareMsg.hash:", prePrepareMsg.hash)

				if !(prePrepareMsg.SignerID == currActor.ProposalNode.index){
					return
				}

				//Save it to somewhere else for every node (actor of consensus engine)
				if currActor.BFTMsgLogs[prePrepareMsg.hash] == nil{
					currActor.BFTMsgLogs[prePrepareMsg.hash] = new(NormalMsg)
					*currActor.BFTMsgLogs[prePrepareMsg.hash] = prePrepareMsg
				} else {
					currActor.BFTMsgLogs[prePrepareMsg.hash].Amount++
				}

				log.Println("[pre prepare] prePrepareMsg.hash:", prePrepareMsg.hash)
				log.Println("[pre prepare] currActor.BFTMsgLogs[prePrepareMsg.hash]:", currActor.BFTMsgLogs[prePrepareMsg.hash])

				// Reset idle timeout here

				// Move to prepare phase

				// Start commit timeout here

				log.Println("[Pre Prepare] actor.CurrNode.index:", currActor.CurrNode.index)

				// Generate 1 prepare message for each nodes and
				// Send it from 1 node to n - 1 nodes
				// Therefore each messages from each node will have different hash

				// Node (not primary node) send prepare msg to other nodes
				if len(currActor.PrePrepareMsgCh) == 0{
					currActor.prepareMutex.Lock()
					if !currActor.CurrNode.IsProposer{
						msg := NormalMsg{
							hash: 	   utils.GenerateHashV1(),
							Type:      PREPARE,
							View:      currActor.chainHandler.View(),
							SeqNum:    currActor.chainHandler.SeqNumber(),
							SignerID:  currActor.CurrNode.index,
							Timestamp: uint64(time.Now().Unix()),
							BlockID:   prePrepareMsg.BlockID,
							block: prePrepareMsg.block,
							prevMsgHash: &prePrepareMsg.hash,
						}

						//log.Println("actor.CurrNode", actor.CurrNode)

						for _, member := range currActor.Validators{
							log.Println("Sending message from pre prepare channel to prepare channel")
							log.Println("member.index:", member.index)
							log.Println("member.consensusEngine.BFTProcess:", member.consensusEngine.BFTProcess)

							//log.Println("Send message for test channel")
							//member.consensusEngine.BFTProcess.PrepareMsgCh <- msg
							//member.consensusEngine.BFTProcess.testCh <- "Test"

							go func(member *Node){
								log.Println("Send message for test channel")
								member.consensusEngine.BFTProcess.PrepareMsgCh <- msg
								member.consensusEngine.BFTProcess.testCh <- "Test"
							}(member)

							log.Println("Success in sending message to prepare channel")
						}
					}
					currActor.prepareMutex.Unlock()
				}

				//actor.prepareMutex.Unlock()

			case prepareMsg := <- actor.PrepareMsgCh:

				// This is still preparing phase

				log.Println(3)
				log.Println("prepareMsg:", prepareMsg)

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				if prepareMsg.prevMsgHash == nil {
					//TODO: Switch to view change mode
					log.Println("[prepare] prevMsgHash == null")
					return
				}

				if currActor.BFTMsgLogs[*prepareMsg.prevMsgHash] == nil {
					//TODO: Switch to view change mode
					log.Println("[prepare] Msg with this prepareMsg.prevMsgHash hash == null" )
					return
				}

				//Save it to somewhere else for every node (actor of consensus engine)
				if currActor.BFTMsgLogs[prepareMsg.hash] == nil {
					currActor.BFTMsgLogs[prepareMsg.hash] = new(NormalMsg)
					*currActor.BFTMsgLogs[prepareMsg.hash] = prepareMsg
				} else {
					currActor.BFTMsgLogs[prepareMsg.hash].Amount++
				}

				//TODO: Checking for > 2n/3

				// Node (not primary node) send prepare msg to other nodes
				//TODO: Optimize by once a node has 2n/3 switch to commit phase
				if len(currActor.PrepareMsgCh) == 0 {

					log.Println("[prepare] Jump into len prep prepare msg channel == 0")

					if currActor.BFTMsgLogs[prepareMsg.hash].Amount <= uint64(2*n/3){
						//TODO: Switch to view change mode
						log.Println("[prepare] amounts of votes <= 2/3")
						return
					}

					//Move to committing phase
					msg := NormalMsg{
						hash: 	   utils.GenerateHashV1(),
						Type:      COMMIT,
						View:      currActor.chainHandler.View(),
						SeqNum:    currActor.chainHandler.SeqNumber(),
						SignerID:  currActor.CurrNode.index,
						Timestamp: uint64(time.Now().Unix()),
						BlockID:   prepareMsg.BlockID,
						block: prepareMsg.block,
						prevMsgHash: &prepareMsg.hash,
					}

					//log.Println("actor.CurrNode", actor.CurrNode)

					for _, member := range currActor.Validators{
						log.Println("member.index:", member.index)
						log.Println("Send to prepare msg channel")
						log.Println("msg:", msg)
						log.Println(0)
						member.consensusEngine.BFTProcess.CommitMsgCh <- msg
						log.Println(2)
						log.Println("Success in sending message")
					}
				}

			case commitMsg := <- actor.CommitMsgCh:
				//This is still committing phase

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


				//TODO: Checking for > 2n/3

				// Node (not primary node) send prepare msg to other nodes
				//TODO: Optimize by once a node has greater 2n/3 switch to commit phase
				if len(actor.PrePrepareMsgCh) == 0{
					if actor.BFTMsgLogs[commitMsg.hash].Amount <= uint64(2*n/3){
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
						//Increase sequence number
						err = actor.chainHandler.IncreaseSeqNum()
						if err != nil {
							return
						}

						// FOR SIMULATE ONLY
						//Return to pre preparing phase
						actor.BroadcastMsgCh <- true
						//

						//TODO:
						// Update status of proposer
						// Calculate for random proposer (for simulating, we just increase index of node)
						// Send a msg to system for switching to other view change mode

						actor.CurrNode.IsProposer = false
						rdNum := rand.Intn(int(n) - 0) + 0

						viewChangeMsg := ViewMsg{
							Type: VIEWCHANGE,

						}

						actor.Validators[rdNum].consensusEngine.BFTProcess.ViewChangeMsgCh <- viewChangeMsg

						///
					}
				}

				//For starting a view change mode:
				// - The idle timeout expires
				// - The commit timeout expires
				// - The view change timeout expires
				// - Multiple PrePrepare messages are received for the same view and sequence number but different blocks
				// - A Prepare message is received from the primary
				// - f + 1 ViewChange messages are received for the same view

			case viewChangeMsg := <- actor.ViewChangeMsgCh:

				if viewChangeMsg.Type == PREPAREVIEWCHANGE {
					// Update its mode to viewChange mode
					err := actor.ViewChanging(actor.chainHandler.View() + 1)
					if err != nil {
						//TODO: Restart new view change mode
						return
					}

					err = actor.chainHandler.IncreaseView()
					if err != nil {
						//TODO: Restart new view change mode
						return
					}

					msg := ViewMsg{
						hash: utils.GenerateHashV1(),
						Type:       VIEWCHANGE,
						View:       actor.chainHandler.View() + 1,
						SignerID:   actor.CurrNode.index,
						Timestamp:  uint64(time.Now().Unix()),
						prevMsgHash: nil,
					}

					//Save view change msg to somewhere
					if actor.ViewChangeMsgLogs[viewChangeMsg.hash] != nil {
						actor.ViewChangeMsgLogs[viewChangeMsg.hash] = new(ViewMsg)
						*actor.ViewChangeMsgLogs[viewChangeMsg.hash] = msg
					}

					//Send messages to other nodes
					for _, element := range actor.Validators{
						element.consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
					}

				} else {
					if viewChangeMsg.Type == VIEWCHANGE{
						if viewChangeMsg.View <= actor.CurrNode.View{
							//TODO: Restart new view change mode
							return
						}

						//Save view change msg to somewhere
						if actor.ViewChangeMsgLogs[viewChangeMsg.hash] != nil {
							actor.ViewChangeMsgLogs[viewChangeMsg.hash] = new(ViewMsg)
							*actor.ViewChangeMsgLogs[viewChangeMsg.hash] = viewChangeMsg
						}

						//Define message for sending back to primary node
						msg := ViewMsg{
							Type:       BACKVIEWCHANGE,
							View:       viewChangeMsg.View,
							SignerID:   actor.CurrNode.index,
							Timestamp:  uint64(time.Now().Unix()),
							prevMsgHash: &viewChangeMsg.hash,
						}

						actor.ViewChanging(viewChangeMsg.View)

						actor.ProposalNode.consensusEngine.BFTProcess.BackViewChangeMsgCh <- msg

					} else {
						if viewChangeMsg.Type == NEWVIEW {
							//if viewChangeMsg.prevMsgHash
						} else {
							//TODO: Restart new view change mode
							return
						}
					}
				}

				log.Println(4)
				log.Println(viewChangeMsg)
			case backViewChangeMsg := <- actor.BackViewChangeMsgCh:
				if backViewChangeMsg.Type == BACKVIEWCHANGE{

					if backViewChangeMsg.prevMsgHash == nil {
						//TODO: Restart new view change mode
						return
					}

					if actor.ViewChangeMsgLogs[*backViewChangeMsg.prevMsgHash] == nil {
						//TODO: Restart new view change mode
						return
					}

					//Save view change msg to somewhere
					if actor.ViewChangeMsgLogs[backViewChangeMsg.hash] != nil {
						actor.ViewChangeMsgLogs[backViewChangeMsg.hash] = new(ViewMsg)
						*actor.ViewChangeMsgLogs[backViewChangeMsg.hash] = backViewChangeMsg
					} else {
						actor.ViewChangeMsgLogs[backViewChangeMsg.hash].amount++
					}


					if len(actor.BackViewChangeMsgCh) == 0{

						if actor.ViewChangeMsgLogs[backViewChangeMsg.hash].amount <= uint64(2*n/3){
							//TODO: Restart new view change mode
							return
						}

						//Send messages to other nodes
						for _, element := range actor.Validators{
							//Define message for sending back to primary node
							msg := ViewMsg{
								Type:       NEWVIEW,
								View:       backViewChangeMsg.View,
								SignerID:   actor.CurrNode.index,
								Timestamp:  uint64(time.Now().Unix()),
								prevMsgHash: nil,
							}
							element.consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
						}
					}
				}
			//default:
			//	log.Println("Jump to default case of selecting")
			//	time.Sleep(time.Millisecond * 1000)
			}
		}
	}()


	return nil
}

//ViewChanging ...
func (actor *Actor)ViewChanging(v uint64) error{
	actor.CurrNode.Mode = ViewChangeMode
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
	engine.BFTProcess.start()
	return nil
}

