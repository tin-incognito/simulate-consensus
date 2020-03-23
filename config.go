package main

//config for storing data from reading config
type config struct{
	MiningKeys        string `long:"miningkeys" description:"keys used for different consensus algorigthm"`
	PrivateKey        string `long:"privatekey" description:"your wallet privatekey"`
	listenAddrs string `long:"listenaddrs" description:"Listenaddrs node will run on"`

}

//loadConfig for load config from file
//or env variables
func loadConfig() (*config, error){
	return nil, nil
}