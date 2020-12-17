package TestServer

import (
	"fmt"
	"utils"
)

func ServerEntry(ConfigPath string){
	utils.Logs("server", fmt.Sprintf("Startup with server mode with file %s.", ConfigPath))
	servConfig := utils.ReadFromServerJSON(ConfigPath)
	utils.Logs("server", "Get server configuration file")
	utils.Logs("server", utils.ServerJSONToString(servConfig))

}
