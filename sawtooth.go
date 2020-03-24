package main

import (
	"github.com/tin-incognito/simulate-consensus/common"
	"log"
)

//Actor
type Actor struct{
	BFTMessageCh     chan PbftMsg
	ProposeMessageCh chan PbftMsg
	VoteMessageCh    chan PbftMsg
	ViewChangeMessageCh chan PbftMsg
	BroadcastMsgCh chan PbftMsg
	isStarted      bool
	StopCh         chan struct{}
	CurrNode 	   *Node
	Validators map[int]*Node

}

func NewActor() *Actor{
	res := &Actor{
		Validators: make(map[int]*Node),
		StopCh: make(chan struct{}),
		ProposeMessageCh: make(chan PbftMsg),
		VoteMessageCh: make(chan PbftMsg),
	}

	return res
}

//Name return name of actor to user
func (actor Actor) Name() (string,error){
	return common.SawToothConsensus, nil
}

//Start ...
func (actor Actor) start() error{

	// Summary consensus of Sawtooth
	//
	//TODO: Implement sawtooth consensus:
	// - Define message send in network and encode it with protoc
	// -


	actor.isStarted = true

	go func(){
		//ticker := time.Tick(300 * time.Millisecond)

		//log.Println("action")

		select {
		case <- actor.StopCh:
			log.Println(0)
			return
		case broadcastMsg := <- actor.BroadcastMsgCh:
			log.Println(1)
			log.Println("broadcastMsg:", broadcastMsg)

		case proposeMsg := <- actor.ProposeMessageCh:
			log.Println(2)
			log.Println(proposeMsg)

		case voteMsg := <- actor.VoteMessageCh:
			log.Println(3)
			log.Println(voteMsg)

		case viewChangeMsg := <- actor.ViewChangeMessageCh:
			log.Println(4)
			log.Println(viewChangeMsg)

		//case <-ticker:
		//	log.Println(5)
		//	return
		}
	}()


	return nil
}

func (actor *Actor) initValidators(m map[int]*Node) {
	for i, element := range m{
		if actor.CurrNode.index != i{
			actor.Validators[i] = element
		}
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
	err := engine.BFTProcess.start()
	return err
}

