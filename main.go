package main

var f, n uint64
var pool *Pool
var nodes map[int]*Node

func main(){

	// Refactor add chain to node class
	// Init array for nodes (not necessary for new class)
	//

	//var err error
	n = 4

	f = uint64((n - 1) / 3)

	initNodes(int(n))
	start()
	simulate()

	select {

	}
}