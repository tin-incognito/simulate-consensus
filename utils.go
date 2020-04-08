package main

import (
	"github.com/tin-incognito/simulate-consensus/utils"
	"log"
	"time"
)

func (actor *Actor) updateAllViewChangeMode(){
	for _, element := range actor.Validators{
		element.Mode = ViewChangeMode
	}
}

func (actor *Actor) updateAllNormalMode(){
	for _, element := range actor.Validators{
		element.Mode = NormalMode
	}
}

//switchToNormalMode ...
func (actor *Actor) switchToNormalMode(){

	log.Println( "View:", actor.CurrNode.View, "Node", actor.CurrNode.index, "switch to normal mode")

	currActor := actor.CurrNode.consensusEngine.BFTProcess

	//defer func(){
	//	currActor.wg.Done()
	//}()

	switchMutex.Lock()

	//log.Println("node ", actor.CurrNode.index, "switch back to normal mode")

	currActor.CurrNode.Mode = NormalMode
	currActor.ProposalNode = currActor.Validators[currActor.calculatePrimaryNode(int(currActor.View()))]

	if currActor.isPrimaryNode(int(currActor.View())){
		currActor.CurrNode.IsProposer = true //Race condition
	}

	currActor.ProposalNode = currActor.Validators[currActor.calculatePrimaryNode(int(currActor.View()))]

	log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[switch mode] switch to normal mode:", currActor.CurrNode)

	switchMutex.Unlock()
}

//switchToViewChangeMode ...
func (actor *Actor) switchToviewChangeMode(){

	currActor := actor.CurrNode.consensusEngine.BFTProcess

	log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "switch to view change mode")

	switchMutex.Lock()

	currActor.viewChangeTimer = time.NewTimer(time.Millisecond * 5000) /// Race condition

	//defer func(){
	//	currActor.wg.Done()
	//}()

	if currActor.CurrNode.IsProposer{
		currActor.CurrNode.IsProposer = false
	}

	currActor.ViewChanging(currActor.CurrNode.View + 1)

	msg := ViewMsg {
		hash: utils.GenerateHashV1(),
		Type:       VIEWCHANGE,
		View:       currActor.CurrNode.View,
		SignerID:   currActor.CurrNode.index,
		Timestamp:  uint64(time.Now().Unix()),
		prevMsgHash: nil,
	}

	//Save view change msg to somewhere
	if currActor.ViewChangeMsgLogs[msg.hash] == nil {
		currActor.ViewChangeMsgLogs[msg.hash] = new(ViewMsg) //Race condition
		*currActor.ViewChangeMsgLogs[msg.hash] = msg
	}

	//actor.sendMsgMutex.Lock()
	//Send messages to other nodes
	for _, element := range currActor.Validators{
		go func(node *Node){
			node.consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
		}(element)
	}

	switchMutex.Unlock()
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
	currActor := actor.CurrNode.consensusEngine.BFTProcess
	currActor.ProposalNode = currActor.Validators[index]
}

//updateNormalMode ...
func (actor *Actor) updateNormalMode(view uint64) {
	currActor := actor.CurrNode.consensusEngine.BFTProcess
	modeMutex.Lock()
	currActor.CurrNode.Mode = NormalMode
	modeMutex.Unlock()
}

//ViewChanging ...
func (actor *Actor) ViewChanging(v uint64) error{
	currActor := actor.CurrNode.consensusEngine.BFTProcess

	modeMutex.Lock()
	currActor.CurrNode.Mode = ViewChangeMode //Race condition
	currActor.CurrNode.View = v
	err := currActor.chainHandler.IncreaseView()
	if err != nil {
		return err
	}
	modeMutex.Unlock()
	return nil
}

func (actor *Actor) initValidators(m map[int]*Node) {
	currActor := actor.CurrNode.consensusEngine.BFTProcess

	for i, element := range m{
		currActor.Validators[i] = element
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
