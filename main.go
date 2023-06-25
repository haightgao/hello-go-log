package main

import (
	"hello-go-log/logger"
)

func main() {
	logger.SetFile("logs", "test.log")
	//logger.SetLevel(logger.InfoLevel)
	logger.Info("hello world %s", "test")
	//logger.Debug("hello world %s", "test")
	//logger.Error("hello world %s", errors.New("test error").Error())
}
