package utils

import (
	"fmt"
	"time"
)

var colored = true

func getTime() string {
	return time.Now().Format("2020-12-17 14:01:30")
}

func Logs(name string, message string){
	if colored {
		fmt.Printf("\033[33m[%s][%s]: %s\033[0m\n", getTime(), name, message)
	} else {
		fmt.Printf("[%s][%s]: %s\n", getTime(), name, message)
	}
}

func Warnings(name string, message string)  {
	if colored {
		fmt.Printf("\033[36m[%s][%s]: %s\033[0m\n", getTime(), name, message)
	} else {
		fmt.Printf("[%s][%s]: %s\n", getTime(), name, message)
	}
}

func CheckError(err error){
	if err != nil{
		fmt.Printf("[%s][FATAL ERROR]: %s", getTime(), err.Error())
		panic(err)
	}
}
