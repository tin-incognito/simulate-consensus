package main

import (
	"log"
	"time"
)

var globalCh chan bool
var globalTimer *time.Timer

func startGlobalTimer(){
	globalTimer = time.NewTimer(time.Millisecond * 5000)
	go func(){
		for {
			select {
			case <- globalCh:
				log.Println("Active global channel")
				globalTimer.Reset(time.Millisecond * 5000)
			case <- globalTimer.C:

				for _, element := range nodes{

					element.IsProposer = false
					element.consensusEngine.BFTProcess.ProposalNode = nil

					if element.consensusEngine.BFTProcess.ProposalNode == nil {
						element.consensusEngine.BFTProcess.ProposalNode = nodes[0]
					} else {
						*element.consensusEngine.BFTProcess.ProposalNode = *nodes[0]
					}
				}

				nodes[0].IsProposer = true
				nodes[0].consensusEngine.BFTProcess.BroadcastMsgCh <- true
			}
		}
	}()
}
