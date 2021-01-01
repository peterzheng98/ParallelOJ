package main

import (
	"TestClient"
	"TestServer"
	"flag"
	_ "github.com/go-git/go-git"
)

var cliType = flag.String("mode", "", "Set the judge type. [client, server]")
var cliConfig = flag.String("config", "", "Set the configure file")

func main(){
	flag.Parse()
	if *cliType == "client"{
		TestClient.ClientEntry(*cliConfig)
	} else if *cliType == "server"{
		TestServer.ServerEntry(*cliConfig)
	} else {
		panic("Server Startup Error. Not valid type for judger")
	}
}