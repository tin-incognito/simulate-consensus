package main

import (
	"time"
)

var f, n uint64
var blockChain *Chain
var pool *Pool

func main(){
	blockChain = &Chain{
		LatestBlock:      nil,
		Height:           0,
		validatorsAmount: n,
	}

	pool = &Pool{}
	var err error

	n = 4

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

	time.Sleep(time.Millisecond * 1000)

	err = pool.simulate()

	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 20)

}