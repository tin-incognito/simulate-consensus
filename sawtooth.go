package main

import (
	"github.com/tin-incognito/simulate-consensus/common"
	"log"
	"time"
)

//Actor
type Actor struct{
	PrePrepareMsgCh chan NormalMsg
	PrepareMsgCh    chan NormalMsg
	CommitMsgCh chan NormalMsg
	FinishMsgCh chan NormalMsg
	ViewChangeMessageCh chan ViewMsg
	BroadcastMsgCh chan struct{}
	isStarted      bool
	StopCh         chan struct{}
	CurrNode 	   *Node
	Validators map[int]*Node
	chainHandler ChainHandler
	Test chan string
}

func NewActor() *Actor{

	res := &Actor{
		PrePrepareMsgCh:     make (chan NormalMsg),
		PrepareMsgCh:        make (chan NormalMsg),
		CommitMsgCh:         make (chan NormalMsg),
		FinishMsgCh:    make (chan NormalMsg),
		ViewChangeMessageCh: make (chan ViewMsg),
		BroadcastMsgCh:      make (chan struct{}),
		isStarted:           true,
		CurrNode:            nil,
		Validators: make(map[int]*Node),
		StopCh: make(chan struct{}),
		chainHandler: &Chain{},
		Test: make(chan string),
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

		select {
		case test := <- actor.Test:
			log.Println(test)
		case <- actor.StopCh:
			log.Println(0)
			return
		case broadcastMsg := <- actor.BroadcastMsgCh:
			log.Println(1)
			log.Println("broadcastMsg:", broadcastMsg)
			block, err := actor.chainHandler.CreateBlock()
			if err != nil {
				log.Println(err)
				return
			}

			log.Println("block:", block)

			msg := NormalMsg{
				Type:      PREPREPARE,
				View:      actor.chainHandler.View(),
				SeqNum:    actor.chainHandler.SeqNumber(),
				SignerID:  actor.CurrNode.index,
				Timestamp: uint64(time.Now().Unix()),
				BlockID:   &block.Index,
			}

			log.Println("msg:", msg)

			for _, member := range actor.Validators{
				member.consensusEngine.BFTProcess.PrePrepareMsgCh <- msg
			}

		case prePrepareMsg := <- actor.PrePrepareMsgCh:
			log.Println(2)
			log.Println(prePrepareMsg)

		case prepareMsg := <- actor.PrepareMsgCh:
			log.Println(3)
			log.Println(prepareMsg)

		case commitMsgCh := <- actor.CommitMsgCh:
			log.Println(3)
			log.Println(commitMsgCh)

		case finishMsgCh := <- actor.FinishMsgCh:
			log.Println(3)
			log.Println(finishMsgCh)

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

