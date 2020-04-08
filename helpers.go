package main

import "sync"

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

//Wait group for actor class
var wgActor sync.WaitGroup