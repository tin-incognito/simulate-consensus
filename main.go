package main

import (
	"time"
)

var f, n uint64

func main(){
	pool := &Pool{}
	var err error

	n = 2

	pool, err = pool.createPool(int(n))

	if err != nil {
		panic(err)
	}

	err = pool.initValidators()

	if err != nil {
		panic(err)
	}

	err = pool.start()
	if err != nil{
		panic(err)
	}

	f = uint64((n - 1) / 3)

	chain := &Chain{
		LatestBlock:      nil,
		Height:           0,
		validatorsAmount: n,
	}

	err = pool.simulate(chain)

	if err != nil {
		panic(err)
	}

	//msg := NormalMsg{
	//	Type:      PREPARE,
	//	View:      1,
	//	SeqNum:    1,
	//	SignerID:  1,
	//	Timestamp: uint64(time.Now().Unix()),
	//	BlockID:   nil,
	//	block: nil,
	//}
	//
	//pool.Nodes[0].consensusEngine.BFTProcess.PrepareMsgCh <- msg

	time.Sleep(time.Millisecond * 1500)

}