package TestServer

import (
	"database/sql"
	"fmt"
	"github.com/nats-io/nuid"
	"net/http"
	"sync"
	"time"
	"utils"
)

// the identification token is set when the server is starting.
// no need for mutex
var identificationToken = "123"
var trust_key = ""
var (
	portListMutex sync.Mutex
	portList      []int
)
var (
	identificationKey      = nuid.New()
	identificationKeyMutex sync.Mutex
)

var (
	JudgeIdKey = nuid.New()
)

var (
	matchIDK      map[int]string
	matchName     map[int]string
	matchIDKMutex sync.Mutex
)

var (
	availTimeStamp      map[int]int64
	availTimeStampMutex sync.Mutex
)

var globalHeartBeat = 0
var globalServerBindAddr = ""

var db *sql.DB

func ServerEntry(ConfigPath string) {
	utils.Logs("server", fmt.Sprintf("Startup with server mode with file %s.", ConfigPath))
	servConfig := utils.ReadFromServerJSON(ConfigPath)
	utils.Logs("server", "Get server configuration file")
	utils.Logs("server", utils.ServerJSONToString(servConfig))
	// make 3 stage: semantic, optimize, codegen
	// here should startup go routine for fetching the data
	stageTestcase = make([][]TestcaseFormat, 3)
	stageTestcaseMutex = make([]sync.Mutex, 3)
	// make sql connection
	db, err := sql.Open(servConfig.DatabaseType, fmt.Sprintf("%s:%s@tcp(%s:%d)%s",
		servConfig.DatabaseUser, servConfig.DatabasePassword, servConfig.DatabaseAddr,
		servConfig.DatabasePort, servConfig.DatabasePath))
	utils.CheckError(err)
	if db == nil {
		panic("Runtime error: failed to open database")
	}
	defer db.Close()
	go func() {
		db, _ := sql.Open(servConfig.DatabaseType, fmt.Sprintf("%s:%s@tcp(%s:%d)%s",
			servConfig.DatabaseUser, servConfig.DatabasePassword, servConfig.DatabaseAddr,
			servConfig.DatabasePort, servConfig.DatabasePath))
		defer db.Close()
		ModifiedSQLCommand := "SELECT semantic, codegen, optim from dataset_updates where type='testcase'"
		SemanticSQLCommand := "SELECT sema_uid, sema_sourceCode, sema_assertion, sema_timeLimit, sema_memoryLimit, sema_testcase from dataset_semantic"
		CodegenSQLCommand := "SELECT cg_uid, cg_sourceCode, cg_inputCtx, cg_outputCtx, cg_outputCode, cg_assertion, cg_timeLimit, cg_instLimit, cg_memoryLimit, cg_testcase from dataset_codegen"
		OptimSQLCommand := "SELECT optim_uid, optim_sourceCode, optim_inputCtx, optim_outputCtx, optim_outputCode, optim_assertion, optim_timeLimit, optim_instLimit, optim_memoryLimit, optim_testcase from dataset_optimize"
		for ;;{
			// update every 5 seconds
			time.Sleep(time.Duration(5) * time.Second)
			modifiedResult, err := db.Query(ModifiedSQLCommand)
			if err != nil {
				continue
			}
			if modifiedResult == nil{
				continue
			}
			var semanticFlag = 0
			var codegenFlag = 0
			var optimizeFlag = 0
			for modifiedResult.Next(){
				err = modifiedResult.Scan(&semanticFlag, &codegenFlag, &optimizeFlag)
				if err != nil{
					utils.Warnings("server::FetchTestcase", err.Error())
					continue
				}
			}
			if semanticFlag != 0 || len(stageTestcase[0]) == 0 {
				stageTestcaseMutex[0].Lock()
				semanticResult, err := db.Query(SemanticSQLCommand)
				if err != nil {
					continue
				}
				if semanticResult == nil{
					continue
				}
				var sema_uid, sema_sourceCode, sema_testcase string
				var sema_assertion, sema_timeLimit, sema_memoryLimit int
				stageTestcase[0] = make([]TestcaseFormat, 0)
				for semanticResult.Next(){
					err = semanticResult.Scan(&sema_uid, &sema_sourceCode, &sema_assertion, &sema_timeLimit, &sema_memoryLimit, &sema_testcase)
					stageTestcase[0] = append(stageTestcase[0], TestcaseFormat{
						IsAssertion:   true,
						ResultId:      "",
						SourceCode:    sema_sourceCode,
						Assertion:     sema_assertion == 1,
						TimeLimit:     float32(sema_timeLimit),
						InstLimit:     0,
						MemoryLimit:   sema_memoryLimit,
						Testcase:      sema_testcase,
						InputContext:  "",
						OutputContext: "",
						OutputCode:    0,
						BasicType:     0,
						Verdict:       0,
						StdOutMessage: "",
						StdErrMessage: "",
						Runtime:       0,
						InstsCount:    0,
						RunningStdOut: "",
						RavelMessage:  "",
						ErrorMessage:  "",
					})
					if err != nil{
						utils.Warnings("server::FetchTestcase::Semantic", err.Error())
						continue
					}
				}
				stageTestcaseMutex[0].Unlock()
				utils.Logs("server::FetchTestcase::Semantic", fmt.Sprintf("Fetched: %d testcases.", len(stageTestcase[0])))
			}
			if codegenFlag != 0 || len(stageTestcase[1]) == 0{
				stageTestcaseMutex[1].Lock()
				codegenResult, err := db.Query(CodegenSQLCommand)
				if err != nil {
					continue
				}
				if codegenResult == nil{
					continue
				}
				var cg_uid, cg_sourceCode, cg_inputCtx, cg_outputCtx, cg_testcase string
				var cg_assertion, cg_timeLimit, cg_memoryLimit, cg_outputCode int
				var cg_instLimit int64
				stageTestcase[1] = make([]TestcaseFormat, 0)
				for codegenResult.Next(){
					err = codegenResult.Scan(&cg_uid, &cg_sourceCode, &cg_inputCtx, &cg_outputCtx, &cg_outputCode,
						&cg_assertion, &cg_timeLimit, &cg_instLimit, &cg_memoryLimit, &cg_testcase)
					stageTestcase[1] = append(stageTestcase[1], TestcaseFormat{
						IsAssertion:   true,
						ResultId:      "",
						SourceCode:    cg_sourceCode,
						Assertion:     cg_assertion == 1,
						TimeLimit:     float32(cg_timeLimit),
						InstLimit:     cg_instLimit,
						MemoryLimit:   cg_memoryLimit,
						Testcase:      cg_testcase,
						InputContext:  cg_inputCtx,
						OutputContext: cg_outputCtx,
						OutputCode:    cg_outputCode,
						BasicType:     0,
						Verdict:       0,
						StdOutMessage: "",
						StdErrMessage: "",
						Runtime:       0,
						InstsCount:    0,
						RunningStdOut: "",
						RavelMessage:  "",
						ErrorMessage:  "",
					})
					if err != nil{
						utils.Warnings("server::FetchTestcase::Codegen", err.Error())
						continue
					}
				}
				stageTestcaseMutex[1].Unlock()
				utils.Logs("server::FetchTestcase::Codegen", fmt.Sprintf("Fetched: %d testcases.", len(stageTestcase[1])))

			}
			if optimizeFlag != 0 || len(stageTestcase[2]) == 0{
				stageTestcaseMutex[2].Lock()
				codegenResult, err := db.Query(OptimSQLCommand)
				if err != nil {
					continue
				}
				if codegenResult == nil{
					continue
				}
				var optim_uid, optim_sourceCode, optim_inputCtx, optim_outputCtx, optim_testcase string
				var optim_assertion, optim_timeLimit, optim_memoryLimit, optim_outputCode int
				var optim_instLimit int64
				stageTestcase[2] = make([]TestcaseFormat, 0)
				for codegenResult.Next(){
					err = codegenResult.Scan(&optim_uid, &optim_sourceCode, &optim_inputCtx, &optim_outputCtx, &optim_outputCode,
						&optim_assertion, &optim_timeLimit, &optim_instLimit, &optim_memoryLimit, &optim_testcase)
					stageTestcase[2] = append(stageTestcase[2], TestcaseFormat{
						IsAssertion:   true,
						ResultId:      "",
						SourceCode:    optim_sourceCode,
						Assertion:     optim_assertion == 1,
						TimeLimit:     float32(optim_timeLimit),
						InstLimit:     optim_instLimit,
						MemoryLimit:   optim_memoryLimit,
						Testcase:      optim_testcase,
						InputContext:  optim_inputCtx,
						OutputContext: optim_outputCtx,
						OutputCode:    optim_outputCode,
						BasicType:     0,
						Verdict:       0,
						StdOutMessage: "",
						StdErrMessage: "",
						Runtime:       0,
						InstsCount:    0,
						RunningStdOut: "",
						RavelMessage:  "",
						ErrorMessage:  "",
					})
					if err != nil{
						utils.Warnings("server::FetchTestcase::Optimize", err.Error())
						continue
					}
				}
				stageTestcaseMutex[2].Unlock()
				utils.Logs("server::FetchTestcase::Optimize", fmt.Sprintf("Fetched: %d testcases.", len(stageTestcase[1])))
			}

		}
	}()
	identificationToken = servConfig.ServerIdentificationToken
	globalHeartBeat = servConfig.Heartbeat
	globalServerBindAddr = servConfig.BindAddr
	trust_key = servConfig.TrustKey
	// make port list, here no need for mutex
	for i := 0; i < servConfig.MaximizeClient; i++ {
		portList = append(portList, i+servConfig.StartPort)
	}
	utils.Logs("server", fmt.Sprintf("make port: from %d to %d", portList[0], portList[len(portList)-1]))
	// bind to the path: http://addr:port/
	// Notice that for different port, we should bind to different multiplexing handler for the same URI
	// e.g.: http://addr:port1/ and http://addr:port2/ should use different handler function

	go func() {
		serverMuxA := http.NewServeMux()
		serverMuxA.HandleFunc("/", RoutineRegisterClient)
		// Dispatch work
		serverMuxA.HandleFunc("/requestWork", RoutineRequestJudgeWork)
		err := http.ListenAndServe(fmt.Sprintf("%s:%d", servConfig.BindAddr, servConfig.Port), serverMuxA)
		utils.CheckError(err)
	}()
	// Dispatch work
	//go func(){
	//	serverMux := http.NewServeMux()
	//	serverMux.HandleFunc("/requestWork", RoutineRequestJudge)
	//	err := http.ListenAndServe(fmt.Sprintf("%s:%d", servConfig.BindAddr, servConfig.ListenPort), serverMux)
	//	utils.CheckError(err)
	//}()
}
