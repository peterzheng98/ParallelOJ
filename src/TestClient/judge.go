package TestClient

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/codeskyblue/go-sh"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
)

func judgeSemantic(source string, dockerTags string, expectedAssertion bool, timeout float32, memoryOut int) *JudgeResult {
	se := JudgeResult{}
	// dump the file out to input.mx
	inputPath := fmt.Sprintf("%s/input.mx", dataPath)
	srcDec, _ := base64.StdEncoding.DecodeString(source)
	_ = ioutil.WriteFile(inputPath, srcDec, 0644)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	// docker cmd
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm", "--cpus=1", "-m", fmt.Sprintf("%dM", memoryOut), "-v",
		fmt.Sprintf("%s:%s", dataPath, "/mounted"), dockerTags, "/bin/bash", "/JudgeSemantic.bash")
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	outputReader := bufio.NewReader(stdout)
	errorReader := bufio.NewReader(stderr)

	stdoutStr := ""
	stderrStr := ""
	startTime := time.Now()
	_ = cmd.Start()
	err := cmd.Wait()
	elapsed := time.Since(startTime).Milliseconds()
	// read stdout
	for {
		stdoutSlice, err := outputReader.ReadString('\n')
		if err == io.EOF {
			break
		}
		stdoutStr = stdoutStr + "\n" + stdoutSlice
	}

	// read stderr
	for {
		stderrSlice, err := errorReader.ReadString('\n')
		if err == io.EOF{
			break
		}
		stderrStr = stderrStr + "\n" + stderrSlice
	}

	if err == context.DeadlineExceeded{
		se.Verdict = VERDICT_TLE
	} else {
		judgeAssertion := err == nil
		if judgeAssertion == expectedAssertion{
			se.Verdict = VERDICT_CORRECT
		} else {
			se.Verdict = VERDICT_WRONG
		}
	}
	se.InstsCount = 0
	se.Runtime = float32(elapsed) / 1000.0
	se.StdErrMessage = base64.StdEncoding.EncodeToString([]byte(stderrStr))
	se.StdOutMessage = base64.StdEncoding.EncodeToString([]byte(stdoutStr))
	se.RavelMessage = ""
	se.RunningStdOut = ""
	se.ErrorMessage = ""
	return &se
}


func judgeCodegen(source string, dockerTags string, inputContext string, outputContext string, exitCode int, timeout float32, memoryOut int, mode int, commitHash string) *JudgeResult {
	se := JudgeResult{
		Verdict:       0,
		StdOutMessage: "",
		StdErrMessage: "",
		Runtime:       0,
		InstsCount:    0,
		RunningStdOut: "",
		RavelMessage:  "",
		ErrorMessage: "",
	}
	bashName := ""
	if mode == JUDGE_MODE_CODEGEN {
		bashName = "/JudgeCodegen.bash"
	} else if mode == JUDGE_MODE_OPTIMIZE {
		bashName = "/JudgeOptimize.bash"
	}
	// dump the file out to input.mx
	inputPath := fmt.Sprintf("%s/input.mx", dataPath)
	srcDec, _ := base64.StdEncoding.DecodeString(source)
	_ = ioutil.WriteFile(inputPath, srcDec, 0644)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout + 3)*time.Second)
	// docker cmd
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm", "--cpus=1", "-m", fmt.Sprintf("%dM", memoryOut), "-v",
		fmt.Sprintf("%s:%s", dataPath, "/mounted"), dockerTags, "/bin/bash", bashName)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	outputReader := bufio.NewReader(stdout)
	errorReader := bufio.NewReader(stderr)

	stdoutStr := ""
	stderrStr := ""
	startTime := time.Now()
	_ = cmd.Start()
	err := cmd.Wait()
	elapsed := time.Since(startTime).Milliseconds()
	// read stdout
	for {
		stdoutSlice, err := outputReader.ReadString('\n')
		if err == io.EOF {
			break
		}
		stdoutStr = stdoutStr + "\n" + stdoutSlice
	}

	// read stderr
	for {
		stderrSlice, err := errorReader.ReadString('\n')
		if err == io.EOF{
			break
		}
		stderrStr = stderrStr + "\n" + stderrSlice
	}
	cancel()
	se.InstsCount = -1
	se.Runtime = float32(elapsed) / 1000.0
	se.StdErrMessage = base64.StdEncoding.EncodeToString([]byte(stderrStr))
	se.StdOutMessage = base64.StdEncoding.EncodeToString([]byte(stdoutStr))
	if err == context.DeadlineExceeded{
		se.Verdict = VERDICT_TLE
		return &se
	} else if err != nil{
		se.Verdict = VERDICT_CE
		return &se
	}
	// validate the output assembly
	// todo

	// run by ravel
	// clean the previous result
	_, _ = sh.Command("touch", "output.s", sh.Dir(ravelHeader)).CombinedOutput()
	_, _ = sh.Command("touch", "test.in", sh.Dir(ravelHeader)).CombinedOutput()
	_, _ = sh.Command("touch", "test.out", sh.Dir(ravelHeader)).CombinedOutput()
	_, _ = sh.Command("touch", "run-ravel.sh", sh.Dir(ravelHeader)).CombinedOutput()
	_, _ = sh.Command("rm", "output.s", sh.Dir(ravelHeader)).CombinedOutput()
	_, _ = sh.Command("rm", "test.in", sh.Dir(ravelHeader)).CombinedOutput()
	_, _ = sh.Command("rm", "test.out", sh.Dir(ravelHeader)).CombinedOutput()
	_, _ = sh.Command("rm", "run-ravel.sh", sh.Dir(ravelHeader)).CombinedOutput()
	_, _ = sh.Command("touch", "report.txt", sh.Dir(ravelHeader)).CombinedOutput()
	_, _ = sh.Command("touch", "report.err", sh.Dir(ravelHeader)).CombinedOutput()
	_, _ = sh.Command("rm", "report.txt", sh.Dir(ravelHeader)).CombinedOutput()
	_, _ = sh.Command("rm", "report.err", sh.Dir(ravelHeader)).CombinedOutput()

	_, err = sh.Command("cp", fmt.Sprintf("%s/output.s", dataPath), ravelHeader).CombinedOutput()
	if err != nil {
		se.Verdict = VERDICT_UNK
		se.ErrorMessage = base64.StdEncoding.EncodeToString([]byte(err.Error()))
		return &se
	}
	inputCtx, _ := base64.StdEncoding.DecodeString(inputContext)
	err = ioutil.WriteFile(fmt.Sprintf("%s/test.in", ravelHeader), inputCtx, 0644)
	if err != nil {
		se.Verdict = VERDICT_UNK
		se.ErrorMessage = base64.StdEncoding.EncodeToString([]byte(err.Error()))
		return &se
	}
	ravelBashContent := []byte("ravel --input-file=test.in --output-file=test.out --enable-cache output.s 1>report.txt 2>report.err")
	_ = ioutil.WriteFile(fmt.Sprintf("%s/run-ravel.bash", ravelHeader), ravelBashContent, 0644)
	_, err = sh.Command("bash", "run-ravel.bash", sh.Dir(ravelHeader)).CombinedOutput()
	if err != nil{
		se.Verdict = VERDICT_RE
		file, err2 := os.Open(fmt.Sprintf("%s/report.err", ravelHeader))
		if err2 == nil {
			contents, _ := ioutil.ReadAll(file)
			se.ErrorMessage = base64.StdEncoding.EncodeToString([]byte(err.Error() + "\nError Output:\n" + string(contents)))
		} else {
			se.ErrorMessage = base64.StdEncoding.EncodeToString([]byte(err.Error()))
		}
		return &se
	}
	outputCtx, _ := base64.StdEncoding.DecodeString(outputContext)
	outputCtxStr := string(outputCtx)
	for outputCtxStr[len(outputCtxStr) - 1] == '\n' {
		outputCtxStr = outputCtxStr[:len(outputCtxStr) - 1]
	}
	contentsMatched := false
	exitcodeMatched := false
	stdoutMesssage := ""
	file, err2 := os.Open(fmt.Sprintf("%s/test.out", ravelHeader))
	if err2 == nil {
		// fetch the simulation output
		contents, _ := ioutil.ReadAll(file)
		contentsStr := string(contents)
		for contentsStr[len(contentsStr) - 1] == '\n' {
			contentsStr = contentsStr[:len(contentsStr) - 1]
		}
		contentsMatched = contentsStr == outputCtxStr
		if !contentsMatched {
			stdoutMesssage = "Output: Passed\n"
		} else {
			stdoutMesssage = fmt.Sprintf("Output: Failed\nExpected:%s\nReceived:%s", outputCtxStr, contentsStr)
		}


	}
	return &se
}