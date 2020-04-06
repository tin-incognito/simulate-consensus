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
	prePrepareMutex sync.Mutex
	PrepareMsgCh    chan NormalMsg
	prepareMutex sync.Mutex
	CommitMsgCh chan NormalMsg
	commitMutex sync.Mutex
	ViewChangeMsgCh chan ViewMsg
	viewChangeMutex sync.Mutex
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
	saveMsgMutex sync.Mutex
	timeOutCh chan bool
	errCh chan error
	wg sync.WaitGroup
	amountMsgTimer *time.Timer
	idleTimer *time.Timer
	blockPublishTimer *time.Timer
	commitTimer *time.Timer
	viewChangeTimer *time.Timer
	newBlock *Block
	viewChangeExpire bool
	viewChangeAmount map[int]int
	sendMsgMutex sync.Mutex
	modeMutex sync.Mutex
	switchMutex sync.Mutex
	testCh chan bool
}

func NewActor() *Actor{

	res := &Actor{
		PrePrepareMsgCh:     make (chan NormalMsg, 100),
		PrepareMsgCh:        make (chan NormalMsg, 100),
		CommitMsgCh:         make (chan NormalMsg, 100),
		ViewChangeMsgCh: make (chan ViewMsg, 100),
		BroadcastMsgCh:      make (chan bool, 100),
		isStarted:           true,
		CurrNode:            nil,
		ProposalNode: nil,
		Validators: make(map[int]*Node),
		StopCh: make(chan struct{}, 20),
		chainHandler: &Chain{},
		BFTMsgLogs: make(map[string]*NormalMsg),
		ViewChangeMsgLogs: make(map[string]*ViewMsg),
		wg:                  sync.WaitGroup{},
		timeOutCh: make(chan bool, 100),
		errCh: make(chan error, 100),
		viewChangeAmount: make(map[int]int),
		testCh: make(chan bool),
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

		for {
			select {
			case <- actor.StopCh:
				log.Println(0)
				return

			case err := <- actor.errCh:
				log.Println("err:", err)
				continue

			case _ = <- actor.testCh:
				log.Println("Test channel")
				continue

			case _ = <- actor.BroadcastMsgCh:

				//log.Println("Broadcast msg")

				// This is Pre prepare phase
				currActor := actor.CurrNode.consensusEngine.BFTProcess

				if currActor.CurrNode.Mode != NormalMode{
					//log.Println("Block by normal mode verifier")
					continue
				}

				//TODO:
				// Start idle timeout here

				currActor.idleTimer = time.NewTimer(time.Millisecond * 30000)

				currActor.blockPublishTimer = time.NewTimer(time.Millisecond * 1000)

				//Timeout generate new block
				currActor.wg.Add(1)
				go func(){
					defer func() {
						currActor.blockPublishTimer.Stop()
						currActor.wg.Done()
					}()

					select {
					case <-currActor.timeOutCh:
						//var err error
						block, err := currActor.chainHandler.CreateBlock()
						if err != nil {
							log.Println(err)
							currActor.errCh <- err
						}

						currActor.newBlock = new(Block)
						*currActor.newBlock = *block

					case <-currActor.blockPublishTimer.C:
						time.Sleep(time.Millisecond * 100)

						currActor.wg.Add(1)
						currActor.switchToviewChangeMode()
						currActor.wg.Wait()
					}
				}()

				currActor.timeOutCh <- true
				currActor.wg.Wait()
				///

				msg := NormalMsg{
					hash: utils.GenerateHashV1(),
					Type:      PREPREPARE,
					View:      currActor.chainHandler.View(),
					SeqNum:    currActor.chainHandler.SeqNumber(),
					SignerID:  currActor.CurrNode.index,
					Timestamp: uint64(time.Now().Unix()),
					BlockID:   &currActor.newBlock.Index,
					prevMsgHash: nil,
				}

				msg.block = currActor.newBlock

				currActor.saveMsgMutex.Lock()

				if currActor.BFTMsgLogs[msg.hash] == nil{
					currActor.BFTMsgLogs[msg.hash] = new(NormalMsg)
					*currActor.BFTMsgLogs[msg.hash] = msg
				} else {
					currActor.BFTMsgLogs[msg.hash].Amount++
				}

				currActor.saveMsgMutex.Unlock()

				//TODO:
				// Need to change to send block info to validator of each nodes
				// Then validators will send block info to other block validators
				// After

				for _, member := range currActor.Validators{
					//TODO: Research for more effective broadcast way
					// Or may be broadcast by go routine

					member.consensusEngine.BFTProcess.PrePrepareMsgCh <- msg

				}

			case prePrepareMsg := <- actor.PrePrepareMsgCh:

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				//log.Println("prePrepare msg:", prePrepareMsg, "from:", prePrepareMsg.SignerID, "to:", currActor.CurrNode.index)

				if currActor.CurrNode.Mode != NormalMode {
					//log.Println("Block by normal mode verifier")
					continue
				}

				if !(prePrepareMsg.SignerID == currActor.ProposalNode.index) {
					//log.Println()
					log.Println("!(prePrepareMsg.SignerID == currActor.ProposalNode.index)")
					continue
				}

				//Save it to somewhere else for every node (actor of consensus engine)

				currActor.saveMsgMutex.Lock()

				if currActor.BFTMsgLogs[prePrepareMsg.hash] == nil{
					currActor.BFTMsgLogs[prePrepareMsg.hash] = new(NormalMsg)
					*currActor.BFTMsgLogs[prePrepareMsg.hash] = prePrepareMsg
				} else {
					currActor.BFTMsgLogs[prePrepareMsg.hash].Amount++
				}

				currActor.saveMsgMutex.Unlock()

				//TODO:
				// Reset idle timeout here

				if currActor.CurrNode.IsProposer{
					currActor.idleTimer.Reset(time.Millisecond * 1000)
				}

				// Move to prepare phase

				//TODO:
				// Start commit timeout here
				currActor.commitTimer = time.NewTimer(time.Millisecond * 10000)

				// Generate 1 prepare message for each nodes and
				// Send it from 1 node to n - 1 nodes
				// Therefore each messages from each node will have different hash

				//currActor.prePrepareMutex.Lock()

				currActor.wg.Add(1)
				go func(){
					defer func() {
						currActor.commitTimer.Stop()
						currActor.wg.Done()
					}()

					select {
					case <-currActor.timeOutCh:
						// Node (not primary node) send prepare msg to other nodes

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

							for _, member := range currActor.Validators{
								//log.Println("Send to prepare channel from:", currActor.CurrNode.index, "to:", member.index)
								member.consensusEngine.BFTProcess.PrepareMsgCh <- msg
							}
						}

						currActor.prepareMutex.Unlock()

					case <-currActor.commitTimer.C:

						currActor.wg.Add(1)
						currActor.switchToviewChangeMode()
						currActor.wg.Wait()
					}
				}()

				currActor.timeOutCh <- true
				currActor.wg.Wait()

				//currActor.prePrepareMutex.Unlock()

			case prepareMsg := <- actor.PrepareMsgCh:

				// This is still preparing phase

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				//log.Println("prepare msg:", prepareMsg, "from:", prepareMsg.SignerID, "to:", currActor.CurrNode.index)

				currActor.amountMsgTimer = time.NewTimer(time.Millisecond * 100)

				if currActor.CurrNode.Mode != NormalMode{
					//log.Println("Block by normal mode verifier")
					continue
				}

				if prepareMsg.prevMsgHash == nil {
					//TODO: Switch to view change mode
					log.Println("[prepare] prevMsgHash == null")
					continue
				}

				if currActor.BFTMsgLogs[*prepareMsg.prevMsgHash] == nil {
					//TODO: Switch to view change mode
					log.Println("[prepare] Msg with this prepareMsg.prevMsgHash hash == null" )
					continue
				}

				//Save it to somewhere else for every node (actor of consensus engine)
				currActor.saveMsgMutex.Lock()

				if currActor.BFTMsgLogs[prepareMsg.hash] == nil {
					currActor.BFTMsgLogs[prepareMsg.hash] = new(NormalMsg)
					*currActor.BFTMsgLogs[prepareMsg.hash] = prepareMsg
				} else {
					currActor.BFTMsgLogs[prepareMsg.hash].Amount++
				}

				currActor.saveMsgMutex.Unlock()

				//TODO: Checking for > 2n/3

				//TODO:
				// Set timeout here

				// Wait for end of time
				// If timeout check for 2n > 3

				//currActor.wg.Add(1)
				go func(){

					defer func() {
						currActor.amountMsgTimer.Stop()
						//currActor.wg.Done()
					}()

					select {
					case <-currActor.amountMsgTimer.C:

						// Node (not primary node) send prepare msg to other nodes
						//TODO: Optimize by once a node has 2n/3 switch to commit phase

						currActor.prepareMutex.Lock()

						//amount := 0

						for _, msg := range currActor.BFTMsgLogs{
							if msg.prevMsgHash != nil && *msg.prevMsgHash == *prepareMsg.prevMsgHash && msg.Type == PREPARE{
								currActor.BFTMsgLogs[*prepareMsg.prevMsgHash].Amount++
								//amount++
							}
						}

						currActor.prepareMutex.Unlock()

						currActor.prepareMutex.Lock()

						//log.Println("amount:", currActor.BFTMsgLogs[*prepareMsg.prevMsgHash].Amount)

						if uint64(currActor.BFTMsgLogs[*prepareMsg.prevMsgHash].Amount) <= uint64(2*n/3){

							//TODO: Switch to view change mode

							currActor.wg.Add(1)
							currActor.switchToviewChangeMode()
							currActor.wg.Wait()

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
									//log.Println("Send commit msg from:", currActor.CurrNode.index, "to:", member.index)
									member.consensusEngine.BFTProcess.CommitMsgCh <- msg
								}

								currActor.BFTMsgLogs[*prepareMsg.prevMsgHash].prepareExpire = true
							}
						}

						currActor.prepareMutex.Unlock()
						//return
					}
				}()

				//currActor.wg.Wait()

			case commitMsg := <- actor.CommitMsgCh:

				//This is still committing phase

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				//log.Println("Receive from ", commitMsg.SignerID, "to:", currActor.CurrNode.index)

				if currActor.CurrNode.Mode != NormalMode{
					//log.Println("Block by normal mode verifier")
					continue
				}

				if commitMsg.prevMsgHash == nil {
					//TODO: Switch to view change mode
					// Send faulty primary node msg to other nodes
					log.Println("[commit] commit.prevMsgHash == nil")
					continue
				}

				if currActor.BFTMsgLogs[*commitMsg.prevMsgHash] == nil{
					//TODO: Switch to view change mode
					// Send faulty primary node msg to other nodes
					continue
				}

				//Save it to somewhere else for every node (actor of consensus engine)

				time.Sleep(time.Millisecond * 100)

				currActor.commitMutex.Lock()

				if currActor.BFTMsgLogs[commitMsg.hash] == nil{
					currActor.saveMsgMutex.Lock()
					currActor.BFTMsgLogs[commitMsg.hash] = new(NormalMsg)
					*currActor.BFTMsgLogs[commitMsg.hash] = commitMsg
					currActor.saveMsgMutex.Unlock()
				} else {
					currActor.saveMsgMutex.Lock()
					currActor.BFTMsgLogs[commitMsg.hash].Amount++
					currActor.saveMsgMutex.Unlock()
				}

				currActor.commitMutex.Unlock()

				//TODO: Checking for > 2n/3

				// Node (not primary node) send prepare msg to other nodes

				//TODO: Optimize by once a node has greater 2n/3 switch to commit phase

				currActor.amountMsgTimer = time.NewTimer(time.Millisecond * 200)

				//currActor.wg.Add(1)

				go func(){
					defer func() {
						currActor.amountMsgTimer.Stop()
						//currActor.wg.Done()
					}()

					select {
					case <-currActor.amountMsgTimer.C:

						currActor.commitMutex.Lock()

						for _, msg := range currActor.BFTMsgLogs{
							if msg.prevMsgHash != nil && *msg.prevMsgHash == *commitMsg.prevMsgHash && msg.Type == COMMIT{
								currActor.BFTMsgLogs[*commitMsg.prevMsgHash].Amount++
							}
						}

						currActor.commitMutex.Unlock()

						currActor.commitMutex.Lock()

						//log.Println("currActor.BFTMsgLogs[*commitMsg.prevMsgHash].Amount:", currActor.BFTMsgLogs[*commitMsg.prevMsgHash].Amount)

						if uint64(currActor.BFTMsgLogs[*commitMsg.prevMsgHash].Amount) <= uint64(2*n/3){

							//log.Println("Not enough 2/3 amount of votes")

							currActor.wg.Add(1)
							currActor.switchToviewChangeMode()
							currActor.wg.Wait()

						} else {

							//Move to finishing phase

							//TODO: Stop commit timeout here

							//currActor.commitTimer.Stop()

							if !currActor.BFTMsgLogs[*commitMsg.prevMsgHash].commitExpire {

								currActor.BFTMsgLogs[*commitMsg.prevMsgHash].commitExpire = true

								//Update current chain
								check, err := currActor.chainHandler.ValidateBlock(commitMsg.block)
								if err != nil || !check {
									log.Println("Error in validating block")
									currActor.wg.Add(1)
									currActor.switchToviewChangeMode()
									currActor.wg.Wait()
								}
								check, err = currActor.chainHandler.InsertBlock(commitMsg.block)
								if err != nil || !check {
									log.Println("Error in inserting block")
									currActor.wg.Add(1)
									currActor.switchToviewChangeMode()
									currActor.wg.Wait()
								}
								//Increase sequence number
								err = currActor.chainHandler.IncreaseSeqNum()
								if err != nil {
									log.Println("Error in increasing sequence number")
									currActor.wg.Add(1)
									currActor.switchToviewChangeMode()
									currActor.wg.Wait()
								}

								if currActor.CurrNode.IsProposer {
									currActor.logBlockMutex.Lock()
									log.Println("proposer:", currActor.CurrNode.index)
									currActor.chainHandler.print()
									currActor.logBlockMutex.Unlock()
								}

								//After normal mode

								err = currActor.CurrNode.updateAfterNormalMode()
								if err != nil{
									log.Println(err)
									currActor.wg.Add(1)
									currActor.switchToviewChangeMode()
									currActor.wg.Wait()
								}

								time.Sleep(time.Millisecond * 200)

								currActor.wg.Add(1)
								currActor.switchToviewChangeMode()
								currActor.wg.Wait()
							}
						}

						currActor.commitMutex.Unlock()
					}
				}()

			case viewChangeMsg := <- actor.ViewChangeMsgCh:

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				//log.Println("viewchangeMsg:", viewChangeMsg)

				if currActor.CurrNode.Mode != ViewChangeMode{
					//log.Println("Block by viewchange mode verifier")
					continue
				}

				switch viewChangeMsg.Type {
				case VIEWCHANGE:

					//log.Println("Receive from:", viewChangeMsg.SignerID, "to:", currActor.CurrNode.index, viewChangeMsg)

					//Timeout generate new block
					//currActor.wg.Add(1)
					go func(){
						//defer func() {
						//	currActor.wg.Done()
						//}()

						select {
						case <-currActor.timeOutCh:
							if currActor.CurrNode.Mode == ViewChangeMode{
								if viewChangeMsg.View < currActor.CurrNode.View {
									//TODO: Restart new view change mode
									// Send faulty node ask to become primary node msg to other nodes
									log.Println("[view change] (already in viewchange mode) viewChangeMsg.View <= currActor.CurrNode.View")
									currActor.wg.Add(1)
									currActor.switchToviewChangeMode()
									currActor.wg.Wait()
								}
							} else {
								if viewChangeMsg.View <= currActor.CurrNode.View{
									//TODO: Restart new view change mode
									// Send faulty node ask to become primary node msg to other nodes
									log.Println("[view change] viewChangeMsg.View <= currActor.CurrNode.View")
									currActor.wg.Add(1)
									currActor.switchToviewChangeMode()
									currActor.wg.Wait()
								}
							}

							currActor.viewChangeMutex.Lock()

							//Save view change msg to somewhere
							if currActor.ViewChangeMsgLogs[viewChangeMsg.hash] == nil {
								isDup := false
								for _, element := range currActor.ViewChangeMsgLogs{
									if element.Type == VIEWCHANGE && element.SignerID == viewChangeMsg.SignerID && element.View == viewChangeMsg.View{
										isDup = true
										break
									}
								}

								if !isDup{
									currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = new(ViewMsg)
									*currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = viewChangeMsg
								}
							}

							for _, msg := range currActor.ViewChangeMsgLogs{
								if msg.View == currActor.View() && msg.Type == VIEWCHANGE{
									currActor.saveMsgMutex.Lock()
									currActor.viewChangeAmount[int(currActor.View())]++
									currActor.saveMsgMutex.Unlock()
								}
							}

							currActor.viewChangeMutex.Unlock()

							//log.Println(currActor.viewChangeAmount[int(currActor.View())])

							currActor.amountMsgTimer = time.NewTimer(time.Millisecond * 100)

							go func(){
								select {
								case <-currActor.amountMsgTimer.C:
									currActor.viewChangeMutex.Lock()

									if uint64(currActor.viewChangeAmount[int(currActor.View())]) > uint64(2*n/3){

										if !currActor.viewChangeExpire {

											currActor.viewChangeExpire = true

											if currActor.isPrimaryNode(int(currActor.View())){
												msg := ViewMsg{
													hash: utils.GenerateHashV1(),
													Type:       NEWVIEW,
													View:       viewChangeMsg.View,
													SignerID:   currActor.CurrNode.index,
													Timestamp:  uint64(time.Now().Unix()),
													prevMsgHash: viewChangeMsg.prevMsgHash, //viewchange msg hash
												}

												for _, msg := range currActor.ViewChangeMsgLogs{
													if msg.View == currActor.View() && msg.Type == VIEWCHANGE{
														msg.hashSignedMsgs = append(msg.hashSignedMsgs, msg.hash)
													}
												}

												//Save view change msg to somewhere
												if currActor.ViewChangeMsgLogs[msg.hash] == nil {
													currActor.ViewChangeMsgLogs[msg.hash] = new(ViewMsg)
													*currActor.ViewChangeMsgLogs[msg.hash] = msg
												}

												//Send messages to other nodes
												for _, element := range currActor.Validators{
													//Define message for sending back to primary node
													//log.Println("Send new view msg from primary node")

													element.consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
												}
											}

											//TODO:
											// Start new view of view change mode here
										}

									} else {
										currActor.wg.Add(1)
										currActor.switchToviewChangeMode()
										currActor.wg.Wait()
									}

									currActor.viewChangeMutex.Unlock()
								}
							}()

						case <-currActor.viewChangeTimer.C:
							//time.Sleep(time.Millisecond * 100)

							currActor.wg.Add(1)
							currActor.switchToviewChangeMode()
							currActor.wg.Wait()
						}
					}()

					currActor.timeOutCh <- true

				case NEWVIEW:

					go func(){
						select {
						case <-currActor.timeOutCh:
							currActor.viewChangeExpire = false

							//log.Println("newview msg from ", viewChangeMsg.SignerID, "to:", currActor.CurrNode.index)

							//log.Println(0)

							if viewChangeMsg.View != currActor.CurrNode.View{
								log.Println("[New view] viewChangeMsg.View <= currActor.CurrNode.View")
								currActor.wg.Add(1)
								currActor.switchToviewChangeMode()
								currActor.wg.Wait()
							}

							if viewChangeMsg.SignerID != currActor.CurrNode.index {
								for _, hash := range viewChangeMsg.hashSignedMsgs {
									if currActor.ViewChangeMsgLogs[hash] == nil {
										log.Println("Wrong signed viewchange messages")
										currActor.wg.Add(1)
										currActor.switchToviewChangeMode()
										currActor.wg.Wait()
										break
									}
								}
							}

							//Save view change msg to somewhere
							if currActor.ViewChangeMsgLogs[viewChangeMsg.hash] == nil {
								currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = new(ViewMsg)
								*currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = viewChangeMsg
							}

							currActor.viewChangeTimer.Stop()

							currActor.wg.Add(1)

							//log.Println("Jump to switch to normal mode")

							currActor.switchToNormalMode()

							currActor.wg.Wait()

							if currActor.isPrimaryNode(int(currActor.View())){
								currActor.BroadcastMsgCh <- true
							}

						case <-currActor.viewChangeTimer.C:
							currActor.wg.Add(1)
							currActor.switchToviewChangeMode()
							currActor.wg.Wait()
						}
					}()

					currActor.timeOutCh <- true
				}
			}
		}
	}()
	return nil
}

