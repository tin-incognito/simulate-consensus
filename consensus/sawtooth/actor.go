package sawtooth

import (
	"github.com/tin-incognito/simulate-consensus/common"
	"log"
	"time"
)

//Actor
type Actor struct{
	BFTMessageCh     chan MessageBFT
	ProposeMessageCh chan BFTPropose
	VoteMessageCh    chan BFTVote
	ViewChangeMessageCh chan MessageBFT
	isStarted      bool
	StopCh         chan struct{}
}

func NewActor() *Actor{
	res := &Actor{}
	
	return res
}

//Name return name of actor to user
func (actor Actor) Name() (string,error){
	return common.SawToothConsensus, nil
}

//Start ...
func (actor Actor) Start() error{

	// Summary consensus of Sawtooth
	//
	//TODO: Implement sawtooth consensus:
	// - Define message send in network and encode it with protoc
	// - 


	actor.isStarted = true
	actor.StopCh = make(chan struct{})
	actor.ProposeMessageCh = make(chan BFTPropose)
	actor.VoteMessageCh = make(chan BFTVote)

	go func(){
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		log.Println("action")

		select {
		case <- actor.StopCh:
			return

		case proposeMsg := <- actor.ProposeMessageCh:

			return
		case voteMsg := <- actor.VoteMessageCh:

			return
		case viewChangeMsg := <- actor.ViewChangeMessageCh:

			return
		case <-ticker:

		}


	}()


	return nil
}