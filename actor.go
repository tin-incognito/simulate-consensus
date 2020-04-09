package main

import (
	"github.com/tin-incognito/simulate-consensus/common"
	"github.com/tin-incognito/simulate-consensus/utils"
	"log"
	"sync"
	"time"
)

//Actor
type Actor struct{
	PrePrepareMsgCh chan NormalMsg
	PrepareMsgCh    chan NormalMsg
	CommitMsgCh chan NormalMsg
	ViewChangeMsgCh chan ViewMsg
	BroadcastMsgCh chan bool
	isStarted      bool
	StopCh         chan struct{}
	CurrNode 	   *Node
	ProposalNode *Node
	Validators map[int]*Node
	chainHandler ChainHandler
	BFTMsgLogs map[string]*NormalMsg
	ViewChangeMsgLogs map[string]*ViewMsg
	timeOutCh chan bool
	errCh chan error
	wg sync.WaitGroup
	postAmountMsgTimerCh chan bool
	postMsgTimerCh chan MsgTimer
	prepareAmountMsgTimer *time.Timer
	commitAmountMsgTimer *time.Timer
	idleTimer *time.Timer
	blockPublishTimer *time.Timer
	commitTimer *time.Timer
	viewChangeTimer *time.Timer
	newBlock *Block
	prePrepareMsg map[int]*NormalMsg
	viewChangeExpire bool
	viewChangeAmount map[int]int
	phaseStatus string
	msgTimerCh chan MsgTimer
	stuckCh chan string
}

func NewActor() *Actor{

	res := &Actor{
		PrePrepareMsgCh:     make (chan NormalMsg, 100),
		PrepareMsgCh:        make (chan NormalMsg, 100),
		CommitMsgCh:         make (chan NormalMsg, 100),
		ViewChangeMsgCh: make (chan ViewMsg, 100),
		BroadcastMsgCh:      make (chan bool, 100),
		isStarted:           true,
		CurrNode:            nil,
		ProposalNode: nil,
		Validators: make(map[int]*Node),
		StopCh: make(chan struct{}, 20),
		chainHandler: &Chain{},
		BFTMsgLogs: make(map[string]*NormalMsg),
		ViewChangeMsgLogs: make(map[string]*ViewMsg),
		wg:                  sync.WaitGroup{},
		timeOutCh: make(chan bool, 100),
		errCh: make(chan error, 100),
		viewChangeAmount: make(map[int]int),
		msgTimerCh: make (chan MsgTimer),
		postAmountMsgTimerCh: make (chan bool),
		prePrepareMsg: make (map[int]*NormalMsg),
		stuckCh: make (chan string),
	}

	return res
}

//Name return name of actor to user
func (actor Actor) Name() (string,error){
	return common.SawToothConsensus, nil
}


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

	//log.Println( "View:", actor.CurrNode.View, "Node", actor.CurrNode.index, "switch to normal mode")

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

	//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index)

	switchMutex.Unlock()
}

//switchToViewChangeMode ...
func (actor *Actor) switchToviewChangeMode(){

	currActor := actor.CurrNode.consensusEngine.BFTProcess

	//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "switch to view change mode")

	switchMutex.Lock()

	timerMutex.Lock()
	currActor.viewChangeTimer = time.NewTimer(time.Millisecond * 5000) /// Race condition
	timerMutex.Unlock()

	//defer func(){
	//	currActor.wg.Done()
	//}()

	if currActor.CurrNode.IsProposer{
		currActor.CurrNode.IsProposer = false
	}

	viewChangingMutex.Lock()
	currActor.ViewChanging(currActor.CurrNode.View + 1)
	viewChangingMutex.Unlock()

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
	modeMutex.Lock()
	currActor := actor.CurrNode.consensusEngine.BFTProcess
	currActor.CurrNode.Mode = ViewChangeMode //Race condition
	currActor.CurrNode.View = v

	//err := currActor.chainHandler.IncreaseView()

	err := currActor.chainHandler.setView(v)

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

//handleMsgTimer ...
func (actor *Actor) handleMsgTimer(msgTimer MsgTimer){

	log.Println("actor.prepareAmountMsgTimer:", actor.prepareAmountMsgTimer)

	switch msgTimer.Type {
	case PREPARE:

		go func(){
			select {
			case <- actor.prepareAmountMsgTimer.C:
				currActor := actor.CurrNode.consensusEngine.BFTProcess

				log.Println(1)

				prepareMutex.Lock()

				prePrepareMsg := currActor.prePrepareMsg[int(currActor.CurrNode.index)]

				for _, msg := range currActor.BFTMsgLogs{
					if msg.prevMsgHash != nil && *msg.prevMsgHash == prePrepareMsg.hash && msg.Type == PREPARE {
						currActor.BFTMsgLogs[prePrepareMsg.hash].Amount++
					}
				}

				if uint64(currActor.BFTMsgLogs[prePrepareMsg.hash].Amount) <= uint64(2*n/3) {

					//switchViewChangeModeMutex.Lock()
					//currActor.switchToviewChangeMode()
					//switchViewChangeModeMutex.Unlock()

					return
				}

				if !currActor.BFTMsgLogs[prePrepareMsg.hash].prepareExpire{
					//Move to committing phase
					msg := NormalMsg{
						hash: 	   utils.GenerateHashV1(),
						Type:      COMMIT,
						View:      currActor.chainHandler.View(),
						SeqNum:    currActor.chainHandler.SeqNumber(),
						SignerID:  currActor.CurrNode.index,
						Timestamp: uint64(time.Now().Unix()),
						BlockID:   prePrepareMsg.BlockID,
						block: prePrepareMsg.block,
						prevMsgHash: &prePrepareMsg.hash,
					}

					for _, member := range currActor.Validators{
						go func(node *Node){
							log.Println("Send to commit channel")
							node.consensusEngine.BFTProcess.CommitMsgCh <- msg
						}(member)
					}
					currActor.BFTMsgLogs[prePrepareMsg.hash].prepareExpire = true
				}

				currActor.postAmountMsgTimerCh <- true
				prepareMutex.Unlock()

			}
		}()

	case COMMIT:

	case VIEWCHANGE:

	}
}