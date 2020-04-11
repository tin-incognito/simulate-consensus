package main

import (
	"github.com/tin-incognito/simulate-consensus/common"
	"github.com/tin-incognito/simulate-consensus/utils"
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
	prepareMsgTimerCh chan bool
	prepareTimeOutCh chan bool
	commitMsgTimerCh chan bool
	commitTimeOutCh chan bool
	viewChangeMsgTimerCh chan bool
	modeMutex sync.Mutex
}

func NewActor() *Actor{

	res := &Actor{
		postMsgTimerCh: make (chan MsgTimer),
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
		prePrepareMsg: make (map[int]*NormalMsg),
		stuckCh: make (chan string),
		prepareMsgTimerCh: make (chan bool),
		commitMsgTimerCh: make (chan bool),
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

	//switchMutex.Lock()

	actor.CurrNode.Mode = NormalMode
	actor.CurrNode.PrimaryNode = actor.Validators[actor.calculatePrimaryNode(int(actor.View()))]

	if actor.isPrimaryNode(int(actor.View())){
		actor.CurrNode.IsProposer = true //Race condition
	}

	actor.CurrNode.PrimaryNode = actor.Validators[actor.calculatePrimaryNode(int(actor.View()))]

	//switchMutex.Unlock()
}

//switchToViewChangeMode ...
func (actor *Actor) switchToviewChangeMode(){

	timerMutex.Lock()
	if actor.viewChangeTimer == nil{
		actor.viewChangeTimer = time.NewTimer(time.Millisecond * 5000) /// Race condition
	}
	timerMutex.Unlock()

	if actor.CurrNode.IsProposer{
		actor.CurrNode.IsProposer = false
	}

	modeMutex.Lock()

	actor.ViewChanging(actor.CurrNode.View + 1)

	//modeMutex.Lock()

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
		actor.ViewChangeMsgLogs[msg.hash] = new(ViewMsg) //Race condition
		*actor.ViewChangeMsgLogs[msg.hash] = msg
	}

	//log.Println(1)

	//Send messages to other nodes
	for _, element := range actor.Validators{
		go func(node *Node){
			node.consensusEngine.BFTProcess.ViewChangeMsgCh <- msg
		}(element)
	}

	modeMutex.Unlock()

	//modeMutex.Unlock()
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
	//currActor := actor.CurrNode.consensusEngine.BFTProcess
	//actor.ProposalNode = currActor.Validators[index]

	actor.CurrNode.PrimaryNode = actor.Validators[index]

}

//updateNormalMode ...
func (actor *Actor) updateNormalMode(view uint64) {
	//modeMutex.Lock()
	actor.CurrNode.Mode = NormalMode
	//modeMutex.Unlock()
}

//ViewChanging ...
func (actor *Actor) ViewChanging(v uint64) error{
	//modeMutex.Lock()
	//currActor := actor.CurrNode.consensusEngine.BFTProcess

	actor.CurrNode.Mode = ViewChangeMode //Race condition
	actor.CurrNode.View = v

	//err := currActor.chainHandler.IncreaseView()

	err := actor.chainHandler.setView(v)

	if err != nil {
		return err
	}
	//modeMutex.Unlock()
	return nil
}

func (actor *Actor) initValidators(m map[int]*Node) {
	//currActor := actor.CurrNode.consensusEngine.BFTProcess

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

////waitForTimeOut ...
//func (actor *Actor) waitForTimeOut(){
//	go func(){
//		select {
//		case actor:
//
//		}
//	}()
//}

//handleMsgTimer ...
func (actor *Actor) handleMsgTimer(msgTimer MsgTimer){

	switch msgTimer.Type {
	case PREPARE:

		go func(){

			select {
			case <- actor.prepareAmountMsgTimer.C:

				PrepareMapMutex.Lock()

				modeMutex.Lock()
				if actor.CurrNode.Mode != NormalMode{
					//log.Println("View", currActor.CurrNode.View, "Node", currActor.CurrNode.index, "[prepare] Block by normal mode verifier")

					//switchViewChangeModeMutex.Lock()
					//currActor.switchToviewChangeMode()
					//switchViewChangeModeMutex.Unlock()

					return
				}
				modeMutex.Unlock()

				prePrepareMsg := actor.prePrepareMsg[int(actor.CurrNode.View)]

				if !actor.BFTMsgLogs[prePrepareMsg.hash].prepareExpire{

					actor.BFTMsgLogs[prePrepareMsg.hash].prepareExpire = true

					prePrepareMsg := actor.prePrepareMsg[int(actor.CurrNode.View)]

					amount := 0

					for _, msg := range actor.BFTMsgLogs {
						if msg.prevMsgHash != nil && *msg.prevMsgHash == prePrepareMsg.hash && msg.Type == PREPARE && msg.View == actor.CurrNode.View {
							amount++
						}
					}

					//Need to refactor with timeout for messages

					if uint64(amount) <= uint64(2*n/3){
						//TODO:
						// Swtich to view change mode
						return
					}

					//Move to committing phase

					msg := NormalMsg{
						hash: 	   utils.GenerateHashV1(),
						Type:      COMMIT,
						View:      actor.chainHandler.View(),
						SeqNum:    actor.chainHandler.SeqNumber(),
						SignerID:  actor.CurrNode.index,
						Timestamp: uint64(time.Now().Unix()),
						BlockID:   prePrepareMsg.BlockID,
						block: prePrepareMsg.block,
						prevMsgHash: &prePrepareMsg.hash,
					}

					for _, member := range actor.Validators{
						go func(node *Node){
							node.consensusEngine.BFTProcess.CommitMsgCh <- msg
						}(member)
					}

					actor.prepareAmountMsgTimer = nil
				}

				PrepareMapMutex.Unlock()
			}
		}()

	case COMMIT:

		go func(){

			select {
			case <- actor.commitAmountMsgTimer.C:

				modeMutex.Lock()
				if actor.CurrNode.Mode != NormalMode{

					//switchViewChangeModeMutex.Lock()
					//currActor.switchToviewChangeMode()
					//switchViewChangeModeMutex.Unlock()

					return
				}
				modeMutex.Unlock()

				PrepareMapMutex.Lock()

				prePrepareMsg := actor.prePrepareMsg[int(actor.CurrNode.View)]

				if !actor.BFTMsgLogs[prePrepareMsg.hash].commitExpire {

					actor.BFTMsgLogs[prePrepareMsg.hash].commitExpire = true

					prePrepareMsg := actor.prePrepareMsg[int(actor.CurrNode.View)]

					amount := 0

					for _, msg := range actor.BFTMsgLogs {
						if msg.prevMsgHash != nil && *msg.prevMsgHash == prePrepareMsg.hash && msg.Type == COMMIT && msg.View == actor.CurrNode.View {
							//actor.BFTMsgLogs[prePrepareMsg.hash].Amount++
							amount++
						}
					}

					//Need to refactor with timeout for messages

					if uint64(amount) <= uint64(2*n/3){
						//TODO:
						// Switch to view change mode
						return
					}

					//Update current chain
					check, err := actor.chainHandler.ValidateBlock(prePrepareMsg.block)
					if err != nil || !check {
						//actor.switchToviewChangeMode()
						return
					}

					check, err = actor.chainHandler.InsertBlock(prePrepareMsg.block)
					if err != nil || !check {
						//switchViewChangeModeMutex.Lock()
						//actor.switchToviewChangeMode()
						//switchViewChangeModeMutex.Unlock()
						return
					}

					//Increase sequence number
					err = actor.chainHandler.IncreaseSeqNum()
					if err != nil {
						//switchViewChangeModeMutex.Lock()
						//actor.switchToviewChangeMode()
						//switchViewChangeModeMutex.Unlock()
						return
					}

					if actor.CurrNode.IsProposer { //Race condition

						actor.chainHandler.print()

						if getEnv("ENV", "prod") == "test"{
							//TEST SIMULATE NORMAL MODE
							time.Sleep(time.Millisecond * 500)
							actor.BroadcastMsgCh <- true
							///
						}
					}

					actor.commitAmountMsgTimer = nil

					//After normal mode
					ModeMapMutex.Lock()
					actor.CurrNode.updateAfterNormalMode()
					actor.switchToviewChangeMode()
					ModeMapMutex.Unlock()
				}

				PrepareMapMutex.Unlock()
			}
		}()

	case VIEWCHANGE:

	}
}