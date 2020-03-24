package main

import "log"

func main(){
	pool := &Pool{}
	var err error

	pool, err = pool.createPool(5)

	if err != nil {
		panic(err)
	}

	log.Println(pool)

	err = pool.initValidators()

	if err != nil {
		panic(err)
	}

	log.Println(pool.Nodes[0])

	err = pool.start()
	if err != nil{
		panic(err)
	}

	chain := &Chain{
		LatestBlock:      nil,
		Height:           0,
		ValidatorsAmount: 0,
	}

	err = pool.simulate(chain)

	if err != nil {
		panic(err)
	}

}