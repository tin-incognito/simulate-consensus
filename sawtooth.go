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
	viewChangeNode *Node
	verifyMutex sync.Mutex
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
					continue
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

				msg.block = block

				if currActor.BFTMsgLogs[msg.hash] == nil{
					currActor.BFTMsgLogs[msg.hash] = new(NormalMsg)
					*currActor.BFTMsgLogs[msg.hash] = msg
				} else {
					currActor.BFTMsgLogs[msg.hash].Amount++
				}

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
					//log.Println()
					log.Println("!(prePrepareMsg.SignerID == currActor.ProposalNode.index)")
					continue
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

				//log.Println("prepare msg")

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				if prepareMsg.prevMsgHash == nil {
					//TODO: Switch to view change mode
					log.Println("[prepare] prevMsgHash == null")
					continue
				}

				//log.Println("[prepare] *prepareMsg.prevMsgHash:", *prepareMsg.prevMsgHash)

				currActor.prepareMutex.Lock()
				if currActor.BFTMsgLogs[*prepareMsg.prevMsgHash] == nil {
					//TODO: Switch to view change mode
					log.Println("[prepare] Msg with this prepareMsg.prevMsgHash hash == null" )
					continue
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

					currActor.prepareMutex.Unlock()
				}

			case commitMsg := <- actor.CommitMsgCh:

				//This is still committing phase

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				if commitMsg.prevMsgHash == nil {
					//TODO: Switch to view change mode
					log.Println("[commit] commit.prevMsgHash == nil")
					continue
				}

				if currActor.BFTMsgLogs[*commitMsg.prevMsgHash] == nil{
					//TODO: Switch to view change mode
					continue
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

					if currActor.CurrNode.IsProposer{
						currActor.CurrNode.IsProposer = false
						rdNum := rand.Intn(int(n) - 0) + 0

						//log.Println("rdNum:", rdNum)

						viewChangeMsg := ViewMsg{
							Type: PREPAREVIEWCHANGE,
							Timestamp: uint64(time.Now().Unix()),
							hash: utils.GenerateHashV1(),
							View: currActor.chainHandler.View(),
							SignerID: currActor.CurrNode.index,
							prevMsgHash: nil,
						}

						time.Sleep(time.Millisecond * 200)
						currActor.Validators[rdNum].consensusEngine.BFTProcess.ViewChangeMsgCh <- viewChangeMsg
					}
				}

				currActor.commitMutex.Unlock()

				//For starting a view change mode:
				// - The idle timeout expires
				// - The commit timeout expires
				// - The view change timeout expires
				// - Multiple PrePrepare messages are received for the same view and sequence number but different blocks
				// - A Prepare message is received from the primary
				// - f + 1 ViewChange messages are received for the same view

			case viewChangeMsg := <- actor.ViewChangeMsgCh:

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				switch viewChangeMsg.Type {
				case PREPAREVIEWCHANGE:

					// Update its mode to viewChange mode
					err := currActor.ViewChanging(currActor.CurrNode.View + 1)
					if err != nil {
						//TODO: Restart new view change mode
						log.Println("Error in view changing method")
						continue
					}

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

					currActor.viewChangeNode = currActor.CurrNode

					//Send messages to other nodes
					for i, element := range currActor.Validators{
						if i != currActor.CurrNode.index{
							element.consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
						}
					}

				case VIEWCHANGE:

					if viewChangeMsg.View <= currActor.CurrNode.View{
						//TODO: Restart new view change mode
						log.Println("viewChangeMsg.View <= currActor.CurrNode.View")
						continue
					}

					currActor.viewChangeMutex.Lock()

					//Save view change msg to somewhere
					if currActor.ViewChangeMsgLogs[viewChangeMsg.hash] == nil {
						currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = new(ViewMsg)

						//TODO:
						// For purpose of simulate we will set default for this will be true
						// We can random in this function for value true or false
						viewChangeMsg.isValid = true
						//

						*currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = viewChangeMsg
					}

					currActor.ViewChanging(currActor.CurrNode.View + 1)

					//Define message for sending back to primary node
					msg := ViewMsg{
						hash: utils.GenerateHashV1(),
						Type:       BACKVIEWCHANGE,
						View:       viewChangeMsg.View,
						SignerID:   currActor.CurrNode.index,
						Timestamp:  uint64(time.Now().Unix()),
						prevMsgHash: &viewChangeMsg.hash, //view change hash
						owner: currActor.CurrNode.index,
					}

					//currActor.viewChangeMutex.Lock()
					if currActor.viewChangeNode == nil {
						currActor.viewChangeNode = currActor.Validators[viewChangeMsg.owner]
					}

					currActor.Validators[viewChangeMsg.owner].consensusEngine.BFTProcess.BackViewChangeMsgCh <- msg
					currActor.viewChangeMutex.Unlock()

				case NEWVIEW:

					if viewChangeMsg.prevMsgHash == nil{
						//TODO: Restart new view change mode
						log.Println("[newview] viewChangeMsg.prevMsgHash == nil")
						continue
					}

					if currActor.ViewChangeMsgLogs[*viewChangeMsg.prevMsgHash] == nil {
						//TODO: Restart new view change mode
						log.Println("[newview] currActor.ViewChangeMsgLogs[*viewChangeMsg.prevMsgHash]")
						continue
					}

					if viewChangeMsg.owner != currActor.ViewChangeMsgLogs[*viewChangeMsg.prevMsgHash].owner{
						//TODO: Restart new view change mode
						log.Println("[newview] viewChangeMsg.owner != currActor.ViewChangeMsgLogs[*viewChangeMsg.prevMsgHash].owner")
						continue
					}

					//Save view change msg to somewhere
					if currActor.ViewChangeMsgLogs[viewChangeMsg.hash] == nil {
						currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = new(ViewMsg)
						*currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = viewChangeMsg
					}

					currActor.viewChangeMutex.Lock()

					//Define message for sending back to primary node
					msg := ViewMsg{
						hash: utils.GenerateHashV1(),
						Type:       VERIFYNEWVIEW,
						View:       viewChangeMsg.View,
						SignerID:   currActor.CurrNode.index,
						Timestamp:  uint64(time.Now().Unix()),
						prevMsgHash: &viewChangeMsg.hash, // type: newview msg hash
						owner: currActor.CurrNode.index,
					}

					for i, element := range currActor.Validators{

						if i == currActor.CurrNode.index || i == currActor.viewChangeNode.index{
							continue
						}

						element.consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
					}

					currActor.viewChangeMutex.Unlock()

				case VERIFYNEWVIEW:

					if viewChangeMsg.prevMsgHash == nil{
						//TODO: Restart new view change mode
						return
					}

					if currActor.ViewChangeMsgLogs[*viewChangeMsg.prevMsgHash] == nil {
						//TODO: Restart new view change mode
						log.Println(0)
						continue
					}

					newViewMsg := currActor.ViewChangeMsgLogs[*viewChangeMsg.prevMsgHash]

					rootViewChangeMsg := currActor.ViewChangeMsgLogs[*newViewMsg.prevMsgHash]

					if rootViewChangeMsg.isValid{

						currActor.viewChangeMutex.Lock()

						//Define message for sending back to primary node
						msg := ViewMsg{
							hash: utils.GenerateHashV1(),
							Type:       BACKVERIFYNEWVIEW,
							View:       viewChangeMsg.View,
							SignerID:   currActor.CurrNode.index,
							Timestamp:  uint64(time.Now().Unix()),
							prevMsgHash: &newViewMsg.hash, // type: newview msg hash
							owner: currActor.CurrNode.index,
						}

						currActor.ViewChangeMsgLogs[msg.hash] = new(ViewMsg)
						*currActor.ViewChangeMsgLogs[msg.hash] = msg

						for i, element := range currActor.Validators{

							if currActor.viewChangeNode == nil {
								log.Println("currActor.viewChangeNode == nil:", currActor.CurrNode.index)
							}

							if i != currActor.viewChangeNode.index {
								if i == viewChangeMsg.owner{
									element.consensusEngine.BFTProcess.BackViewChangeMsgCh <- msg
								}
							}
						}

						currActor.viewChangeMutex.Unlock()
					}

				case VIEWCHANGEFINISH:

					//log.Println(VIEWCHANGEFINISH, viewChangeMsg)

					if viewChangeMsg.prevMsgHash == nil {
						//TODO: Restart new view change mode
						log.Println("[back newview] backViewChangeMsg.prevMsgHash")
						continue
					}

					currActor.viewChangeNode = nil
					currActor.updateNormalMode(currActor.CurrNode.View)
					currActor.updateProposalNode(viewChangeMsg.owner)

					if currActor.CurrNode.index == viewChangeMsg.owner{
						currActor.BroadcastMsgCh <- true
					}
					currActor.ViewChangeMsgLogs[*viewChangeMsg.prevMsgHash].backVerifyNewViewExpire = true
				}

				case backViewChangeMsg := <- actor.BackViewChangeMsgCh:

				currActor := actor.CurrNode.consensusEngine.BFTProcess

				switch backViewChangeMsg.Type {
				case BACKVIEWCHANGE:

					if backViewChangeMsg.prevMsgHash == nil {
						log.Println("backViewChangeMsg.prevMsgHash")
						//TODO: Restart new view change mode
						continue
					}

					if currActor.ViewChangeMsgLogs[*backViewChangeMsg.prevMsgHash] == nil {
						log.Println("currActor.ViewChangeMsgLogs[*backViewChangeMsg.prevMsgHash]")
						//TODO: Restart new view change mode
						continue
					}

					//Save view change msg to somewhere
					if currActor.ViewChangeMsgLogs[backViewChangeMsg.hash] == nil {
						currActor.ViewChangeMsgLogs[backViewChangeMsg.hash] = new(ViewMsg)
						*currActor.ViewChangeMsgLogs[backViewChangeMsg.hash] = backViewChangeMsg
					}

					currActor.ViewChangeMsgLogs[backViewChangeMsg.hash].amount++

					if len(currActor.BackViewChangeMsgCh) == 0{

						currActor.backViewChangeMutex.Lock()

						amount := 0

						for _, msg := range currActor.ViewChangeMsgLogs{
							if msg.prevMsgHash != nil && *msg.prevMsgHash == *backViewChangeMsg.prevMsgHash && msg.Type == BACKVIEWCHANGE{
								amount++
							}
						}

						currActor.backViewChangeMutex.Unlock()

						if uint64(amount) <= uint64(2*n/3){
							//TODO: Restart new view change mode
							//log.Println("[back viewchange] Amount view change < 2n/3")
							//currActor.backViewChangeMutex.Unlock()
							continue
						}

						if !backViewChangeMsg.backViewChangeExpire{

							currActor.backViewChangeMutex.Lock()

							msg := ViewMsg{
								hash: utils.GenerateHashV1(),
								Type:       NEWVIEW,
								View:       backViewChangeMsg.View,
								SignerID:   currActor.CurrNode.index,
								Timestamp:  uint64(time.Now().Unix()),
								prevMsgHash: backViewChangeMsg.prevMsgHash, //viewchange msg hash
								owner: currActor.CurrNode.index,
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

							backViewChangeMsg.backViewChangeExpire = true

							currActor.backViewChangeMutex.Unlock()
						}
					}

				case BACKNEWVIEW:

					if backViewChangeMsg.prevMsgHash == nil {
						//TODO: Restart new view change mode
						log.Println("[back newview] backViewChangeMsg.prevMsgHash")
						continue
					}

					if currActor.ViewChangeMsgLogs[*backViewChangeMsg.prevMsgHash] == nil {
						//TODO: Restart new view change mode
						log.Println("[back newview] currActor.ViewChangeMsgLogs[*backViewChangeMsg.prevMsgHash]")
						continue
					}

					//Save view change msg to somewhere
					if currActor.ViewChangeMsgLogs[backViewChangeMsg.hash] == nil {
						currActor.ViewChangeMsgLogs[backViewChangeMsg.hash] = new(ViewMsg)
						*currActor.ViewChangeMsgLogs[backViewChangeMsg.hash] = backViewChangeMsg
					}

					currActor.ViewChangeMsgLogs[backViewChangeMsg.hash].amount++


					if len(currActor.BackViewChangeMsgCh) == 0{

						currActor.backViewChangeMutex.Lock()

						amount := 0

						for _, msg := range currActor.ViewChangeMsgLogs{
							if msg.prevMsgHash != nil && *msg.prevMsgHash == *backViewChangeMsg.prevMsgHash && msg.Type == BACKNEWVIEW{
								amount++
							}
						}

						if uint64(amount) <= uint64(2*n/3){
							//TODO: Restart new view change mode
							//log.Println("[Back newview] Amount view change < 2n/3")
							currActor.backViewChangeMutex.Unlock()
							continue
						}
						if !currActor.ViewChangeMsgLogs[*backViewChangeMsg.prevMsgHash].backVerifyNewViewExpire{

							currActor.CurrNode.IsProposer = true

							msg := ViewMsg{
								hash: utils.GenerateHashV1(),
								Type:       VIEWCHANGEFINISH,
								View:       backViewChangeMsg.View,
								SignerID:   currActor.CurrNode.index,
								Timestamp:  uint64(time.Now().Unix()),
								prevMsgHash: backViewChangeMsg.prevMsgHash,
								owner: currActor.CurrNode.index,
							}

							for _, element := range currActor.Validators{
								element.consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
							}
						}
						currActor.backViewChangeMutex.Unlock()
					}

				case BACKVERIFYNEWVIEW:

					if backViewChangeMsg.prevMsgHash == nil{
						//TODO: Restart new view change mode
						continue
					}

					if currActor.ViewChangeMsgLogs[*backViewChangeMsg.prevMsgHash] == nil {
						//TODO: Restart new view change mode
						continue
					}

					//Save view change msg to somewhere
					if currActor.ViewChangeMsgLogs[backViewChangeMsg.hash] == nil {
						currActor.ViewChangeMsgLogs[backViewChangeMsg.hash] = new(ViewMsg)
						*currActor.ViewChangeMsgLogs[backViewChangeMsg.hash] = backViewChangeMsg
					}

					currActor.backViewChangeMutex.Lock()

					amount := 0

					for _, msg := range currActor.ViewChangeMsgLogs{
						if msg.prevMsgHash != nil && *msg.prevMsgHash == *backViewChangeMsg.prevMsgHash && msg.Type == BACKVERIFYNEWVIEW{
							amount++
						}
					}

					if uint64(amount) <= uint64(2*n/3){
						//TODO: Restart new view change mode
						//log.Println("[back verify newview] Amount view change < 2n/3")
						currActor.backViewChangeMutex.Unlock()
						continue
					}

					if !currActor.ViewChangeMsgLogs[*backViewChangeMsg.prevMsgHash].backViewChangeExpire{

						msg := ViewMsg{
							hash: utils.GenerateHashV1(),
							Type:       BACKNEWVIEW,
							View:       backViewChangeMsg.View,
							SignerID:   currActor.CurrNode.index,
							Timestamp:  uint64(time.Now().Unix()),
							prevMsgHash: backViewChangeMsg.prevMsgHash,
							owner: currActor.CurrNode.index,
						}

						currActor.Validators[currActor.viewChangeNode.index].consensusEngine.BFTProcess.BackViewChangeMsgCh <- msg

						currActor.ViewChangeMsgLogs[*backViewChangeMsg.prevMsgHash].backViewChangeExpire = true
					}
					currActor.backViewChangeMutex.Unlock()
				}
			}
		}
	}()
	return nil
}

//switchToViewChangeMode ...
func (actor *Actor) swtichToViewChangeMode(){

	actor.CurrNode.IsProposer = false
	rdNum := rand.Intn(int(n) - 0) + 0

	//log.Println("rdNum:", rdNum)

	viewChangeMsg := ViewMsg{
		Type: PREPAREVIEWCHANGE,
		Timestamp: uint64(time.Now().Unix()),
		hash: utils.GenerateHashV1(),
		View: actor.chainHandler.View(),
		SignerID: actor.CurrNode.index,
		prevMsgHash: nil,
	}

	time.Sleep(time.Millisecond * 200)

	actor.Validators[rdNum].consensusEngine.BFTProcess.ViewChangeMsgCh <- viewChangeMsg

}

func (actor *Actor) updateProposalNode(index int) {
	actor.ProposalNode = actor.Validators[index]
}

//updateNormalMode ...
func (actor *Actor) updateNormalMode(view uint64) {
	actor.CurrNode.Mode = NormalMode
}

//ViewChanging ...
func (actor *Actor) ViewChanging(v uint64) error{
	actor.CurrNode.Mode = ViewChangeMode
	actor.CurrNode.View = v
	err := actor.chainHandler.IncreaseView()
	if err != nil {
		return err
	}
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

