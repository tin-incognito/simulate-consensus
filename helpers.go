package main

import (
	"log"
	"sync"
)

// Mutex for blockchain class
var logBlockMutex sync.Mutex
var insertBlockMutex sync.Mutex

// Mutex for actor class
var prePrepareMutex sync.Mutex
var prepareMutex sync.Mutex
var commitMutex sync.Mutex
var viewChangeMutex sync.Mutex
var saveMsgMutex sync.Mutex
var sendMsgMutex sync.Mutex
var modeMutex sync.Mutex
var switchMutex sync.Mutex
var msgTimerMutex sync.Mutex
var timerMutex sync.Mutex
var updateModeMutex sync.Mutex
var switchViewChangeModeMutex sync.Mutex
var switchNormalModeMutex sync.Mutex
var timeOutChMutex sync.Mutex
var viewChangingMutex sync.Mutex

var (
	handleMsgTimerMutex sync.Mutex
	handleTimerMutex sync.Mutex
	PutMapMutex sync.RWMutex
	GetMapMutex sync.RWMutex
	prepareTimerMutex sync.RWMutex
	commitTimerMutex sync.RWMutex
	tempMutex sync.Mutex
	PrepareMapMutex sync.RWMutex
	CommitMapMutex sync.RWMutex
)

//Wait group for actor class
var wgActor sync.WaitGroup

func debug(){
	for _, node := range nodes{
		log.Println("node:", node.index)
		log.Println("info:", node)
		log.Println("actor:", node.consensusEngine.BFTProcess)
	}
}