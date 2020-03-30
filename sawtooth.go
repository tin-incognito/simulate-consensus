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
	viewChangeMutex sync.Mutex
	BackViewChangeMsgCh chan ViewMsg
	backViewChangeMutex sync.Mutex
	BroadcastMsgCh chan bool
	isStarted      bool
	StopCh         chan struct{}
	CurrNode 	   *Node
	ProposalNode *Node
	Validators map[int]*Node
	chainHandler ChainHandler
	BFTMsgLogs map[string]*NormalMsg
	ViewChangeMsgLogs map[string]*ViewMsg
	logBlockMutex sync.Mutex
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

	//count := 0

	go func(){
		//ticker := time.Tick(300 * time.Millisecond)
		for {
			select {
			case <- actor.StopCh:
				log.Println(0)
				return
			case _ = <- actor.BroadcastMsgCh:

				// This is Pre prepare phase

				//TODO:
				// Start idle timeout here

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				block, err := currActor.chainHandler.CreateBlock()
				if err != nil {
					log.Println(err)
					return
				}

				msg := NormalMsg{
					hash: utils.GenerateHashV1(),
					Type:      PREPREPARE,
					View:      currActor.chainHandler.View(),
					SeqNum:    currActor.chainHandler.SeqNumber(),
					SignerID:  currActor.CurrNode.index,
					Timestamp: uint64(time.Now().Unix()),
					BlockID:   &block.Index,
					prevMsgHash: nil,
				}

				//log.Println("[Broadcast]", msg.hash)

				msg.block = block

				if currActor.BFTMsgLogs[msg.hash] == nil{
					currActor.BFTMsgLogs[msg.hash] = new(NormalMsg)
					*currActor.BFTMsgLogs[msg.hash] = msg
				} else {
					currActor.BFTMsgLogs[msg.hash].Amount++
				}

				// From one node to n - 1 nodes
				// So the hash of message need to be similar

				//log.Println(len(currActor.Validators))

				for _, member := range currActor.Validators{
					//TODO: Research for more effective broadcast way
					// Or may be broadcast by go routine
					//log.Println("Send pre prepare msg")

					if member.index == currActor.CurrNode.index{
						continue
					}

					member.consensusEngine.BFTProcess.PrePrepareMsgCh <- msg
				}

			case prePrepareMsg := <- actor.PrePrepareMsgCh:

				//log.Println("Receive pre prepare msg")

				// This is still pre prepare phase

				currActor := actor.CurrNode.consensusEngine.BFTProcess

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

				//TODO:
				// Reset idle timeout here

				// Move to prepare phase

				//TODO:
				// Start commit timeout here

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


						//log.Println("curActor.Validators:", currActor.Validators)

						for _, member := range currActor.Validators{
							go func(member *Node){
								//log.Println("Send prepare msg")
								member.consensusEngine.BFTProcess.PrepareMsgCh <- msg
							}(member)
						}
					}
					currActor.prepareMutex.Unlock()
				}

				//actor.prepareMutex.Unlock()

			case prepareMsg := <- actor.PrepareMsgCh:

				// This is still preparing phase

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				if prepareMsg.prevMsgHash == nil {
					//TODO: Switch to view change mode
					log.Println("[prepare] prevMsgHash == null")

					return
				}

				//log.Println("[prepare] *prepareMsg.prevMsgHash:", *prepareMsg.prevMsgHash)

				currActor.prepareMutex.Lock()
				if currActor.BFTMsgLogs[*prepareMsg.prevMsgHash] == nil {
					//TODO: Switch to view change mode
					log.Println("[prepare] Msg with this prepareMsg.prevMsgHash hash == null" )
					//currActor.startNormalMode()
					return
				}
				currActor.prepareMutex.Unlock()

				//Save it to somewhere else for every node (actor of consensus engine)
				if currActor.BFTMsgLogs[prepareMsg.hash] == nil {
					currActor.BFTMsgLogs[prepareMsg.hash] = new(NormalMsg)
					*currActor.BFTMsgLogs[prepareMsg.hash] = prepareMsg

				} else {
					currActor.BFTMsgLogs[prepareMsg.hash].Amount++
				}

				//TODO: Checking for > 2n/3

				// Set timeout here

				//currActor.prepareMutex.Lock()
				//
				//amount := 0
				//
				//for _, msg := range currActor.BFTMsgLogs{
				//	if msg.prevMsgHash != nil && *msg.prevMsgHash == *prepareMsg.prevMsgHash && msg.Type == PREPARE{
				//		amount++
				//	}
				//}

				// Wait for end of time
				// If timeout check for 2n > 3

				// Node (not primary node) send prepare msg to other nodes
				//TODO: Optimize by once a node has 2n/3 switch to commit phase
				if len(currActor.PrepareMsgCh) == 0 {

					//log.Println("[prepare] Jump into len prep prepare msg channel == 0")

					currActor.prepareMutex.Lock()

					amount := 0

					for _, msg := range currActor.BFTMsgLogs{
						if msg.prevMsgHash != nil && *msg.prevMsgHash == *prepareMsg.prevMsgHash && msg.Type == PREPARE{
							amount++
						}
					}

					if uint64(amount) <= uint64(2*n/3){
						//TODO: Switch to view change mode
						//log.Println("[prepare] amounts of votes <= 2/3")
						//log.Println("amount:", amount)
						//return
					} else {
						if !currActor.BFTMsgLogs[*prepareMsg.prevMsgHash].prepareExpire{
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
								prevMsgHash: prepareMsg.prevMsgHash,
							}

							for _, member := range currActor.Validators{
								go func(member *Node){
									member.consensusEngine.BFTProcess.CommitMsgCh <- msg
								}(member)
							}

							amount = 0
							currActor.BFTMsgLogs[*prepareMsg.prevMsgHash].prepareExpire = true
						}
					}

					currActor.prepareMutex.Unlock()
				}

			case commitMsg := <- actor.CommitMsgCh:
				//This is still committing phase

				//log.Println("commitMsg:", commitMsg)

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				if commitMsg.prevMsgHash == nil {
					//TODO: Switch to view change mode
					log.Println("[commit] commit.prevMsgHash == nil")
					return
				}

				if currActor.BFTMsgLogs[*commitMsg.prevMsgHash] == nil{
					//TODO: Switch to view change mode
					return
				}

				//Save it to somewhere else for every node (actor of consensus engine)
				if currActor.BFTMsgLogs[commitMsg.hash] == nil{
					currActor.BFTMsgLogs[commitMsg.hash] = new(NormalMsg)
					*currActor.BFTMsgLogs[commitMsg.hash] = commitMsg
				} else {
					currActor.BFTMsgLogs[commitMsg.hash].Amount++
				}


				//TODO: Checking for > 2n/3

				// Node (not primary node) send prepare msg to other nodes

				//currActor.commitMutex.Lock()
				//time.Sleep(time.Millisecond * 500)
				//time.Sleep(time.Millisecond * 500)
				//currActor.commitMutex.Unlock()

				//TODO: Optimize by once a node has greater 2n/3 switch to commit phase

				currActor.commitMutex.Lock()

				//timeOutVar := false

				amount := 0

				for _, msg := range currActor.BFTMsgLogs{
					if msg.prevMsgHash != nil && *msg.prevMsgHash == *commitMsg.prevMsgHash && msg.Type == COMMIT{
						amount++
					}
				}

				currActor.commitMutex.Unlock()

				go func(amount int) {
					time.Sleep(time.Millisecond * 500)
					currActor.commitMutex.Lock()
					if uint64(amount) > uint64(2*n/3){
						//Move to finishing phase

						//TODO: Stop commit timeout here

						if !currActor.BFTMsgLogs[*commitMsg.prevMsgHash].commitExpire {

							//Update current chain
							check, err := currActor.chainHandler.ValidateBlock(commitMsg.block)
							if err != nil || !check {
								//TODO: Switch to view change mode
								log.Println("Error in validating block")
								return
							}
							check, err = currActor.chainHandler.InsertBlock(commitMsg.block)
							if err != nil || !check {
								//TODO: Switch to view change mode
								log.Println("Error in inserting block")
								return
							}
							//Increase sequence number
							err = currActor.chainHandler.IncreaseSeqNum()
							if err != nil {
								log.Println("Error in increasing sequence number")
								return
							}

							if currActor.CurrNode.IsProposer {
								currActor.logBlockMutex.Lock()
								log.Println("proposer:", currActor.CurrNode.index)
								currActor.chainHandler.print()
								currActor.logBlockMutex.Unlock()
							}

							currActor.BFTMsgLogs[*commitMsg.prevMsgHash].commitExpire = true

							//// FOR REAL SIMULATE
							//TODO:
							// Update status of proposer
							// Calculate for random proposer (for simulating, we just increase index of node)
							// Send a msg to system for switching to other view change mode

							currActor.CurrNode.IsProposer = false
							rdNum := rand.Intn(int(n) - 0) + 0

							viewChangeMsg := ViewMsg{
								Type: PREPAREVIEWCHANGE,
								Timestamp: uint64(time.Now().Unix()),
								hash: utils.GenerateHashV1(),
								View: currActor.chainHandler.View(),
								SignerID: currActor.CurrNode.index,
								prevMsgHash: nil,
							}

							currActor.Validators[rdNum].consensusEngine.BFTProcess.ViewChangeMsgCh <- viewChangeMsg

						}
					}

					currActor.commitMutex.Unlock()
				}(amount)

				//For starting a view change mode:
				// - The idle timeout expires
				// - The commit timeout expires
				// - The view change timeout expires
				// - Multiple PrePrepare messages are received for the same view and sequence number but different blocks
				// - A Prepare message is received from the primary
				// - f + 1 ViewChange messages are received for the same view

			case viewChangeMsg := <- actor.ViewChangeMsgCh:

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				log.Println(4)
				log.Println(viewChangeMsg)

				switch viewChangeMsg.Type {
				case PREPAREVIEWCHANGE:
					// Update its mode to viewChange mode
					err := currActor.ViewChanging(currActor.chainHandler.View() + 1)
					if err != nil {
						//TODO: Restart new view change mode
						log.Println("Error in viewchanging method")
						return
					}

					err = currActor.chainHandler.IncreaseView()
					if err != nil {
						//TODO: Restart new view change mode
						log.Println("Error in increase view for chain")
						return
					}

					msg := ViewMsg{
						hash: utils.GenerateHashV1(),
						Type:       VIEWCHANGE,
						View:       currActor.chainHandler.View() + 1,
						SignerID:   currActor.CurrNode.index,
						Timestamp:  uint64(time.Now().Unix()),
						prevMsgHash: nil,
						owner: currActor.CurrNode.index,
					}

					//Save view change msg to somewhere
					if currActor.ViewChangeMsgLogs[viewChangeMsg.hash] != nil {
						currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = new(ViewMsg)
						*currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = msg
					}

					//Send messages to other nodes
					for _, element := range currActor.Validators{
						element.consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
					}

				case VIEWCHANGE:

					if viewChangeMsg.View <= currActor.CurrNode.View{
						//TODO: Restart new view change mode
						return
					}

					//Save view change msg to somewhere
					if currActor.ViewChangeMsgLogs[viewChangeMsg.hash] != nil {
						currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = new(ViewMsg)

						//TODO:
						// For purpose of simulate we will set default for this will be true
						// We can random in this function for value true or false
						viewChangeMsg.isValid = true
						//

						*currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = viewChangeMsg
					}

					currActor.viewChangeMutex.Lock()

					//Define message for sending back to primary node
					msg := ViewMsg{
						Type:       BACKVIEWCHANGE,
						View:       viewChangeMsg.View,
						SignerID:   currActor.CurrNode.index,
						Timestamp:  uint64(time.Now().Unix()),
						prevMsgHash: &viewChangeMsg.hash, //view change hash
						owner: currActor.CurrNode.index,
					}

					currActor.ViewChanging(viewChangeMsg.View)

					currActor.ProposalNode.consensusEngine.BFTProcess.BackViewChangeMsgCh <- msg

					currActor.viewChangeMutex.Unlock()

				case NEWVIEW:

					if viewChangeMsg.View <= currActor.CurrNode.View{
						//TODO: Restart new view change mode
						return
					}

					if viewChangeMsg.prevMsgHash == nil{
						//TODO: Restart new view change mode
						return
					}

					if currActor.ViewChangeMsgLogs[*viewChangeMsg.prevMsgHash] != nil {
						//TODO: Restart new view change mode
						return
					}

					if viewChangeMsg.owner != currActor.ViewChangeMsgLogs[*viewChangeMsg.prevMsgHash].owner{
						//TODO: Restart new view change mode
						return
					}

					//Save view change msg to somewhere
					if currActor.ViewChangeMsgLogs[viewChangeMsg.hash] != nil {
						currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = new(ViewMsg)
						*currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = viewChangeMsg
					}

					currActor.viewChangeMutex.Lock()

					//Define message for sending back to primary node
					msg := ViewMsg{
						Type:       VERIFYNEWVIEW,
						View:       viewChangeMsg.View,
						SignerID:   currActor.CurrNode.index,
						Timestamp:  uint64(time.Now().Unix()),
						prevMsgHash: &viewChangeMsg.hash, // type: newview msg hash
						owner: currActor.CurrNode.index,
					}

					for i, element := range currActor.Validators{

						if i == currActor.CurrNode.index || i == currActor.ProposalNode.index{
							continue
						}

						element.consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
					}

					currActor.viewChangeMutex.Unlock()

				case VERIFYNEWVIEW:

					if viewChangeMsg.View <= currActor.CurrNode.View{
						//TODO: Restart new view change mode
						return
					}

					if viewChangeMsg.prevMsgHash == nil{
						//TODO: Restart new view change mode
						return
					}

					if currActor.ViewChangeMsgLogs[*viewChangeMsg.prevMsgHash] != nil {
						//TODO: Restart new view change mode
						return
					}

					newViewMsg := currActor.ViewChangeMsgLogs[*viewChangeMsg.prevMsgHash]

					rootViewChangeMsg := currActor.ViewChangeMsgLogs[*newViewMsg.prevMsgHash]

					if rootViewChangeMsg.isValid{

						currActor.viewChangeMutex.Lock()

						//Define message for sending back to primary node
						msg := ViewMsg{
							Type:       BACKVERIFYNEWVIEW,
							View:       viewChangeMsg.View,
							SignerID:   currActor.CurrNode.index,
							Timestamp:  uint64(time.Now().Unix()),
							prevMsgHash: &newViewMsg.hash, // type: newview msg hash
							owner: currActor.CurrNode.index,
						}

						for i, element := range currActor.Validators{

							if i == viewChangeMsg.owner {
								element.consensusEngine.BFTProcess.BackViewChangeMsgCh <- msg
							}
						}

						currActor.viewChangeMutex.Unlock()
					}

				}

				case backViewChangeMsg := <- actor.BackViewChangeMsgCh:

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				switch backViewChangeMsg.Type {
				case BACKVIEWCHANGE:

					if backViewChangeMsg.prevMsgHash == nil {
						//TODO: Restart new view change mode
						return
					}

					if currActor.ViewChangeMsgLogs[*backViewChangeMsg.prevMsgHash] == nil {
						//TODO: Restart new view change mode
						return
					}

					//Save view change msg to somewhere
					if currActor.ViewChangeMsgLogs[backViewChangeMsg.hash] != nil {
						currActor.ViewChangeMsgLogs[backViewChangeMsg.hash] = new(ViewMsg)
						*currActor.ViewChangeMsgLogs[backViewChangeMsg.hash] = backViewChangeMsg
					} else {
						currActor.ViewChangeMsgLogs[backViewChangeMsg.hash].amount++
					}

					if len(currActor.BackViewChangeMsgCh) == 0{

						currActor.backViewChangeMutex.Lock()

						amount := 0

						for _, msg := range currActor.ViewChangeMsgLogs{
							if msg.prevMsgHash != nil && *msg.prevMsgHash == *backViewChangeMsg.prevMsgHash && msg.Type == BACKVIEWCHANGE{
								amount++
							}
						}

						if uint64(amount) <= uint64(2*n/3){
							//TODO: Restart new view change mode
							log.Println("Amount view change < 2n/3")
							return
						}

						//Send messages to other nodes
						for _, element := range currActor.Validators{
							//Define message for sending back to primary node
							msg := ViewMsg{
								Type:       NEWVIEW,
								View:       backViewChangeMsg.View,
								SignerID:   currActor.CurrNode.index,
								Timestamp:  uint64(time.Now().Unix()),
								prevMsgHash: backViewChangeMsg.prevMsgHash, //viewchange msg hash
								owner: currActor.CurrNode.index,
							}
							element.consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
						}

						currActor.backViewChangeMutex.Unlock()
					}

				case BACKNEWVIEW:

					if backViewChangeMsg.prevMsgHash == nil {
						//TODO: Restart new view change mode
						return
					}

					if currActor.ViewChangeMsgLogs[*backViewChangeMsg.prevMsgHash] == nil {
						//TODO: Restart new view change mode
						return
					}

					//Save view change msg to somewhere
					if currActor.ViewChangeMsgLogs[backViewChangeMsg.hash] != nil {
						currActor.ViewChangeMsgLogs[backViewChangeMsg.hash] = new(ViewMsg)
						*currActor.ViewChangeMsgLogs[backViewChangeMsg.hash] = backViewChangeMsg
					} else {
						currActor.ViewChangeMsgLogs[backViewChangeMsg.hash].amount++
					}


					if len(currActor.BackViewChangeMsgCh) == 0{

						currActor.backViewChangeMutex.Lock()

						amount := 0

						for _, msg := range currActor.ViewChangeMsgLogs{
							if msg.prevMsgHash != nil && *msg.prevMsgHash == *backViewChangeMsg.prevMsgHash && msg.Type == BACKNEWVIEW{
								amount++
							}
						}

						currActor.backViewChangeMutex.Unlock()

						if uint64(amount) <= uint64(2*n/3){
							//TODO: Restart new view change mode
							log.Println("Amount view change < 2n/3")
							return
						}

						if !backViewChangeMsg.backVerifyNewViewExpire{
							currActor.backViewChangeMutex.Lock()

							currActor.BroadcastMsgCh <- true
							currActor.CurrNode.IsProposer = true
							//Turn off all viewchanging mode

							currActor.backViewChangeMutex.Unlock()
						}
					}

				case BACKVERIFYNEWVIEW:

					if backViewChangeMsg.View <= currActor.CurrNode.View{
						//TODO: Restart new view change mode
						return
					}

					if backViewChangeMsg.prevMsgHash == nil{
						//TODO: Restart new view change mode
						return
					}

					if currActor.ViewChangeMsgLogs[*backViewChangeMsg.prevMsgHash] != nil {
						//TODO: Restart new view change mode
						return
					}

					//Save view change msg to somewhere
					if currActor.ViewChangeMsgLogs[backViewChangeMsg.hash] != nil {
						currActor.ViewChangeMsgLogs[backViewChangeMsg.hash] = new(ViewMsg)
						*currActor.ViewChangeMsgLogs[backViewChangeMsg.hash] = backViewChangeMsg
					} else {
						currActor.ViewChangeMsgLogs[backViewChangeMsg.hash].amount++
					}

					newViewMsg := currActor.ViewChangeMsgLogs[*backViewChangeMsg.prevMsgHash]

					currActor.backViewChangeMutex.Lock()

					amount := 0

					for _, msg := range currActor.ViewChangeMsgLogs{
						if msg.prevMsgHash != nil && *msg.prevMsgHash == *backViewChangeMsg.prevMsgHash && msg.Type == BACKVERIFYNEWVIEW{
							amount++
						}
					}

					currActor.backViewChangeMutex.Unlock()

					if uint64(amount) <= uint64(2*n/3){
						//TODO: Restart new view change mode
						log.Println("Amount view change < 2n/3")
						return
					}

					if !backViewChangeMsg.backVerifyNewViewExpire{
						currActor.backViewChangeMutex.Lock()

						msg := ViewMsg{
							Type:       BACKNEWVIEW,
							View:       backViewChangeMsg.View,
							SignerID:   currActor.CurrNode.index,
							Timestamp:  uint64(time.Now().Unix()),
							prevMsgHash: backViewChangeMsg.prevMsgHash,
							owner: currActor.CurrNode.index,
						}

						//Send messages to other nodes
						for i, element := range currActor.Validators{
							//Define message for sending back to primary node

							if i == newViewMsg.owner{
								element.consensusEngine.BFTProcess.BackViewChangeMsgCh <- msg
							}
						}

						backViewChangeMsg.backVerifyNewViewExpire = true

						currActor.backViewChangeMutex.Unlock()
					}
				}
			}
		}
	}()
	return nil
}

func (actor *Actor) startNormalMode(){

	time.Sleep(time.Millisecond * 1000)

	err := pool.start()
	if err != nil{
		return
	}

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

	//log.Println("[Broadcast]", msg.hash)

	msg.block = block

	if actor.BFTMsgLogs[msg.hash] == nil{
		actor.BFTMsgLogs[msg.hash] = new(NormalMsg)
		*actor.BFTMsgLogs[msg.hash] = msg
	} else {
		actor.BFTMsgLogs[msg.hash].Amount++
	}

	for _, member := range actor.Validators{
		//TODO: Research for more effective broadcast way
		// Or may be broadcast by go routine
		//log.Println("Send pre prepare msg")

		if member.index == actor.CurrNode.index{
			continue
		}

		member.consensusEngine.BFTProcess.PrePrepareMsgCh <- msg
	}
}

//updateNormalMode ...
func (actor *Actor) updateNormalMode(view uint64) {
	actor.CurrNode.Mode = NormalMode
}

//ViewChanging ...
func (actor *Actor) ViewChanging(v uint64) error{
	actor.CurrNode.Mode = ViewChangeMode
	return nil
}

func (actor *Actor) initValidators(m map[int]*Node) {
	for i, element := range m{
		actor.Validators[i] = element
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

