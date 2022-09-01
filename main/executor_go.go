package main

import (
	"executor-go/config"
	"executor-go/handler"
	"executor-go/joblog"
	"executor-go/task"
	"executor-go/util"
	"github.com/sirupsen/logrus"
	"log"
)

func main() {
	config.LogPath = "/data01/xxl-job/log/jobhandler"
	execLog := logrus.New()
	execLog.SetReportCaller(true)
	execLog.SetFormatter(&util.LogFormatter{})
	exec := handler.NewExecutor(
		handler.ServerAddr("http://127.0.0.1:18822/xxl-job-admin"),
		handler.AccessToken("default_token"),    //请求令牌(默认为空)
		handler.ExecutorIp("127.0.0.1"),         //可自动获取
		handler.ExecutorPort("18801"),           //默认9999（非必填）
		handler.RegistryKey("golang-jobs"),      //执行器名称
		handler.SetLogger(execLog), //自定义日志
	)
	exec.Init()
	//设置日志查看handler
	exec.LogHandler(joblog.GetJobLog)
	//注册任务handler
	exec.RegTask("task.test", task.Test)
	log.Fatal(exec.Run())
}

