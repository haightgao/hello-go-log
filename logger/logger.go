package logger

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	infoLogger  *log.Logger
	debugLogger *log.Logger
	errLogger   *log.Logger

	logOut       *os.File
	logLevel     int
	logPath      string
	showFileName bool
	currentDay   int

	lock sync.RWMutex //sync.Mutex 与 sync.RWMutex 的区别是 RWMutex 可以加多个读锁，而 Mutex 只能加一个
)

const (
	DebugLevel = iota
	InfoLevel
	ErrorLevel
)

func init() {
	lock = sync.RWMutex{}
}

func SetLevel(level int) {
	logLevel = level
}

func SetShowFileName(show bool) {
	showFileName = show
}

func SetFile(path, file string) {
	lock.Lock()
	defer lock.Unlock()
	file = checkHavePostfix(file)
	checkPath(path)
	fullName := getFullName(file)
	openFile(fullName)

	//infoLogger = log.New(logOut, "[INFO] ", log.LstdFlags)
}

func Info(format string, v ...any) {
	if logLevel <= InfoLevel {
		checkDayChange()
		info := fmt.Sprintf(format, v...)
		infoLogger.Printf("%s %s", getPrefix(), info)
	}
}

func Debug(format string, v ...any) {
	if logLevel <= DebugLevel {
		checkDayChange()
		info := fmt.Sprintf(format, v...)
		debugLogger.Printf("%s %s", getPrefix(), info)
	}
}

func Error(format string, v ...any) {
	getCallTrace()
	if logLevel <= ErrorLevel {
		checkDayChange()
		info := fmt.Sprintf(format, v...)
		errLogger.Printf("%s %s", getPrefix(), info)
	}
}

func openFile(fileName string) {
	currentDay = time.Now().Local().Day()
	var err error
	logOut, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		panic(fmt.Sprintf("Failed to open log file:%s", err.Error()))
	}
	infoLogger = log.New(logOut, "[INFO] ", log.Ldate|log.Ltime) //Ldate:日期 Ltime:时间 Lshortfile:文件名+行号
	debugLogger = log.New(logOut, "[DEBUG] ", log.Ldate|log.Ltime)
	errLogger = log.New(logOut, "[ERROR] ", log.Ldate|log.Ltime)
}

func getCallTrace() (string, string, int) {
	pc, fileName, lineNo, ok := runtime.Caller(3)
	if ok {
		fnName := runtime.FuncForPC(pc).Name()
		return fileName, fnName, lineNo
	}

	return "", "", 0
}

func getPrefix() string {
	fileName, fnName, lineNo := getCallTrace()
	if !showFileName {
		return fmt.Sprintf("%s:%d", fnName, lineNo)
	}
	return fmt.Sprintf("%s->%s:%d", fileName, fnName, lineNo)
}

func checkDayChange() {
	lock.Lock()
	defer lock.Unlock()
	if currentDay != time.Now().Local().Day() {
		logOut.Close()
		fileName := logOut.Name()
		historyFileName := getHistoryFileName(fileName)
		err := os.Rename(fileName, historyFileName)
		if err != nil {
			panic(errors.New(fmt.Sprintf("Failed to rename file:%s", err.Error())))
		}
		openFile(fileName)
	}
}

func checkPath(path string) {
	logPath = path
	// 判断path是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 不存在则创建
		os.Mkdir(path, os.ModePerm)
	}
}

func getFullName(filename string) string {
	// 获取系统路径分隔符
	pathSep := string(os.PathSeparator)
	// 获取根路径
	rootPath, err := os.Getwd()
	if err != nil {
		panic(errors.New(fmt.Sprintf("Failed to get root path:%s", err.Error())))
	}
	// 拼接路径
	return rootPath + pathSep + logPath + pathSep + filename
}

func checkHavePostfix(filename string) string {
	// 检查文件名是否存在英文点
	if index := strings.Index(filename, "."); index == -1 {
		// 不存在则添加后缀
		return filename + ".log"
	}

	return filename
}

func getHistoryFileName(filename string) string {
	// 获取系统文件分隔符
	pathSep := string(os.PathSeparator)
	// 获取文件名
	fileName := filename[strings.LastIndex(filename, pathSep)+1:]
	path := filename[:strings.LastIndex(filename, pathSep)]
	newFileName := fileName + "." + time.Now().Local().Add(-24*time.Hour).Format("2006-01-02")
	return path + pathSep + newFileName
}
