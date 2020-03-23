package sawtooth

import (
	"time"
)

const (
	proposePhase  = "PROPOSE"
	changeViewPhase   = "VIEW"
	votePhase     = "VOTE"
	commitPhase = "COMMIT"
)

//
const (
	timeout             = 40 * time.Second       // must be at least twice the time of block interval
	maxNetworkDelayTime = 150 * time.Millisecond // in ms
)
