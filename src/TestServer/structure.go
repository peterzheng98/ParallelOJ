package TestServer

type ServerConfig_FileJSON struct {
	Mode                      string `json:"mode"`
	Port                      int    `json:"port"`
	BindAddr                  string `json:"bind_addr"`
	ListenPort                int    `json:"listen_port"`
	Heartbeat                 int    `json:"heartbeat"`
	MaximizeClient            int    `json:"maximize_client"`
	HeartbeatTimeout          int    `json:"heartbeat_timeout"`
	DatabaseType              string `json:"database_type"`
	DatabaseUser              string `json:"database_user"`
	DatabasePassword          string `json:"database_password"`
	DatabaseAddr              string `json:"database_addr"`
	DatabasePort              int    `json:"database_port"`
	DatabasePath              string `json:"database_path"`
	ServerIdentificationToken string `json:"server_identification_token"`
	StartPort                 int    `json:"start_port"`
	TrustKey                  string `json:"trust_key"`
}

type Ports struct {
	Port              int    `json:"port"`
	IdentificationKey string `json:"identification_key"`
}

type DispatchedWorkSlice struct {
	User    string           `json:"user"`
	GitRepo string           `json:"git_repo"`
	GitHash string           `json:"git_hash"`
	PhaseId int              `json:"phase_id"` // -1 as compile, 0-n as index
	WorkCnt int              `json:"work_cnt"`
	Cases   []TestcaseFormat `json:"cases"`
}
