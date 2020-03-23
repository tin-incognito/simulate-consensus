package consensus

//ConsensusHandler ...
type ConsensusHandler interface {
	Name() (string,error)
	Start() error
}

