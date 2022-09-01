package joblog

import (
	"bufio"
	"executor-go/config"
	"executor-go/model"
	"executor-go/util"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

/*
1. 根据配置文件找到写日志的地址
2. 写入日志
3. 请求日志返回日志数据

*/

type JobLog struct {
	jobID	int64
	logFile *os.File
}

func InitJobLog(id int64, execTime int64) *JobLog {
	log := &JobLog{
		jobID: id,
	}
	jobExecTime := execTime/1000
	pathPrefix := time.Unix(jobExecTime, 0).Format("2006-01-02")
	log.openFile(config.LogPath + "/" + pathPrefix)
	return log
}

func (log *JobLog) openFile(path string) {
	if err := os.MkdirAll(path, 0750); err != nil && !os.IsExist(err) {
		fmt.Printf("open file err, mkdir path err, err=%s\n", err)
		return
	}

	logPath := fmt.Sprintf("%s/%d.log", path, log.jobID)
	fmt.Printf("open file = %s\n", logPath)
	file, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("log file open fail: ", err)
		return
	}
	log.logFile = file
}

func getLogPosition() string {
	_, path, line, ok := runtime.Caller(2)
	paths := strings.Split(path, "/")
	if ok {
		return fmt.Sprintf("%s:%d", paths[len(paths)-1], line)
	}
	return "pos_err"
}

func (log *JobLog) Info(format string, args ...interface{}) {
	file := getLogPosition()
	nowTime := time.Now()
	logTime := nowTime.Format("[2006-01-02 15:04:05]")
	info := fmt.Sprintf("[INFO]%s[%s]%s\n", logTime, file, format)
	format = info
	fmt.Fprintf(log.logFile, format, args...)
}

func (log *JobLog) Warn(format string, args ...interface{}) {
	file := getLogPosition()
	nowTime := time.Now()
	logTime := nowTime.Format("[2006-01-02 15:04:05]")
	info := fmt.Sprintf("[WARN]%s[%s]%s\n", logTime, file, format)
	format = info
	fmt.Fprintf(log.logFile, format, args...)
}

func (log *JobLog) Error(format string, args ...interface{}) {
	file := getLogPosition()
	nowTime := time.Now()
	logTime := nowTime.Format("[2006-01-02 15:04:05]")
	info := fmt.Sprintf("[ERROR]%s[%s]%s\n", logTime, file, format)
	format = info
	fmt.Fprintf(log.logFile, format, args...)
}

func (log *JobLog) Debug(format string, args ...interface{}) {
	file := getLogPosition()
	nowTime := time.Now()
	logTime := nowTime.Format("[2006-01-02 15:04:05]")
	info := fmt.Sprintf("[Debug]%s[%s]%s\n", logTime, file, format)
	format = info
	fmt.Fprintf(log.logFile, format, args...)
}

func (log *JobLog) Fatal(format string, args ...interface{}) {
	nowTime := time.Now()
	logTime := nowTime.Format("[2006-01-02 15:04:05]")
	info := fmt.Sprintf("[Debug]%s%s\n", logTime, format)
	format = info
	fmt.Fprintf(log.logFile, format, args...)
}

func (log *JobLog) Close() {
	log.logFile.Close()
}

// GetJobLog 通过req获得任务执行时间和jobID 根据执行时间选择文件夹, 根据jobID得到log
func GetJobLog(req *model.LogReq) *model.LogRes {
	var content string
	resp := &model.LogRes{}
	jobExecTime := req.LogDateTim/1000
	pathPrefix := time.Unix(jobExecTime, 0).Format("2006-01-02")
	path := fmt.Sprintf("%s/%s/%s.log", config.LogPath, pathPrefix, util.Int64ToStr(req.LogID))
	file, err := os.Open(path)
	if err != nil {
		resp.Msg = fmt.Sprintf("Open file err = %s", err)
		resp.Code = model.FailureCode
		return resp
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		if err != nil {
			if err == io.EOF {
				fmt.Println("File read ok!")
				break
			} else {
				resp.Msg = fmt.Sprintf("Read file error = %s, path=%s", err, path)
				resp.Code = model.FailureCode
				return resp
			}
		}
		content = fmt.Sprintf("%s\n%s", content, line)
	}
	resp.Msg = ""
	resp.Code = model.SuccessCode
	resp.Content = model.LogResContent{
		FromLineNum: req.FromLineNum,
		ToLineNum:   2,
		LogContent:  content,
		IsEnd:       true,
	}
	return resp
}