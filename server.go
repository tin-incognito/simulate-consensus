package main

import (
	"github.com/tin-incognito/simulate-consensus/consensus"
)

//Server for storing consensus engine
//and blockchain interface
//For purpose build this machine to become a node in network
type Server struct{
	//enable bool
	startUpTime int64
	miningKeys      string
	privateKey      string
	isEnableMining    bool
	consensusEngine *consensus.Engine
	listenAddrs string
}

////NewServer return a server object
//func (server *Server) NewServer(cfg config) (*Server, error){
//	var err error
//
//	consensusEngine := &consensus.Engine{}
//	consensusEngine, err = consensusEngine.NewEngine()
//
//	res := &Server{
//		startUpTime: time.Now().Unix(),
//		miningKeys: cfg.MiningKeys,
//		privateKey: cfg.PrivateKey,
//		listenAddrs: cfg.listenAddrs,
//		consensusEngine: consensusEngine,
//	}
//
//	return res, err
//}

//start the server for starting node
//to start working in the system
func (server *Server) start() error {
	//server.consensusEngine.Init(&consensus.EngineConfig{Node: serverObj, Blockchain: serverObj.blockChain, PubSubManager: serverObj.pusubManager})
	return nil
}
