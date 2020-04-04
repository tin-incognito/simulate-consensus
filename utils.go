package main

import (
	"github.com/tin-incognito/simulate-consensus/utils"
	"log"
	"math/rand"
	"time"
)

//switchToNormalMode ...
func (actor *Actor) switchToNormalMode(){
	actor.CurrNode.Mode = NormalMode
}

//switchToViewChangeMode ...
func (actor *Actor) switchToviewChangeMode(){

	actor.ViewChanging(actor.CurrNode.View + 1)

	msg := ViewMsg{
		hash: utils.GenerateHashV1(),
		Type:       VIEWCHANGE,
		View:       actor.CurrNode.View,
		SignerID:   actor.CurrNode.index,
		Timestamp:  uint64(time.Now().Unix()),
		prevMsgHash: nil,
		owner: actor.CurrNode.index,
	}

	//Save view change msg to somewhere
	if actor.ViewChangeMsgLogs[msg.hash] == nil {
		actor.ViewChangeMsgLogs[msg.hash] = new(ViewMsg)
		*actor.ViewChangeMsgLogs[msg.hash] = msg
	}

	actor.sendMsgMutex.Lock()
	//Send messages to other nodes
	for i, _ := range actor.Validators{
		//log.Println("Send from node:", i)
		//element.consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
		actor.Validators[i].consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
	}
	actor.sendMsgMutex.Unlock()
}

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

func (actor *Actor) View() uint64{
	return actor.CurrNode.View
}
