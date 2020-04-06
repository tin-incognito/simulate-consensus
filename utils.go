package main

import (
	"github.com/tin-incognito/simulate-consensus/utils"
	"log"
	"time"
)

//switchToNormalMode ...
func (actor *Actor) switchToNormalMode(){

	defer func(){
		actor.wg.Done()
	}()

	actor.switchMutex.Lock()

	log.Println("node ", actor.CurrNode.index, "switch back to normal mode")

	actor.CurrNode.Mode = NormalMode
	actor.ProposalNode = actor.Validators[actor.calculatePrimaryNode(int(actor.View()))]

	if actor.isPrimaryNode(int(actor.View())){
		actor.CurrNode.IsProposer = true
	}

	actor.switchMutex.Unlock()
}

//switchToViewChangeMode ...
func (actor *Actor) switchToviewChangeMode(){

	actor.switchMutex.Lock()

	actor.viewChangeTimer = time.NewTimer(time.Millisecond * 5000)

	defer func(){
		actor.wg.Done()
	}()

	if actor.CurrNode.IsProposer{
		actor.CurrNode.IsProposer = false
	}

	actor.ViewChanging(actor.CurrNode.View + 1)

	msg := ViewMsg {
		hash: utils.GenerateHashV1(),
		Type:       VIEWCHANGE,
		View:       actor.CurrNode.View,
		SignerID:   actor.CurrNode.index,
		Timestamp:  uint64(time.Now().Unix()),
		prevMsgHash: nil,
	}

	//Save view change msg to somewhere
	if actor.ViewChangeMsgLogs[msg.hash] == nil {
		actor.ViewChangeMsgLogs[msg.hash] = new(ViewMsg)
		*actor.ViewChangeMsgLogs[msg.hash] = msg
	}

	//actor.sendMsgMutex.Lock()
	//Send messages to other nodes
	for _, element := range actor.Validators{
		element.consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
	}

	actor.switchMutex.Unlock()
}

//calculatePrimaryNode ...
func (actor *Actor) calculatePrimaryNode(view int) int{
	return view % len(actor.Validators)
}

func (actor *Actor) isPrimaryNode(view int) bool{
	res := view % len(actor.Validators)
	return res == actor.CurrNode.index
}

func (actor *Actor) updateProposalNode(index int) {
	actor.ProposalNode = actor.Validators[index]
}

//updateNormalMode ...
func (actor *Actor) updateNormalMode(view uint64) {
	actor.modeMutex.Lock()
	actor.CurrNode.Mode = NormalMode
	actor.modeMutex.Unlock()
}

//ViewChanging ...
func (actor *Actor) ViewChanging(v uint64) error{
	actor.modeMutex.Lock()
	actor.CurrNode.Mode = ViewChangeMode
	actor.CurrNode.View = v
	err := actor.chainHandler.IncreaseView()
	if err != nil {
		return err
	}
	actor.modeMutex.Unlock()
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
