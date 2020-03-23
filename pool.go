package main

import (
	"github.com/tin-incognito/simulate-consensus/node"
	"log"
)

//Pool ...
//Pool will stores the nodes in network
type Pool struct{
	Nodes map[int]*node.Node
}

//createPool for storing nodes
func (pool *Pool) createPool(amountOfNode int) (*Pool, error){
	res := &Pool{}
	res.Nodes = make(map[int]*node.Node)

	for i := 0; i < amountOfNode; i++{
		n := &node.Node{}
		var err error
		n, err = n.CreateNode(i)
		if err != nil {
			log.Println(err)
			continue
		}
		res.Nodes[i] = n
	}

	return res, nil
}

//start ..
func (pool *Pool) start() error {

	for _, node := range pool.Nodes{
		node.Start()
	}
	return nil
}


