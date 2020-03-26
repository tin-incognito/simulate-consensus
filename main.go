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

	//time.Sleep(time.Millisecond * 500)

	err = pool.simulate(chain)

	if err != nil {
		panic(err)
	}

	time.Sleep(time.Millisecond * 1500)

}