package TestServer

import (
	"fmt"
	"utils"
)

func ServerEntry(ConfigPath string){
	utils.Logs("client", fmt.Sprintf("Startup with server mode with file %s.", ConfigPath))
}
