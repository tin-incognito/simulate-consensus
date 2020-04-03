package main

import (
	"github.com/tin-incognito/simulate-consensus/utils"
	"log"
	"math/rand"
	"time"
)

func (actor *Actor) isPrimaryNode(view int) bool{
	res := view % len(actor.Validators)
	return res == actor.CurrNode.index
}

//switchToViewChangeMode ...
func (actor *Actor) swtichToViewChangeMode() error{

	err := actor.CurrNode.updateAfterNormalMode()
	if err != nil{
		log.Println(err)
		return err
	}

	if actor.CurrNode.IsProposer{
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

		//time.Sleep(time.Millisecond * 200)
		actor.Validators[rdNum].consensusEngine.BFTProcess.ViewChangeMsgCh <- viewChangeMsg
	}

	return nil
}
