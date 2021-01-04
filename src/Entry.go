package main

import (
	"TestClient"
	"TestServer"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/codeskyblue/go-sh"
	_ "github.com/docker/docker/api/types"
	"github.com/go-git/go-git"
	_ "github.com/go-git/go-git"
	"github.com/go-git/go-git/plumbing"
	"io/ioutil"
	"time"
	"utils"
)

var cliType = flag.String("mode", "", "Set the judge type. [client, server]")
var cliConfig = flag.String("config", "", "Set the configure file")

func main(){
	t1 := time.Now()
	tags := make([]string, 1)
	tags[0] = "abc:b8ed972e"
	//sess, err := sh.Command("docker", "build", "-t", fmt.Sprintf("%s", tags[0]), ".", sh.Dir("/tmp/clone-temp")).CombinedOutput()
	////var q []byte
	////var p []byte
	////cnt, _ := sess.Stderr.Write(p)
	////cnt2, err := sess.Stdout.Write(q)
	//fmt.Printf("output: count:%d \n%s\n", 0, sess)
	//fmt.Printf("stderr: count:%d \n%s\n", 0, sess)
	//
	//utils.CheckError(err)
	//
	//fmt.Printf("Encoded:\n%s\n", base64.StdEncoding.EncodeToString(sess))
	//_ = sh.Command("rm", "-rf", "docker.compiler.tar", "clone-temp", sh.Dir("/tmp")).Run()
	r, err := git.PlainClone("/tmp/clone-temp/src", false, &git.CloneOptions{
		URL: "https://github.com/peterzheng98/ParallelOJCompatCompiler.git",
	})
	w, err := r.Worktree()
	if err != nil {
		//utils.Warnings(fmt.Sprintf("TestClient:[Clone user %s, git address: %s, target hash: %s]", replyMess.User, replyMess.GitRepo, replyMess.GitHash), err.Error())
		// TODO: reply: bad package
	}
	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash("b8ed972e89d69811d711b372d59cd6f96f6df75e"),
	})
	if err != nil {
		//utils.Warnings(fmt.Sprintf("TestClient:[Build user %s, git address: %s, target hash: %s]", replyMess.User, replyMess.GitRepo, replyMess.GitHash), err.Error())
		// TODO: reply: bad package
	}
	contents := []byte("FROM base:v1.0\nWORKDIR /src\nCOPY src .\nRUN /bin/bash build.bash")
	err = ioutil.WriteFile("/tmp/clone-temp/Dockerfile", contents, 0644)
	utils.CheckError(err)
	_ = sh.Command("tar", "-czf", "../docker.compiler.tar", ".", sh.Dir("/tmp/clone-temp")).Run()
	elapsed := time.Since(t1)
	fmt.Println("Elapsed: ", elapsed.Milliseconds())
	//ctx := context.Background()
	//cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	//if err != nil{
	//	utils.CheckError(err)
	//}
	//images, err := cli.ImageList(ctx, types.ImageListOptions{
	//	All: false,
	//})
	//for _, l := range images{
	//	fmt.Printf("%s ", l.RepoTags)
	//}
	// file, err := os.Open("/tmp/docker.compiler.tar")
	utils.CheckError(err)

	out, err := sh.Command("docker", "build", "-t", fmt.Sprintf("%s", tags[0]), ".", sh.Dir("/tmp/clone-temp")).Output()
	utils.CheckError(err)
	fmt.Printf("output:\n%s\n", out)
	fmt.Printf("Encoded:\n%s\n", base64.StdEncoding.EncodeToString(out))
	flag.Parse()
	if *cliType == "client"{
		TestClient.ClientEntry(*cliConfig)
	} else if *cliType == "server"{
		TestServer.ServerEntry(*cliConfig)
	} else {
		panic("Server Startup Error. Not valid type for judger")
	}
}