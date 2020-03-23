package main

import "log"

func main(){
	//load config from anywhere available
	cfg , err := loadConfig()

	if err != nil {
		log.Fatal(err)
	}

	server := &Server{}
	server, err = server.NewServer(*cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := server.start(); err !=nil{
		log.Fatal(err)
	}


}