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
	PrepareMsgCh    chan NormalMsg
	CommitMsgCh chan NormalMsg
	ViewChangeMsgCh chan ViewMsg
	BroadcastMsgCh chan bool
	isStarted      bool
	StopCh         chan struct{}
	CurrNode 	   *Node
	ProposalNode *Node
	Validators map[int]*Node
	chainHandler ChainHandler
	BFTMsgLogs map[string]*NormalMsg
	ViewChangeMsgLogs map[string]*ViewMsg
	timeOutCh chan bool
	errCh chan error
	wg sync.WaitGroup
	postAmountMsgTimerCh chan bool
	prepareAmountMsgTimer *time.Timer
	commitAmountMsgTimer *time.Timer
	idleTimer *time.Timer
	blockPublishTimer *time.Timer
	commitTimer *time.Timer
	viewChangeTimer *time.Timer
	newBlock *Block
	prePrepareMsg map[int]*NormalMsg
	viewChangeExpire bool
	viewChangeAmount map[int]int
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
		postAmountMsgTimerCh: make (chan bool),
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

				log.Println("Broadcast msg")

				// This is Pre prepare phase
				currActor := actor.CurrNode.consensusEngine.BFTProcess

				if currActor.CurrNode.Mode != NormalMode{
					//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[broadcast] Block by normal mode verifier")
					//currActor.switchToviewChangeMode()
					continue
				}

				//TODO:
				// Start idle timeout here


				timerMutex.Lock()
				currActor.idleTimer = time.NewTimer(time.Millisecond * 30000)
				currActor.blockPublishTimer = time.NewTimer(time.Millisecond * 1000)
				timerMutex.Unlock()

				//Timeout generate new block
				currActor.wg.Add(1)
				go func(){
					defer func() {
						currActor.blockPublishTimer.Stop()
						currActor.wg.Done()
					}()

					select {
					case <-currActor.timeOutCh:

						block, err := currActor.chainHandler.CreateBlock()
						if err != nil {
							log.Println(err)
							//currActor.errCh <- err
							return
						}

						currActor.newBlock = new(Block)
						*currActor.newBlock = *block

					case <-currActor.blockPublishTimer.C:

						//time.Sleep(time.Millisecond * 100)

						//switchViewChangeModeMutex.Lock()
						//currActor.switchToviewChangeMode()
						//switchViewChangeModeMutex.Unlock()

						return
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

				saveMsgMutex.Lock()

				if currActor.BFTMsgLogs[msg.hash] == nil{
					currActor.BFTMsgLogs[msg.hash] = new(NormalMsg)
					*currActor.BFTMsgLogs[msg.hash] = msg
				} else {
					currActor.BFTMsgLogs[msg.hash].Amount++
				}

				saveMsgMutex.Unlock()

				//TODO:
				// Need to change to send block info to validator of each nodes
				// Then validators will send block info to other block validators
				// After

				for _, member := range currActor.Validators{
					//TODO: Research for more effective broadcast way
					// Or may be broadcast by go routine

					if !member.IsProposer{
						go func(node *Node){
							node.consensusEngine.BFTProcess.PrePrepareMsgCh <- msg
						}(member)
					}
				}

			case prePrepareMsg := <- actor.PrePrepareMsgCh:

				//log.Println("pre prepare")

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "pre prepare msg:", prePrepareMsg)

				if currActor.CurrNode.Mode != NormalMode {
					//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[pre prepare] Block by normal mode verifier")

					//TODO:
					// How about switching to view change mode

					//switchViewChangeModeMutex.Lock()
					//currActor.switchToviewChangeMode()
					//switchViewChangeModeMutex.Unlock()

					continue
				}

				if !(prePrepareMsg.SignerID == currActor.CurrNode.PrimaryNode.index){

					//switchViewChangeModeMutex.Lock()
					//currActor.switchToviewChangeMode()
					//switchViewChangeModeMutex.Unlock()

					continue
				}

				if currActor.CurrNode.IsProposer {

					log.Println("currActor.CurrNode.IsProposer")

					//switchViewChangeModeMutex.Lock()
					//currActor.switchToviewChangeMode()
					//switchViewChangeModeMutex.Unlock()

					return
				}

				if currActor.prePrepareMsg[int(currActor.CurrNode.View)] != nil{
					//switchViewChangeModeMutex.Lock()
					//currActor.switchToviewChangeMode()
					//switchViewChangeModeMutex.Unlock()
					continue
				}

				//if !(prePrepareMsg.SignerID == currActor.ProposalNode.index) {
				//	//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[pre prepare] !(prePrepareMsg.SignerID == currActor.ProposalNode.index)")
				//
				//	//switchViewChangeModeMutex.Lock()
				//	//currActor.switchToviewChangeMode()
				//	//switchViewChangeModeMutex.Unlock()
				//
				//	continue
				//}

				//Save it to somewhere else for every node (actor of consensus engine)

				saveMsgMutex.Lock()

				if currActor.BFTMsgLogs[prePrepareMsg.hash] == nil{
					currActor.BFTMsgLogs[prePrepareMsg.hash] = new(NormalMsg)
					*currActor.BFTMsgLogs[prePrepareMsg.hash] = prePrepareMsg
				} else {
					currActor.BFTMsgLogs[prePrepareMsg.hash].Amount++
				}

				currActor.prePrepareMsg[int(currActor.CurrNode.View)] = new(NormalMsg)
				*currActor.prePrepareMsg[int(currActor.CurrNode.View)] = prePrepareMsg

				saveMsgMutex.Unlock()

				//Reset idle time out
				timerMutex.Lock()
				if currActor.CurrNode.IsProposer{
					currActor.idleTimer.Reset(time.Millisecond * 1000)
				}

				// Move to prepare phase

				//TODO:
				// Start commit timeout here
				currActor.commitTimer = time.NewTimer(time.Millisecond * 10000)
				timerMutex.Unlock()

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

						prepareMutex.Lock()

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
								go func(node *Node){
									node.consensusEngine.BFTProcess.PrepareMsgCh <- msg
								}(member)
							}
						}

						prepareMutex.Unlock()

						return

					case <-currActor.commitTimer.C:

						//switchViewChangeModeMutex.Lock()
						//currActor.switchToviewChangeMode()
						//switchViewChangeModeMutex.Unlock()

						return
					}
				}()

				currActor.timeOutCh <- true
				currActor.wg.Wait()

			case prepareMsg := <- actor.PrepareMsgCh:

				// This is still preparing phase

				//log.Println("prepare:", prepareMsg)

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				timerMutex.Lock()
				if currActor.prepareAmountMsgTimer == nil{
					currActor.prepareAmountMsgTimer = time.NewTimer(time.Millisecond * 100) // Race condition
				}
				timerMutex.Unlock()

				//currActor.wg.Add(1)
				go func(){

					select {
					case <-currActor.timeOutCh:

						log.Println("prepare:", prepareMsg)

						if currActor.CurrNode.Mode != NormalMode{
							//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[prepare] Block by normal mode verifier")

							//switchViewChangeModeMutex.Lock()
							//currActor.switchToviewChangeMode()
							//switchViewChangeModeMutex.Unlock()

							return
						}

						if prepareMsg.prevMsgHash == nil {

							//log.Println("[prepare] prepareMsg.prevMsgHash")

							//switchViewChangeModeMutex.Lock()
							//currActor.switchToviewChangeMode()
							//switchViewChangeModeMutex.Unlock()

							return
						}

						saveMsgMutex.Lock()

						if currActor.BFTMsgLogs[*prepareMsg.prevMsgHash] == nil {

							//log.Println("[prepare] currActor.BFTMsgLogs[*prepareMsg.prevMsgHash]")

							//switchViewChangeModeMutex.Lock()
							//currActor.switchToviewChangeMode()
							//switchViewChangeModeMutex.Unlock()

							return
						}

						//Save it to somewhere else for every node (actor of consensus engine)
						if currActor.BFTMsgLogs[prepareMsg.hash] == nil {
							currActor.BFTMsgLogs[prepareMsg.hash] = new(NormalMsg)
							*currActor.BFTMsgLogs[prepareMsg.hash] = prepareMsg
						} else {
							currActor.BFTMsgLogs[prepareMsg.hash].Amount++
						}

						//currActor.wg.Done()

						saveMsgMutex.Unlock()
					}
				}()

				currActor.timeOutCh <- true
				//time.Sleep(time.Millisecond * 500)
				//currActor.wg.Wait()

			//case commitMsg := <- actor.CommitMsgCh:
			//
			//	//This is still committing phase
			//
			//	//log.Println("commit")
			//
			//	currActor := actor.CurrNode.consensusEngine.BFTProcess
			//
			//	//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "commitMsg:", commitMsg)
			//	//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[commit] mode:", currActor.CurrNode.Mode)
			//
			//	if currActor.CurrNode.Mode != NormalMode{
			//		//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[commit] Block by normal mode verifier")
			//		//switchViewChangeModeMutex.Lock()
			//		//currActor.switchToviewChangeMode()
			//		//switchViewChangeModeMutex.Unlock()
			//		continue
			//	}
			//
			//	if commitMsg.prevMsgHash == nil {
			//		//TODO: Switch to view change mode
			//		// Send faulty primary node msg to other nodes
			//		//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[commit] commit.prevMsgHash == nil")
			//		switchViewChangeModeMutex.Lock()
			//		currActor.switchToviewChangeMode()
			//		switchViewChangeModeMutex.Unlock()
			//		continue
			//	}
			//
			//	if currActor.BFTMsgLogs[*commitMsg.prevMsgHash] == nil{
			//		//TODO: Switch to view change mode
			//		// Send faulty primary node msg to other nodes
			//		//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[commit] currActor.BFTMsgLogs[*commitMsg.prevMsgHash] == nil")
			//		switchViewChangeModeMutex.Lock()
			//		currActor.switchToviewChangeMode()
			//		switchViewChangeModeMutex.Unlock()
			//		continue
			//	}
			//
			//	//Save it to somewhere else for every node (actor of consensus engine)
			//
			//	time.Sleep(time.Millisecond * 100)
			//
			//	commitMutex.Lock()
			//
			//	if currActor.BFTMsgLogs[commitMsg.hash] == nil{
			//		saveMsgMutex.Lock()
			//		currActor.BFTMsgLogs[commitMsg.hash] = new(NormalMsg) //Race condition
			//		*currActor.BFTMsgLogs[commitMsg.hash] = commitMsg
			//		saveMsgMutex.Unlock()
			//	} else {
			//		saveMsgMutex.Lock()
			//		currActor.BFTMsgLogs[commitMsg.hash].Amount++
			//		saveMsgMutex.Unlock()
			//	}
			//
			//	commitMutex.Unlock()
			//
			//	//TODO: Checking for > 2n/3
			//
			//	// Node (not primary node) send prepare msg to other nodes
			//
			//	//TODO: Optimize by once a node has greater 2n/3 switch to commit phase
			//
			//	msgTimerMutex.Lock()
			//	currActor.amountMsgTimer = time.NewTimer(time.Millisecond * 200) //Race condition
			//	msgTimerMutex.Unlock()
			//
			//	//currActor.wg.Add(1)
			//
			//	go func(){
			//		defer func() {
			//			msgTimerMutex.Lock()
			//			currActor.amountMsgTimer.Stop()
			//			msgTimerMutex.Unlock()
			//			//currActor.wg.Done()
			//		}()
			//
			//		select {
			//		case <-currActor.amountMsgTimer.C:
			//
			//			commitMutex.Lock()
			//
			//			for _, msg := range currActor.BFTMsgLogs{
			//				if msg.prevMsgHash != nil && *msg.prevMsgHash == *commitMsg.prevMsgHash && msg.Type == COMMIT{
			//					currActor.BFTMsgLogs[*commitMsg.prevMsgHash].Amount++
			//				}
			//			}
			//
			//			commitMutex.Unlock()
			//
			//			commitMutex.Lock()
			//
			//			//log.Println("currActor.BFTMsgLogs[*commitMsg.prevMsgHash].Amount:", currActor.BFTMsgLogs[*commitMsg.prevMsgHash].Amount)
			//
			//			if uint64(currActor.BFTMsgLogs[*commitMsg.prevMsgHash].Amount) <= uint64(2*n/3){
			//				//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[commit] uint64(currActor.BFTMsgLogs[*commitMsg.prevMsgHash].Amount) <= 2n/3")
			//				switchViewChangeModeMutex.Lock()
			//				currActor.switchToviewChangeMode()
			//				switchViewChangeModeMutex.Unlock()
			//				return
			//			}
			//
			//			//Move to finishing phase
			//
			//			//TODO: Stop commit timeout here
			//
			//			//currActor.commitTimer.Stop()
			//
			//			if !currActor.BFTMsgLogs[*commitMsg.prevMsgHash].commitExpire {
			//
			//				currActor.BFTMsgLogs[*commitMsg.prevMsgHash].commitExpire = true
			//
			//				//Update current chain
			//				check, err := currActor.chainHandler.ValidateBlock(commitMsg.block)
			//				if err != nil || !check {
			//					log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[commit] Error in validating block")
			//					//currActor.wg.Add(1)
			//					currActor.switchToviewChangeMode()
			//					return
			//					//currActor.wg.Wait()
			//				}
			//				check, err = currActor.chainHandler.InsertBlock(commitMsg.block)
			//				if err != nil || !check {
			//					//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[commit] Error in inserting block")
			//					//currActor.wg.Add(1)
			//					switchViewChangeModeMutex.Lock()
			//					currActor.switchToviewChangeMode()
			//					switchViewChangeModeMutex.Unlock()
			//					return
			//					//currActor.wg.Wait()
			//				}
			//				//Increase sequence number
			//				err = currActor.chainHandler.IncreaseSeqNum()
			//				if err != nil {
			//					//log.Println("View", currActor.CurrNode.View, "[commit] Error in increasing sequence number")
			//					//currActor.wg.Add(1)
			//					switchViewChangeModeMutex.Lock()
			//					currActor.switchToviewChangeMode()
			//					switchViewChangeModeMutex.Unlock()
			//					return
			//					//currActor.wg.Wait()
			//				}
			//
			//				if currActor.CurrNode.IsProposer { //Race condition
			//					//logBlockMutex.Lock()
			//					//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[commit] proposer:", currActor.CurrNode.index)
			//					currActor.chainHandler.print()
			//					//logBlockMutex.Unlock()
			//				}
			//
			//				//After normal mode
			//				updateModeMutex.Lock()
			//				err = currActor.CurrNode.updateAfterNormalMode()
			//				if err != nil{
			//					//log.Println(err)
			//					switchViewChangeModeMutex.Lock()
			//					currActor.switchToviewChangeMode()
			//					switchViewChangeModeMutex.Unlock()
			//					return
			//				}
			//				updateModeMutex.Unlock()
			//
			//				time.Sleep(time.Millisecond * 100)
			//
			//				switchViewChangeModeMutex.Lock()
			//				currActor.switchToviewChangeMode()
			//				switchViewChangeModeMutex.Unlock()
			//
			//			}
			//
			//			commitMutex.Unlock()
			//		}
			//	}()

			//case viewChangeMsg := <- actor.ViewChangeMsgCh:
			//
			//	currActor := actor.CurrNode.consensusEngine.BFTProcess
			//
			//	if currActor.CurrNode.Mode != ViewChangeMode{
			//		//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[viewchange] Block by viewchange mode verifier")
			//		//currActor.switchToNormalMode() //Race condition
			//		continue
			//	}
			//
			//	switch viewChangeMsg.Type {
			//	case VIEWCHANGE:
			//
			//		//log.Println("Viewchange")
			//
			//		//Timeout generate new block
			//		//currActor.wg.Add(1)
			//		go func(){
			//			//defer func() {
			//			//	currActor.wg.Done()
			//			//}()
			//
			//			select {
			//			case <-currActor.timeOutCh:
			//				if currActor.CurrNode.Mode == ViewChangeMode{
			//
			//					//log.Println("View", currActor.CurrNode.View ,"View change msg from:", viewChangeMsg.SignerID, "to:", currActor.CurrNode.index)
			//
			//					if viewChangeMsg.View < currActor.CurrNode.View {
			//						//TODO: Restart new view change mode
			//						// Send faulty node ask to become primary node msg to other nodes
			//						//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[view change] (already in viewchange mode) viewChangeMsg.View <= currActor.CurrNode.View")
			//						switchViewChangeModeMutex.Lock()
			//						currActor.switchToviewChangeMode()
			//						switchViewChangeModeMutex.Unlock()
			//						return
			//					}
			//				} else {
			//					if viewChangeMsg.View <= currActor.CurrNode.View{
			//						//TODO: Restart new view change mode
			//						// Send faulty node ask to become primary node msg to other nodes
			//						//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[view change] viewChangeMsg.View <= currActor.CurrNode.View")
			//						switchViewChangeModeMutex.Lock()
			//						currActor.switchToviewChangeMode()
			//						switchViewChangeModeMutex.Unlock()
			//						return
			//						//currActor.wg.Wait()
			//					}
			//				}
			//
			//				viewChangeMutex.Lock()
			//
			//				//Save view change msg to somewhere
			//				if currActor.ViewChangeMsgLogs[viewChangeMsg.hash] == nil {
			//					isDup := false
			//					for _, element := range currActor.ViewChangeMsgLogs{
			//						if element.Type == VIEWCHANGE && element.SignerID == viewChangeMsg.SignerID && element.View == viewChangeMsg.View{
			//							isDup = true
			//							break
			//						}
			//					}
			//
			//					if !isDup{
			//
			//						//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[view change] currActor.ViewChangeMsgLogs", currActor.ViewChangeMsgLogs)
			//						//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[view change] viewChangeMsg.hash", viewChangeMsg.hash)
			//
			//						currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = new(ViewMsg)
			//						*currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = viewChangeMsg
			//					}
			//				}
			//
			//				for _, msg := range currActor.ViewChangeMsgLogs{
			//					if msg.View == currActor.View() && msg.Type == VIEWCHANGE{
			//						saveMsgMutex.Lock()
			//						currActor.viewChangeAmount[int(currActor.View())]++
			//						saveMsgMutex.Unlock()
			//					}
			//				}
			//
			//				viewChangeMutex.Unlock()
			//
			//				timerMutex.Lock()
			//				currActor.amountMsgTimer = time.NewTimer(time.Millisecond * 100) //Race condition
			//				timerMutex.Unlock()
			//
			//				go func(){
			//					select {
			//					case <-currActor.amountMsgTimer.C: //Race condition
			//						viewChangeMutex.Lock()
			//
			//						if uint64(currActor.viewChangeAmount[int(currActor.View())]) > uint64(2*n/3){
			//
			//							if !currActor.viewChangeExpire {
			//
			//								currActor.viewChangeExpire = true
			//
			//								if currActor.isPrimaryNode(int(currActor.View())){
			//									msg := ViewMsg{
			//										hash: utils.GenerateHashV1(),
			//										Type:       NEWVIEW,
			//										View:       viewChangeMsg.View,
			//										SignerID:   currActor.CurrNode.index,
			//										Timestamp:  uint64(time.Now().Unix()),
			//										prevMsgHash: viewChangeMsg.prevMsgHash, //viewchange msg hash
			//									}
			//
			//									for _, msg := range currActor.ViewChangeMsgLogs{
			//										if msg.View == currActor.View() && msg.Type == VIEWCHANGE{
			//											msg.hashSignedMsgs = append(msg.hashSignedMsgs, msg.hash)
			//										}
			//									}
			//
			//									//Save view change msg to somewhere
			//									if currActor.ViewChangeMsgLogs[msg.hash] == nil {
			//										currActor.ViewChangeMsgLogs[msg.hash] = new(ViewMsg)
			//										*currActor.ViewChangeMsgLogs[msg.hash] = msg
			//									}
			//
			//									//Send messages to other nodes
			//									for _, element := range currActor.Validators{
			//										//Define message for sending back to primary node
			//
			//										go func(node *Node){
			//											node.consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
			//										}(element)
			//									}
			//								}
			//
			//								//TODO:
			//								// Start new view of view change mode here
			//
			//							}
			//
			//						} else {
			//							//currActor.wg.Add(1)
			//							switchViewChangeModeMutex.Lock()
			//							currActor.switchToviewChangeMode()
			//							switchViewChangeModeMutex.Unlock()
			//							return
			//							//currActor.wg.Wait()
			//						}
			//
			//						viewChangeMutex.Unlock()
			//					}
			//				}()
			//
			//			case <-currActor.viewChangeTimer.C:
			//				time.Sleep(time.Millisecond * 100)
			//
			//				//currActor.wg.Add(1)
			//				switchViewChangeModeMutex.Lock()
			//				currActor.switchToviewChangeMode()
			//				switchViewChangeModeMutex.Unlock()
			//				return
			//				//currActor.wg.Wait()
			//			}
			//		}()
			//
			//		timeOutChMutex.Lock()
			//		currActor.timeOutCh <- true
			//		timeOutChMutex.Unlock()
			//
			//	case NEWVIEW:
			//
			//		//log.Println("Newview")
			//
			//		//log.Println("View", currActor.CurrNode.View , "New view msg from:", viewChangeMsg.SignerID, "to:", currActor.CurrNode.index)
			//
			//		go func(){
			//			select {
			//			case <-currActor.timeOutCh:
			//				currActor.viewChangeExpire = false
			//
			//				if viewChangeMsg.View != currActor.CurrNode.View{
			//					//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[New view] viewChangeMsg.View <= currActor.CurrNode.View")
			//					switchViewChangeModeMutex.Lock()
			//					currActor.switchToviewChangeMode()
			//					switchViewChangeModeMutex.Unlock()
			//					return
			//				}
			//
			//				if viewChangeMsg.SignerID != currActor.CurrNode.index {
			//					for _, hash := range viewChangeMsg.hashSignedMsgs {
			//						if currActor.ViewChangeMsgLogs[hash] == nil {
			//							//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[New view] Wrong signed viewchange messages")
			//
			//							switchViewChangeModeMutex.Lock()
			//							currActor.switchToviewChangeMode()
			//							switchViewChangeModeMutex.Unlock()
			//
			//							break
			//						}
			//					}
			//				}
			//
			//				//Save view change msg to somewhere
			//				if currActor.ViewChangeMsgLogs[viewChangeMsg.hash] == nil {
			//					currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = new(ViewMsg)
			//					*currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = viewChangeMsg
			//				}
			//
			//				currActor.viewChangeTimer.Stop()
			//
			//				//currActor.wg.Add(1)
			//
			//				//log.Println("Jump to switch to normal mode")
			//
			//				switchNormalModeMutex.Lock()
			//				currActor.switchToNormalMode()
			//
			//				if currActor.isPrimaryNode(int(currActor.View())){
			//					currActor.BroadcastMsgCh <- true
			//				} else {
			//					time.Sleep(time.Millisecond * 50)
			//				}
			//
			//				switchNormalModeMutex.Unlock()
			//
			//			case <-currActor.viewChangeTimer.C:
			//				//currActor.wg.Add(1)
			//				switchViewChangeModeMutex.Lock()
			//				currActor.switchToviewChangeMode()
			//				switchViewChangeModeMutex.Unlock()
			//				return
			//				//currActor.wg.Wait()
			//			}
			//		}()
			//
			//		currActor.timeOutCh <- true
			//	}
			case <-actor.prepareAmountMsgTimer.C:

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				log.Println(1)

				prepareMutex.Lock()

				prePrepareMsg := currActor.prePrepareMsg[int(currActor.CurrNode.index)]

				for _, msg := range currActor.BFTMsgLogs{
					if msg.prevMsgHash != nil && *msg.prevMsgHash == *prepareMsg.prevMsgHash && msg.Type == PREPARE {
						currActor.BFTMsgLogs[*prepareMsg.prevMsgHash].Amount++
					}
				}

				//log.Println("uint64(currActor.BFTMsgLogs[*prepareMsg.prevMsgHash].Amount):", uint64(currActor.BFTMsgLogs[*prepareMsg.prevMsgHash].Amount))

				if uint64(currActor.BFTMsgLogs[*prepareMsg.prevMsgHash].Amount) <= uint64(2*n/3) {
					//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[prepare] Amount of prepare msg is <= 2n/3")

					//switchViewChangeModeMutex.Lock()
					//currActor.switchToviewChangeMode()
					//switchViewChangeModeMutex.Unlock()

					return
				}

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
						go func(node *Node){
							log.Println("Send to commit channel")
							node.consensusEngine.BFTProcess.CommitMsgCh <- msg
						}(member)
					}
					currActor.BFTMsgLogs[*prepareMsg.prevMsgHash].prepareExpire = true
				}

				currActor.postAmountMsgTimerCh <- true
				prepareMutex.Unlock()
				return
			}
		}
	}()
	return nil
}

