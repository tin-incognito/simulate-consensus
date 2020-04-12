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

				modeMutex.Lock()

				if actor.CurrNode.Mode != NormalMode{
					//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[broadcast] Block by normal mode verifier")
					//currActor.switchToviewChangeMode()
					continue
				}

				modeMutex.Unlock()

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
							//log.Println("Send from:", actor.CurrNode.index, "to:", node.index)
							node.consensusEngine.BFTProcess.PrePrepareMsgCh <- msg
						}(member)
					}
				}

			case prePrepareMsg := <- actor.PrePrepareMsgCh:

				//log.Println("pre prepare msg:", prePrepareMsg)

				modeMutex.Lock()

				if actor.CurrNode.Mode != NormalMode {
					//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[pre prepare] Block by normal mode verifier")

					//TODO:
					// How about switching to view change mode

					//switchViewChangeModeMutex.Lock()
					//currActor.switchToviewChangeMode()
					//switchViewChangeModeMutex.Unlock()

					continue
				}

				modeMutex.Unlock()

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

				PrepareMapMutex.Lock()

				if actor.prePrepareMsg[int(actor.CurrNode.View)] != nil && getEnv("ENV", "prod") == Production{
					//switchViewChangeModeMutex.Lock()
					//currActor.switchToviewChangeMode()
					//switchViewChangeModeMutex.Unlock()
					continue
				}

				if actor.BFTMsgLogs[prePrepareMsg.hash] == nil{
					actor.BFTMsgLogs[prePrepareMsg.hash] = new(NormalMsg)
					*actor.BFTMsgLogs[prePrepareMsg.hash] = prePrepareMsg
				} else {
					actor.BFTMsgLogs[prePrepareMsg.hash].Amount++
				}

				actor.prePrepareMsg[int(actor.CurrNode.View)] = new(NormalMsg)
				*actor.prePrepareMsg[int(actor.CurrNode.View)] = prePrepareMsg

				PrepareMapMutex.Unlock()

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
					//defer func() {
					//	actor.commitTimer.Stop()
					//	actor.wg.Done()
					//}()

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

							//log.Println(msg)

							for _, member := range actor.Validators{

								go func(node *Node){
									//log.Println("Send message from:", actor.CurrNode, "to:", node.index)
									node.consensusEngine.BFTProcess.PrepareMsgCh <- msg
								}(member)
							}
						}

						actor.commitTimer.Stop()
						actor.wg.Done()

						prepareMutex.Unlock()

						return

					case <-actor.commitTimer.C:

						//switchViewChangeModeMutex.Lock()
						//currActor.switchToviewChangeMode()
						//switchViewChangeModeMutex.Unlock()

						actor.commitTimer.Stop()
						actor.wg.Done()

						return
					}
				}()

				actor.timeOutCh <- true
				actor.wg.Wait()

			case prepareMsg := <- actor.PrepareMsgCh:

				//log.Println("prepareMsg:", prepareMsg)

				// This is still preparing phase

				modeMutex.Lock()

				if actor.CurrNode.Mode != NormalMode{
					//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[prepare] Block by normal mode verifier")

					//switchViewChangeModeMutex.Lock()
					//currActor.switchToviewChangeMode()
					//switchViewChangeModeMutex.Unlock()

					return
				}

				modeMutex.Unlock()

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

				modeMutex.Lock()
				if actor.CurrNode.Mode != NormalMode{

					//switchViewChangeModeMutex.Lock()
					//currActor.switchToviewChangeMode()
					//switchViewChangeModeMutex.Unlock()

					return
				}
				modeMutex.Unlock()

				msgTimerMutex.Lock()
				if actor.commitAmountMsgTimer == nil{
					actor.commitAmountMsgTimer = time.NewTimer(time.Millisecond * 100)

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


			case viewChangeMsg := <- actor.ViewChangeMsgCh:

				//modeMutex.Lock()

				//log.Println("[outside] actor.CurrNode:", actor.CurrNode)

				//if actor.CurrNode.Mode != ViewChangeMode{
				//	//currActor.switchToNormalMode() //Race condition
				//	log.Println(1)
				//	continue
				//}

				//modeMutex.Unlock()

				switch viewChangeMsg.Type {

				case VIEWCHANGE:

					msgTimerMutex.Lock()
					if actor.viewChangeAmountMsgTimer == nil {
						actor.viewChangeAmountMsgTimer = time.NewTimer(time.Millisecond * 100)

						msgTimer := MsgTimer{VIEWCHANGE}
						go func() {
							actor.msgTimerCh <- msgTimer
						}()

					}
					msgTimerMutex.Unlock()

					actor.wg.Add(1)
					go func() {
						select {
						case <- actor.viewChangeMsgTimerCh:

							modeMutex.Lock()

							if actor.CurrNode.Mode == ViewChangeMode{

								if viewChangeMsg.View < actor.CurrNode.View {
									//switchViewChangeModeMutex.Lock()
									//currActor.switchToviewChangeMode()
									//switchViewChangeModeMutex.Unlock()
									return
								}

							} else {
								if viewChangeMsg.View <= actor.CurrNode.View{
									//switchViewChangeModeMutex.Lock()
									//currActor.switchToviewChangeMode()
									//switchViewChangeModeMutex.Unlock()
									return
								}
							}

							//Try to figure other way
							//Save view change msg to somewhere
							if actor.ViewChangeMsgLogs[viewChangeMsg.hash] == nil {
								isDup := false
								for _, element := range actor.ViewChangeMsgLogs {
									if element.Type == VIEWCHANGE && element.SignerID == viewChangeMsg.SignerID && element.View == viewChangeMsg.View {
										isDup = true
										break
									}
								}

								if !isDup{
									actor.ViewChangeMsgLogs[viewChangeMsg.hash] = new(ViewMsg)
									*actor.ViewChangeMsgLogs[viewChangeMsg.hash] = viewChangeMsg
								}
								actor.viewChangeAmount[int(actor.View())]++
							}

							for _, msg := range actor.ViewChangeMsgLogs{
								if msg.View == actor.View() && msg.Type == VIEWCHANGE{
									actor.viewChangeAmount[int(actor.View())]++
								}
							}

							actor.wg.Done()
							modeMutex.Unlock()
						}
					}()

					actor.viewChangeMsgTimerCh <- true
					actor.wg.Wait()

				case NEWVIEW:

					actor.wg.Add(1)
					go func() {
						select {
						case <- actor.newViewMsgTimerCh:

							//log.Println(1)

							if viewChangeMsg.View != actor.CurrNode.View{
								//ModeMapMutex.Lock()
								//actor.switchToviewChangeMode()
								//ModeMapMutex.Unlock()
								return
							}

							if viewChangeMsg.SignerID != actor.CurrNode.index {
								for _, hash := range viewChangeMsg.hashSignedMsgs{
									if actor.ViewChangeMsgLogs[hash] == nil {
										//ModeMapMutex.Lock()
										//actor.switchToviewChangeMode()
										//ModeMapMutex.Unlock()
										break
									}
								}
							}
							//Save view change msg to somewhere
							if actor.ViewChangeMsgLogs[viewChangeMsg.hash] == nil {
								actor.ViewChangeMsgLogs[viewChangeMsg.hash] = new(ViewMsg)
								*actor.ViewChangeMsgLogs[viewChangeMsg.hash] = viewChangeMsg
							}

							timerMutex.Lock()
							if actor.viewChangeTimer != nil{
								actor.viewChangeTimer.Stop()
							}
							timerMutex.Unlock()

							actor.switchToNormalMode()
							if actor.isPrimaryNode(int(actor.View())){
								actor.BroadcastMsgCh <- true
							}

							actor.wg.Done()
						}
					}()
				}

				//log.Println(0)

				go func() {
					//actor.stuckCh <- "Test"
					actor.newViewMsgTimerCh <- true
				}()
				actor.wg.Wait()

			case msgTimer := <- actor.msgTimerCh:
				actor.handleMsgTimer(msgTimer)

			case <- actor.viewChangeTimerCh:
				go func() {
					select {
					case <- actor.viewChangeTimer.C:
						actor.switchToviewChangeMode()
						return
					}
				}()
			}

		}
	}()
	return nil
}

