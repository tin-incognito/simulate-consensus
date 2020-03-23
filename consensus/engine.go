package consensus

//Engine ...
type Engine struct{
	consensusName        string
	IsEnabled            int //0 > stop, 1: running

	curringMiningState struct {
		layer   string
		role    string
		//chainID int
	}

	version int

}

//Name ...
func (engine Engine) Name() (string, error){
	return engine.consensusName, nil
}

//func (engine *Engine) NewEngine() (*Engine, error){
//	res := &Engine{
//		version:
//	}
//}

func (engine *Engine) UserLayer() (string, int) {
	//return engine.curringMiningState.layer, engine.curringMiningState.chainID
	return "", 0
}

func (engine *Engine) UserRole() (string, string, int) {
	//return s.curringMiningState.layer, s.curringMiningState.role, s.curringMiningState.chainID
	return "", "", 0
}

//func (engine *Engine) Init(config *EngineConfig) {
//
//	engine.config = config
//	go engine.WatchCommitteeChange()
//
//}