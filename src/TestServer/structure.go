package TestServer

type ServerConfig_FileJSON struct {
	Mode                      string `json:"mode"`
	Port                      int    `json:"port"`
	BindAddr                  string `json:"bind_addr"`
	ListenAddr                string `json:"listen_addr"`
	Heartbeat                 int    `json:"heartbeat"`
	MaximizeClient            int    `json:"maximize_client"`
	HeartbeatTimeout          int    `json:"heartbeat_timeout"`
	DatabaseType              string `json:"database_type"`
	DatabaseUser              string `json:"database_user"`
	DatabasePassword          string `json:"database_password"`
	DatabaseAddr              string `json:"database_addr"`
	DatabasePort              string `json:"database_port"`
	DatabasePath              string `json:"database_path"`
	ServerIdentificationToken string `json:"server_identification_token"`
	StartPort                 int    `json:"start_port"`
}
