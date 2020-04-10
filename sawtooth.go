package main

import (
	"github.com/tin-incognito/simulate-consensus/utils"
	"log"
	"time"
)

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

			case test := <- actor.stuckCh:
				log.Println(test)

			case _ = <- actor.BroadcastMsgCh:

				//log.Println("Broadcast msg")

				// This is Pre prepare phase
				//currActor := actor.CurrNode.consensusEngine.BFTProcess

				if actor.CurrNode.Mode != NormalMode{
					//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[broadcast] Block by normal mode verifier")
					//currActor.switchToviewChangeMode()
					continue
				}

				//TODO:
				// Start idle timeout here


				timerMutex.Lock()
				actor.idleTimer = time.NewTimer(time.Millisecond * 30000)
				actor.blockPublishTimer = time.NewTimer(time.Millisecond * 1000)
				timerMutex.Unlock()

				//Timeout generate new block
				actor.wg.Add(1)
				go func(){
					defer func() {
						actor.blockPublishTimer.Stop()
						actor.wg.Done()
					}()

					select {
					case <- actor.timeOutCh:

						block, err := actor.chainHandler.CreateBlock()
						if err != nil {
							log.Println(err)
							//currActor.errCh <- err
							return
						}

						actor.newBlock = new(Block)
						*actor.newBlock = *block

					case <- actor.blockPublishTimer.C:

						//time.Sleep(time.Millisecond * 100)

						//switchViewChangeModeMutex.Lock()
						//currActor.switchToviewChangeMode()
						//switchViewChangeModeMutex.Unlock()

						return
					}
				}()

				actor.timeOutCh <- true
				actor.wg.Wait()
				///

				msg := NormalMsg{
					hash: utils.GenerateHashV1(),
					Type:      PREPREPARE,
					View:      actor.chainHandler.View(),
					SeqNum:    actor.chainHandler.SeqNumber(),
					SignerID:  actor.CurrNode.index,
					Timestamp: uint64(time.Now().Unix()),
					BlockID:   &actor.newBlock.Index,
					prevMsgHash: nil,
				}

				msg.block = actor.newBlock

				saveMsgMutex.Lock()

				if actor.BFTMsgLogs[msg.hash] == nil{
					actor.BFTMsgLogs[msg.hash] = new(NormalMsg)
					*actor.BFTMsgLogs[msg.hash] = msg
				} else {
					actor.BFTMsgLogs[msg.hash].Amount++
				}

				actor.prePrepareMsg[int(actor.CurrNode.View)] = new(NormalMsg)
				*actor.prePrepareMsg[int(actor.CurrNode.View)] = msg

				saveMsgMutex.Unlock()

				//TODO:
				// Need to change to send block info to validator of each nodes
				// Then validators will send block info to other block validators
				// After

				for _, member := range actor.Validators{
					//TODO: Research for more effective broadcast way
					// Or may be broadcast by go routine

					if !member.IsProposer{
						go func(node *Node){
							node.consensusEngine.BFTProcess.PrePrepareMsgCh <- msg
						}(member)
					}
				}

			case prePrepareMsg := <- actor.PrePrepareMsgCh:

				//log.Println("pre prepare msg:", prePrepareMsg)

				if actor.CurrNode.Mode != NormalMode {
					//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[pre prepare] Block by normal mode verifier")

					//TODO:
					// How about switching to view change mode

					//switchViewChangeModeMutex.Lock()
					//currActor.switchToviewChangeMode()
					//switchViewChangeModeMutex.Unlock()

					continue
				}

				if !(prePrepareMsg.SignerID == actor.CurrNode.PrimaryNode.index){

					//switchViewChangeModeMutex.Lock()
					//currActor.switchToviewChangeMode()
					//switchViewChangeModeMutex.Unlock()

					continue
				}

				if actor.CurrNode.IsProposer {

					log.Println("currActor.CurrNode.IsProposer")

					//switchViewChangeModeMutex.Lock()
					//currActor.switchToviewChangeMode()
					//switchViewChangeModeMutex.Unlock()

					return
				}

				if actor.prePrepareMsg[int(actor.CurrNode.View)] != nil{
					//switchViewChangeModeMutex.Lock()
					//currActor.switchToviewChangeMode()
					//switchViewChangeModeMutex.Unlock()
					continue
				}

				saveMsgMutex.Lock()

				if actor.BFTMsgLogs[prePrepareMsg.hash] == nil{
					actor.BFTMsgLogs[prePrepareMsg.hash] = new(NormalMsg)
					*actor.BFTMsgLogs[prePrepareMsg.hash] = prePrepareMsg
				} else {
					actor.BFTMsgLogs[prePrepareMsg.hash].Amount++
				}

				actor.prePrepareMsg[int(actor.CurrNode.View)] = new(NormalMsg)
				*actor.prePrepareMsg[int(actor.CurrNode.View)] = prePrepareMsg

				saveMsgMutex.Unlock()

				//Reset idle time out
				timerMutex.Lock()
				if actor.CurrNode.IsProposer{
					actor.idleTimer.Reset(time.Millisecond * 1000)
				}

				// Move to prepare phase

				//TODO:
				// Start commit timeout here
				actor.commitTimer = time.NewTimer(time.Millisecond * 10000)
				timerMutex.Unlock()

				// Generate 1 prepare message for each nodes and
				// Send it from 1 node to n - 1 nodes
				// Therefore each messages from each node will have different hash

				//currActor.prePrepareMutex.Lock()

				actor.wg.Add(1)
				go func(){
					defer func() {
						actor.commitTimer.Stop()
						actor.wg.Done()
					}()

					select {
					case <- actor.timeOutCh:

						// Node (not primary node) send prepare msg to other nodes

						prepareMutex.Lock()

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

							for _, member := range actor.Validators{
								go func(node *Node){
									log.Println("Send message from:", node.index, "to:", member.index)
									node.consensusEngine.BFTProcess.PrepareMsgCh <- msg
								}(member)
							}
						}

						prepareMutex.Unlock()

						return

					case <-actor.commitTimer.C:

						//switchViewChangeModeMutex.Lock()
						//currActor.switchToviewChangeMode()
						//switchViewChangeModeMutex.Unlock()

						return
					}
				}()

				actor.timeOutCh <- true
				actor.wg.Wait()

			case prepareMsg := <- actor.PrepareMsgCh:

				log.Println("prepareMsg:", prepareMsg)

				// This is still preparing phase

				timerMutex.Lock()
				if actor.prepareAmountMsgTimer == nil{
					actor.prepareAmountMsgTimer = time.NewTimer(time.Millisecond * 100) // Race condition

					msgTimer := MsgTimer{ Type:PREPARE }
					go func (){
						actor.msgTimerCh <- msgTimer
					}()

				}
				timerMutex.Unlock()

				actor.wg.Add(1)
				go func(){

					select {
					case <- actor.prepareMsgTimerCh:

						if actor.CurrNode.Mode != NormalMode{
							//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[prepare] Block by normal mode verifier")

							//switchViewChangeModeMutex.Lock()
							//currActor.switchToviewChangeMode()
							//switchViewChangeModeMutex.Unlock()

							return
						}

						PrepareMapMutex.Lock()

						if prepareMsg.prevMsgHash == nil {

							//log.Println("[prepare] prepareMsg.prevMsgHash")

							//switchViewChangeModeMutex.Lock()
							//currActor.switchToviewChangeMode()
							//switchViewChangeModeMutex.Unlock()

							return
						}

						if actor.BFTMsgLogs[*prepareMsg.prevMsgHash] == nil {

							//switchViewChangeModeMutex.Lock()
							//currActor.switchToviewChangeMode()
							//switchViewChangeModeMutex.Unlock()

							return
						}

						//Save it to somewhere else for every node (actor of consensus engine)
						if actor.BFTMsgLogs[prepareMsg.hash] == nil {
							actor.BFTMsgLogs[prepareMsg.hash] = new(NormalMsg)
							*actor.BFTMsgLogs[prepareMsg.hash] = prepareMsg
						}
						actor.BFTMsgLogs[prepareMsg.hash].Amount++

						actor.wg.Done()

						PrepareMapMutex.Unlock()
					}
				}()

				actor.prepareMsgTimerCh <- true
				actor.wg.Wait()

			case commitMsg := <- actor.CommitMsgCh:

				//This is still committing phase

				//log.Println(commitMsg)

				msgTimerMutex.Lock()
				if actor.commitAmountMsgTimer == nil{
					actor.commitAmountMsgTimer = time.NewTimer(time.Millisecond * 200)

					msgTimer := MsgTimer{ Type:COMMIT }
					go func (){
						actor.msgTimerCh <- msgTimer
					}()
				}
				msgTimerMutex.Unlock()

				actor.wg.Add(1)
				go func(){

					select {
					case <- actor.commitMsgTimerCh:

						if actor.CurrNode.Mode != NormalMode{

							//switchViewChangeModeMutex.Lock()
							//currActor.switchToviewChangeMode()
							//switchViewChangeModeMutex.Unlock()

							return
						}

						//CommitMapMutex.Lock()
						PrepareMapMutex.Lock()

						if commitMsg.prevMsgHash == nil {

							//log.Println("[prepare] prepareMsg.prevMsgHash")

							//switchViewChangeModeMutex.Lock()
							//currActor.switchToviewChangeMode()
							//switchViewChangeModeMutex.Unlock()

							return
						}

						if actor.BFTMsgLogs[*commitMsg.prevMsgHash] == nil {

							//log.Println("[prepare] currActor.BFTMsgLogs[*prepareMsg.prevMsgHash]")

							//switchViewChangeModeMutex.Lock()
							//currActor.switchToviewChangeMode()
							//switchViewChangeModeMutex.Unlock()

							return
						}

						//log.Println(actor.BFTMsgLogs[commitMsg.hash])

						//Save it to somewhere else for every node (actor of consensus engine)

						if actor.BFTMsgLogs[commitMsg.hash] == nil {
							actor.BFTMsgLogs[commitMsg.hash] = new(NormalMsg)
							*actor.BFTMsgLogs[commitMsg.hash] = commitMsg
						}
						actor.BFTMsgLogs[commitMsg.hash].Amount++

						actor.wg.Done()

						PrepareMapMutex.Unlock()
						//CommitMapMutex.Unlock()
					}
				}()

				actor.commitMsgTimerCh <- true
				actor.wg.Wait()

			//
			////case viewChangeMsg := <- actor.ViewChangeMsgCh:
			////
			////	currActor := actor.CurrNode.consensusEngine.BFTProcess
			////
			////	if currActor.CurrNode.Mode != ViewChangeMode{
			////		//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[viewchange] Block by viewchange mode verifier")
			////		//currActor.switchToNormalMode() //Race condition
			////		continue
			////	}
			////
			////	switch viewChangeMsg.Type {
			////	case VIEWCHANGE:
			////
			////		//log.Println("Viewchange")
			////
			////		//Timeout generate new block
			////		//currActor.wg.Add(1)
			////		go func(){
			////			//defer func() {
			////			//	currActor.wg.Done()
			////			//}()
			////
			////			select {
			////			case <-currActor.timeOutCh:
			////				if currActor.CurrNode.Mode == ViewChangeMode{
			////
			////					//log.Println("View", currActor.CurrNode.View ,"View change msg from:", viewChangeMsg.SignerID, "to:", currActor.CurrNode.index)
			////
			////					if viewChangeMsg.View < currActor.CurrNode.View {
			////						//TODO: Restart new view change mode
			////						// Send faulty node ask to become primary node msg to other nodes
			////						//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[view change] (already in viewchange mode) viewChangeMsg.View <= currActor.CurrNode.View")
			////						switchViewChangeModeMutex.Lock()
			////						currActor.switchToviewChangeMode()
			////						switchViewChangeModeMutex.Unlock()
			////						return
			////					}
			////				} else {
			////					if viewChangeMsg.View <= currActor.CurrNode.View{
			////						//TODO: Restart new view change mode
			////						// Send faulty node ask to become primary node msg to other nodes
			////						//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[view change] viewChangeMsg.View <= currActor.CurrNode.View")
			////						switchViewChangeModeMutex.Lock()
			////						currActor.switchToviewChangeMode()
			////						switchViewChangeModeMutex.Unlock()
			////						return
			////						//currActor.wg.Wait()
			////					}
			////				}
			////
			////				viewChangeMutex.Lock()
			////
			////				//Save view change msg to somewhere
			////				if currActor.ViewChangeMsgLogs[viewChangeMsg.hash] == nil {
			////					isDup := false
			////					for _, element := range currActor.ViewChangeMsgLogs{
			////						if element.Type == VIEWCHANGE && element.SignerID == viewChangeMsg.SignerID && element.View == viewChangeMsg.View{
			////							isDup = true
			////							break
			////						}
			////					}
			////
			////					if !isDup{
			////
			////						//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[view change] currActor.ViewChangeMsgLogs", currActor.ViewChangeMsgLogs)
			////						//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[view change] viewChangeMsg.hash", viewChangeMsg.hash)
			////
			////						currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = new(ViewMsg)
			////						*currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = viewChangeMsg
			////					}
			////				}
			////
			////				for _, msg := range currActor.ViewChangeMsgLogs{
			////					if msg.View == currActor.View() && msg.Type == VIEWCHANGE{
			////						saveMsgMutex.Lock()
			////						currActor.viewChangeAmount[int(currActor.View())]++
			////						saveMsgMutex.Unlock()
			////					}
			////				}
			////
			////				viewChangeMutex.Unlock()
			////
			////				timerMutex.Lock()
			////				currActor.amountMsgTimer = time.NewTimer(time.Millisecond * 100) //Race condition
			////				timerMutex.Unlock()
			////
			////				go func(){
			////					select {
			////					case <-currActor.amountMsgTimer.C: //Race condition
			////						viewChangeMutex.Lock()
			////
			////						if uint64(currActor.viewChangeAmount[int(currActor.View())]) > uint64(2*n/3){
			////
			////							if !currActor.viewChangeExpire {
			////
			////								currActor.viewChangeExpire = true
			////
			////								if currActor.isPrimaryNode(int(currActor.View())){
			////									msg := ViewMsg{
			////										hash: utils.GenerateHashV1(),
			////										Type:       NEWVIEW,
			////										View:       viewChangeMsg.View,
			////										SignerID:   currActor.CurrNode.index,
			////										Timestamp:  uint64(time.Now().Unix()),
			////										prevMsgHash: viewChangeMsg.prevMsgHash, //viewchange msg hash
			////									}
			////
			////									for _, msg := range currActor.ViewChangeMsgLogs{
			////										if msg.View == currActor.View() && msg.Type == VIEWCHANGE{
			////											msg.hashSignedMsgs = append(msg.hashSignedMsgs, msg.hash)
			////										}
			////									}
			////
			////									//Save view change msg to somewhere
			////									if currActor.ViewChangeMsgLogs[msg.hash] == nil {
			////										currActor.ViewChangeMsgLogs[msg.hash] = new(ViewMsg)
			////										*currActor.ViewChangeMsgLogs[msg.hash] = msg
			////									}
			////
			////									//Send messages to other nodes
			////									for _, element := range currActor.Validators{
			////										//Define message for sending back to primary node
			////
			////										go func(node *Node){
			////											node.consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
			////										}(element)
			////									}
			////								}
			////
			////								//TODO:
			////								// Start new view of view change mode here
			////
			////							}
			////
			////						} else {
			////							//currActor.wg.Add(1)
			////							switchViewChangeModeMutex.Lock()
			////							currActor.switchToviewChangeMode()
			////							switchViewChangeModeMutex.Unlock()
			////							return
			////							//currActor.wg.Wait()
			////						}
			////
			////						viewChangeMutex.Unlock()
			////					}
			////				}()
			////
			////			case <-currActor.viewChangeTimer.C:
			////				time.Sleep(time.Millisecond * 100)
			////
			////				//currActor.wg.Add(1)
			////				switchViewChangeModeMutex.Lock()
			////				currActor.switchToviewChangeMode()
			////				switchViewChangeModeMutex.Unlock()
			////				return
			////				//currActor.wg.Wait()
			////			}
			////		}()
			////
			////		timeOutChMutex.Lock()
			////		currActor.timeOutCh <- true
			////		timeOutChMutex.Unlock()
			////
			////	case NEWVIEW:
			////
			////		//log.Println("Newview")
			////
			////		//log.Println("View", currActor.CurrNode.View , "New view msg from:", viewChangeMsg.SignerID, "to:", currActor.CurrNode.index)
			////
			////		go func(){
			////			select {
			////			case <-currActor.timeOutCh:
			////				currActor.viewChangeExpire = false
			////
			////				if viewChangeMsg.View != currActor.CurrNode.View{
			////					//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[New view] viewChangeMsg.View <= currActor.CurrNode.View")
			////					switchViewChangeModeMutex.Lock()
			////					currActor.switchToviewChangeMode()
			////					switchViewChangeModeMutex.Unlock()
			////					return
			////				}
			////
			////				if viewChangeMsg.SignerID != currActor.CurrNode.index {
			////					for _, hash := range viewChangeMsg.hashSignedMsgs {
			////						if currActor.ViewChangeMsgLogs[hash] == nil {
			////							//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[New view] Wrong signed viewchange messages")
			////
			////							switchViewChangeModeMutex.Lock()
			////							currActor.switchToviewChangeMode()
			////							switchViewChangeModeMutex.Unlock()
			////
			////							break
			////						}
			////					}
			////				}
			////
			////				//Save view change msg to somewhere
			////				if currActor.ViewChangeMsgLogs[viewChangeMsg.hash] == nil {
			////					currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = new(ViewMsg)
			////					*currActor.ViewChangeMsgLogs[viewChangeMsg.hash] = viewChangeMsg
			////				}
			////
			////				currActor.viewChangeTimer.Stop()
			////
			////				//currActor.wg.Add(1)
			////
			////				//log.Println("Jump to switch to normal mode")
			////
			////				switchNormalModeMutex.Lock()
			////				currActor.switchToNormalMode()
			////
			////				if currActor.isPrimaryNode(int(currActor.View())){
			////					currActor.BroadcastMsgCh <- true
			////				} else {
			////					time.Sleep(time.Millisecond * 50)
			////				}
			////
			////				switchNormalModeMutex.Unlock()
			////
			////			case <-currActor.viewChangeTimer.C:
			////				//currActor.wg.Add(1)
			////				switchViewChangeModeMutex.Lock()
			////				currActor.switchToviewChangeMode()
			////				switchViewChangeModeMutex.Unlock()
			////				return
			////				//currActor.wg.Wait()
			////			}
			////		}()
			////
			////		currActor.timeOutCh <- true
			////	}

			case msgTimer := <- actor.msgTimerCh:
				actor.handleMsgTimer(msgTimer)

			//case msgTimer := <- actor.postMsgTimerCh:
			//
			//	switch msgTimer.Type {
			//	case PREPARE:
			//
			//		timerMutex.Lock()
			//		actor.prepareAmountMsgTimer.Stop()
			//		actor.prepareAmountMsgTimer = nil
			//		timerMutex.Unlock()
			//
			//	case COMMIT:
			//
			//		timerMutex.Lock()
			//		actor.commitAmountMsgTimer.Stop()
			//		actor.commitAmountMsgTimer = nil
			//		timerMutex.Unlock()
			//
			//	}

			}

		}
	}()
	return nil
}

