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
	idleTimer *time.Timer
	blockPublishTimer *time.Timer
	commitTimer *time.Timer
	viewChangeTimer *time.Timer
	newBlock *Block
	viewChangeExpire bool
	viewChangeAmount map[int]int
	sendMsgMutex sync.Mutex
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
		StopCh: make(chan struct{}),
		chainHandler: &Chain{},
		BFTMsgLogs: make(map[string]*NormalMsg),
		ViewChangeMsgLogs: make(map[string]*ViewMsg),
		wg:                  sync.WaitGroup{},
		timeOutCh: make(chan bool, 20),
		errCh: make(chan error, 20),
		viewChangeAmount: make(map[int]int),
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

			case _ = <- actor.BroadcastMsgCh:

				// This is Pre prepare phase
				currActor := actor.CurrNode.consensusEngine.BFTProcess

				if currActor.CurrNode.Mode != NormalMode{
					continue
				}

				//TODO:
				// Start idle timeout here

				currActor.idleTimer = time.NewTimer(time.Millisecond * 1000)

				currActor.blockPublishTimer = time.NewTimer(time.Millisecond * 1000)

				//Timeout generate new block
				currActor.wg.Add(1)
				go func(){
					defer func() {
						currActor.wg.Done()
						currActor.blockPublishTimer.Stop()
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
						//
						time.Sleep(time.Millisecond * 200)
						currActor.swtichToViewChangeMode()
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

					if member.index == currActor.CurrNode.index{
						continue
					}

					member.consensusEngine.BFTProcess.PrePrepareMsgCh <- msg
				}

			case prePrepareMsg := <- actor.PrePrepareMsgCh:

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				if currActor.CurrNode.Mode != NormalMode{
					continue
				}

				if !(prePrepareMsg.SignerID == currActor.ProposalNode.index){
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
				currActor.commitTimer = time.NewTimer(time.Millisecond * 1000)

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

				if currActor.CurrNode.Mode != NormalMode{
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
						currActor.prepareMutex.Unlock()
						continue
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
							go func(member *Node){
								member.consensusEngine.BFTProcess.CommitMsgCh <- msg
							}(member)
						}

						currActor.BFTMsgLogs[*prepareMsg.prevMsgHash].prepareExpire = true
					}

					//TODO:
					// Check a timeout here if yes send faulty primary node msg to other nodes

					currActor.prepareMutex.Unlock()
				}

			case commitMsg := <- actor.CommitMsgCh:

				//This is still committing phase

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				if currActor.CurrNode.Mode != NormalMode{
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
				currActor.saveMsgMutex.Lock()

				if currActor.BFTMsgLogs[commitMsg.hash] == nil{
					currActor.BFTMsgLogs[commitMsg.hash] = new(NormalMsg)
					*currActor.BFTMsgLogs[commitMsg.hash] = commitMsg
				} else {
					currActor.BFTMsgLogs[commitMsg.hash].Amount++
				}

				currActor.saveMsgMutex.Unlock()

				//TODO: Checking for > 2n/3

				// Node (not primary node) send prepare msg to other nodes

				//TODO: Optimize by once a node has greater 2n/3 switch to commit phase

				currActor.commitMutex.Lock()

				amount := 0

				for _, msg := range currActor.BFTMsgLogs{
					if msg.prevMsgHash != nil && *msg.prevMsgHash == *commitMsg.prevMsgHash && msg.Type == COMMIT{
						amount++
					}
				}

				if uint64(amount) <= uint64(2*n/3){
					currActor.commitMutex.Unlock()
					continue
				}

				//Move to finishing phase

				//TODO: Stop commit timeout here

				if !currActor.BFTMsgLogs[*commitMsg.prevMsgHash].commitExpire {

					//Update current chain
					check, err := currActor.chainHandler.ValidateBlock(commitMsg.block)
					if err != nil || !check {
						//TODO: Switch to view change mode
						log.Println("Error in validating block")
						continue
					}
					check, err = currActor.chainHandler.InsertBlock(commitMsg.block)
					if err != nil || !check {
						//TODO: Switch to view change mode
						log.Println("Error in inserting block")
						continue
					}
					//Increase sequence number
					err = currActor.chainHandler.IncreaseSeqNum()
					if err != nil {
						log.Println("Error in increasing sequence number")
						continue
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

					err = currActor.CurrNode.updateAfterNormalMode()
					if err != nil{
						log.Println(err)
						continue
					}

					//time.Sleep(time.Millisecond * 200)
					//currActor.switchToviewChangeMode()

					currActor.ViewChanging(currActor.CurrNode.View + 1)

					msg := ViewMsg{
						hash: utils.GenerateHashV1(),
						Type:       VIEWCHANGE,
						View:       currActor.CurrNode.View,
						SignerID:   currActor.CurrNode.index,
						Timestamp:  uint64(time.Now().Unix()),
						prevMsgHash: nil,
						owner: currActor.CurrNode.index,
					}

					//Save view change msg to somewhere
					if currActor.ViewChangeMsgLogs[msg.hash] == nil {
						currActor.ViewChangeMsgLogs[msg.hash] = new(ViewMsg)
						*currActor.ViewChangeMsgLogs[msg.hash] = msg
					}

					//Send messages to other nodes
					for i, element := range currActor.Validators{
						log.Println("Send from node:", i)
						element.consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
					}

				}

				currActor.commitMutex.Unlock()

			case viewChangeMsg := <- actor.ViewChangeMsgCh:

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				log.Println("from:", viewChangeMsg.SignerID, "to:", currActor.CurrNode.index)

				if currActor.CurrNode.Mode != ViewChangeMode{
					continue
				}

				switch viewChangeMsg.Type {

				case VIEWCHANGE:

					//log.Println("from:", viewChangeMsg.SignerID, "to:", currActor.CurrNode.index)

					if currActor.CurrNode.Mode == ViewChangeMode{
						if viewChangeMsg.View < currActor.CurrNode.View {
							//TODO: Restart new view change mode
							// Send faulty node ask to become primary node msg to other nodes
							log.Println("[view change] (already in viewchange mode) viewChangeMsg.View <= currActor.CurrNode.View")
							continue
						}
					} else {
						if viewChangeMsg.View <= currActor.CurrNode.View{
							//TODO: Restart new view change mode
							// Send faulty node ask to become primary node msg to other nodes
							log.Println("[view change] viewChangeMsg.View <= currActor.CurrNode.View")
							continue
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

							//TODO:
							// For purpose of simulate we will set default for this will be true
							// We can random in this function for value true or false
							viewChangeMsg.isValid = true
							//

							*currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = viewChangeMsg
						}
					}

					currActor.viewChangeMutex.Unlock()

					currActor.viewChangeMutex.Lock()

					for _, msg := range currActor.ViewChangeMsgLogs{
						if msg.View == currActor.View() && msg.Type == VIEWCHANGE{
							currActor.viewChangeAmount[int(currActor.View())]++
						}
					}

					log.Println(currActor.viewChangeAmount[int(currActor.View())])

					if uint64(currActor.viewChangeAmount[int(currActor.View())]) <= uint64(2*n/3){
						//TODO: Restart new view change mode
						// Send faulty node ask to become primary node msg to other nodes
						continue
					}

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
								owner: currActor.CurrNode.index,
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

								//log.Println("Send new view msg to other nodes")

								if currActor.CurrNode.index != element.index{
									element.consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
								}
							}
						}

						//TODO:
						// Start new view of view change mode here
					}

					currActor.viewChangeMutex.Unlock()

				case NEWVIEW:

					if viewChangeMsg.View != currActor.CurrNode.View{
						//TODO: Restart new view change mode
						// Send faulty node ask to become primary node msg to other nodes
						log.Println("[New view] viewChangeMsg.View <= currActor.CurrNode.View")
						continue
					}

					for _, hash := range viewChangeMsg.hashSignedMsgs{
						if currActor.ViewChangeMsgLogs[hash] == nil{
							currActor.switchToviewChangeMode()
							break
						}
					}

					//Save view change msg to somewhere
					if currActor.ViewChangeMsgLogs[viewChangeMsg.hash] == nil {
						currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = new(ViewMsg)
						*currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = viewChangeMsg
					}

					currActor.switchToviewChangeMode()

				}
			}
		}
	}()
	return nil
}

