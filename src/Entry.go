package main

import (
	"TestClient"
	"TestServer"
	"context"
	"flag"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	_ "github.com/go-git/go-git"
	"utils"
)

var cliType = flag.String("mode", "", "Set the judge type. [client, server]")
var cliConfig = flag.String("config", "", "Set the configure file")

func main(){
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil{
		utils.CheckError(err)
	}
	images, err := cli.ImageList(ctx, types.ImageListOptions{
		All: false,
	})
	for _, l := range images{
		fmt.Printf("%s ", l.RepoTags)
	}
	flag.Parse()
	if *cliType == "client"{
		TestClient.ClientEntry(*cliConfig)
	} else if *cliType == "server"{
		TestServer.ServerEntry(*cliConfig)
	} else {
		panic("Server Startup Error. Not valid type for judger")
	}
}