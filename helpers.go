package main

import (
	"sync"
)

// Mutex for blockchain class
var logBlockMutex sync.Mutex
var insertBlockMutex sync.Mutex

// Mutex for actor class
var prepareMutex sync.RWMutex
var saveMsgMutex sync.Mutex
var modeMutex sync.Mutex
var switchMutex sync.Mutex
var msgTimerMutex sync.Mutex
var timerMutex sync.Mutex
var viewChangeMutex sync.Mutex

var (
	ModeMapMutex sync.RWMutex
	PrepareMapMutex sync.RWMutex
	CommitMapMutex sync.RWMutex
	PutMapMutex sync.RWMutex
)

//Wait group for actor class
var wgActor sync.WaitGroup