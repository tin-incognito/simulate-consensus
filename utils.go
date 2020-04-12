package main

import "os"

//getEnv ...
func getEnv(key, defaultVal string) string{
	res := os.Getenv(key)
	if res == ""{
		res = defaultVal
	}
	return res
}