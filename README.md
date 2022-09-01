个人添加了task任务log以及返回的log
服务log与任务log区分
# xxl-job-executor-go
很多公司java与go开发共存，java中有xxl-job做为任务调度引擎，为此也出现了go执行器(客户端)，使用起来比较简单：
# 支持
```	
1.执行器注册
2.耗时任务取消
3.任务注册，像写http.Handler一样方便
4.任务panic处理
5.阻塞策略处理
6.任务完成支持返回执行备注
7.任务超时取消 (单位：秒，0为不限制)
8.失败重试次数(在参数param中，目前由任务自行处理)
9.可自定义日志
10.自定义日志查看handler
11.支持外部路由（可与gin集成）
```

# Example
```
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
	config.LogPath = "/data01/xxl-job/log/jobhandler"   //存放日志的地址
	execLog := logrus.New()
	execLog.SetReportCaller(true)
	execLog.SetFormatter(&util.LogFormatter{})
	exec := handler.NewExecutor(
		handler.ServerAddr("http://127.0.0.1:11111/xxl-job-admin"),
		handler.AccessToken("default_token"),    //请求令牌(默认为空)
		handler.ExecutorIp("127.0.0.1"),         //可自动获取
		handler.ExecutorPort("11112"),           //默认9999（非必填）
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



//xxl.Logger接口实现
type logger struct{}

joblog:
    task的log
logrus:
    服务的log
