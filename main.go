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

	err = pool.start()
	if err != nil{
		panic(err)
	}
}