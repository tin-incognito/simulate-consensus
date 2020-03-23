package consensus

import (
	"fmt"
	"github.com/tin-incognito/simulate-consensus/common"
	"github.com/tin-incognito/simulate-consensus/consensus/sawtooth"
	"log"
	"time"
)

//Engine ...
type Engine struct{

	BFTProcess ConsensusHandler
	consensusName        string
	IsEnabled            int //0 > stop, 1: running
	Type string

	//curringMiningState struct {
	//	layer   string
	//	role    string
	//	chainID int
	//}


	version int

}

//Name ...
func (engine Engine) Name() (string, error){
	return engine.consensusName, nil
}

func (engine *Engine) NewEngine(config *Config) (*Engine, error){
	res := &Engine{
		version: config.Version,
		consensusName: config.Name,
	}
	return res ,nil
}

func (engine *Engine) UserLayer() (string, int) {
	//return engine.curringMiningState.layer, engine.curringMiningState.chainID
	return "", 0
}

func (engine *Engine) UserRole() (string, string, int) {
	//return s.curringMiningState.layer, s.curringMiningState.role, s.curringMiningState.chainID
	return "", "", 0
}

//
func (engine *Engine) Start(){

	//fmt.Println("CONSENSUS: Start")
	//if engine.config.Node.GetPrivateKey() != "" {
	//	keyList, err := engine.GenMiningKeyFromPrivateKey(engine.config.Node.GetPrivateKey())
	//	if err != nil {
	//		panic(err)
	//	}
	//	engine.userKeyListString = keyList
	//} else if engine.config.Node.GetMiningKeys() != "" {
	//	engine.userKeyListString = engine.config.Node.GetMiningKeys()
	//}
	//err := engine.LoadMiningKeys(engine.userKeyListString)
	//if err != nil {
	//	panic(err)
	//}

	go engine.watchCommitteeChange()
	engine.IsEnabled = 1
}

//watchCommitteeChange ...
func (engine Engine) watchCommitteeChange(){
	defer func() {
		time.AfterFunc(time.Second*3, engine.watchCommitteeChange)
	}()

	//check if enable
	if engine.IsEnabled == 0 || engine.consensusName == ""  {
		fmt.Println("CONSENSUS: enable", engine.IsEnabled, engine.consensusName == "")
		return
	}

	if engine.version == 1{
		if engine.Type == common.SawToothConsensus {
			engine.BFTProcess = sawtooth.NewActor()
		} else {
			//.....
		}
	} else {
		//....
	}


	// Start process of consensus in this code block
	if err := engine.BFTProcess.Start(); err != nil {
		//Log error in this function
		log.Println(err)
		return
	}

}