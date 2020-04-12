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
var msgTimerMutex sync.Mutex
var timerMutex sync.Mutex
var printLock sync.Mutex

var (
	ModeMapMutex sync.RWMutex
	PrepareMapMutex sync.RWMutex
)