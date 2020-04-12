package main

import "sync"

var tempLock sync.Mutex

//MsgTimer ...
type MsgTimer struct{
	//msg NormalMsg
	Type string
}
