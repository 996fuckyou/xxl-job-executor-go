package handler

import (
	"context"
	"encoding/json"
	"executor-go/joblog"
	"executor-go/model"
	"executor-go/util"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Executor 执行器
type Executor interface {
	// Init 初始化
	Init(...Option)
	// LogHandler 日志查询
	LogHandler(handler LogHandler)
	// RegTask 注册任务
	RegTask(pattern string, task TaskFunc)
	// RunTask 运行任务
	RunTask(writer http.ResponseWriter, request *http.Request)
	// KillTask 杀死任务
	KillTask(writer http.ResponseWriter, request *http.Request)
	// TaskLog 任务日志
	TaskLog(writer http.ResponseWriter, request *http.Request)
	// Beat 心跳检测
	Beat(writer http.ResponseWriter, request *http.Request)
	// IdleBeat 忙碌检测
	IdleBeat(writer http.ResponseWriter, request *http.Request)
	// Run 运行服务
	Run() error
	// Stop 停止服务
	Stop()
}

// NewExecutor 创建执行器
func NewExecutor(opts ...Option) Executor {
	return newExecutor(opts...)
}

func newExecutor(opts ...Option) *executor {
	options := newOptions(opts...)
	executor := &executor{
		opts: options,
	}
	return executor
}

type executor struct {
	opts    Options
	address string
	regList *taskList //注册任务列表
	runList *taskList //正在执行任务列表
	mu      sync.RWMutex
	log     *logrus.Logger

	logHandler LogHandler //日志查询handler
}

func (e *executor) Init(opts ...Option) {
	for _, o := range opts {
		o(&e.opts)
	}
	e.log = e.opts.l
	e.regList = &taskList{
		data: make(map[string]*Task),
	}
	e.runList = &taskList{
		data: make(map[string]*Task),
	}
	e.address = e.opts.ExecutorIp + ":" + e.opts.ExecutorPort
	go e.registry()
}

// LogHandler 日志handler
func (e *executor) LogHandler(handler LogHandler) {
	e.logHandler = handler
}

func (e *executor) Run() (err error) {
	// 创建路由器
	mux := http.NewServeMux()
	// 设置路由规则
	mux.HandleFunc("/run", e.runTask)
	mux.HandleFunc("/kill", e.killTask)
	mux.HandleFunc("/log", e.taskLog)
	mux.HandleFunc("/beat", e.beat)
	mux.HandleFunc("/idleBeat", e.idleBeat)
	// 创建服务器
	server := &http.Server{
		Addr:         e.address,
		WriteTimeout: time.Second * 3,
		Handler:      mux,
	}
	// 监听端口并提供服务
	e.log.Infof("Starting server at %s", e.address)
	go server.ListenAndServe()
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	e.registryRemove()
	return nil
}

func (e *executor) Stop() {
	e.registryRemove()
}

// RegTask 注册任务
func (e *executor) RegTask(pattern string, task TaskFunc) {
	var t = &Task{}
	t.fn = task
	e.regList.Set(pattern, t)
	return
}

//运行一个任务
func (e *executor) runTask(writer http.ResponseWriter, request *http.Request) {
	e.mu.Lock()
	defer e.mu.Unlock()
	req, _ := ioutil.ReadAll(request.Body)
	param := &model.RunReq{}
	err := json.Unmarshal(req, &param)
	if err != nil {
		_, _ = writer.Write(util.ReturnCall(param, model.FailureCode, "params err"))
		e.log.Errorf("参数解析错误:%s", string(req))
		return
	}
	e.log.Infof("任务参数:%+v", param)
	if !e.regList.Exists(param.ExecutorHandler) {
		_, _ = writer.Write(util.ReturnCall(param, model.FailureCode, "Task not registered"))
		e.log.Errorf("任务[%d]没有注册:%s", param.JobID, param.ExecutorHandler)
		return
	}

	//阻塞策略处理
	if e.runList.Exists(util.Int64ToStr(param.JobID)) {
		if param.ExecutorBlockStrategy == model.CoverEarly { //覆盖之前调度
			oldTask := e.runList.Get(util.Int64ToStr(param.JobID))
			if oldTask != nil {
				oldTask.Cancel()
				e.runList.Del(util.Int64ToStr(oldTask.Id))
			}
		} else { //单机串行,丢弃后续调度 都进行阻塞
			_, _ = writer.Write(util.ReturnCall(param, model.FailureCode, "There are tasks running"))
			e.log.Errorf("任务[%d]已经在运行了:%s", param.JobID, param.ExecutorHandler)
			return
		}
	}

	cxt := context.Background()
	task := e.regList.Get(param.ExecutorHandler)
	if param.ExecutorTimeout > 0 {
		task.Ext, task.Cancel = context.WithTimeout(cxt, time.Duration(param.ExecutorTimeout)*time.Second)
	} else {
		task.Ext, task.Cancel = context.WithCancel(cxt)
	}
	task.Id = param.JobID
	task.Name = param.ExecutorHandler
	task.Param = param
	task.log = joblog.InitJobLog(param.JobID, param.LogDateTime)

	e.runList.Set(util.Int64ToStr(task.Id), task)
	go task.Run(func(code int64, msg string) {
		e.callback(task, code, msg)
	})
	e.log.Infof("任务[%d]开始执行:%+v", param.JobID, param)
	_, _ = writer.Write(util.ReturnGeneral())
}

//删除一个任务
func (e *executor) killTask(writer http.ResponseWriter, request *http.Request) {
	e.mu.Lock()
	defer e.mu.Unlock()
	req, _ := ioutil.ReadAll(request.Body)
	param := &model.KillReq{}
	_ = json.Unmarshal(req, &param)
	if !e.runList.Exists(util.Int64ToStr(param.JobID)) {
		_, _ = writer.Write(util.ReturnKill(param, model.FailureCode))
		e.log.Errorf("任务[%d]没有运行", param.JobID)
		return
	}
	task := e.runList.Get(util.Int64ToStr(param.JobID))
	task.Cancel()
	e.runList.Del(util.Int64ToStr(param.JobID))
	_, _ = writer.Write(util.ReturnGeneral())
}

//任务日志
func (e *executor) taskLog(writer http.ResponseWriter, request *http.Request) {
	var res *model.LogRes
	data, err := ioutil.ReadAll(request.Body)
	req := &model.LogReq{}
	if err != nil {
		e.log.Errorf("日志请求失败:%s", err.Error())
		reqErrLogHandler(writer, req, err)
		return
	}
	err = json.Unmarshal(data, &req)
	if err != nil {
		e.log.Errorf("日志请求解析失败:%s", err.Error())
		reqErrLogHandler(writer, req, err)
		return
	}
	e.log.Infof("日志请求参数:%+v", req)
	if e.logHandler != nil {
		res = e.logHandler(req)
	} else {
		res = defaultLogHandler(req)
	}
	str, _ := json.Marshal(res)
	_, _ = writer.Write(str)
}

// 心跳检测
func (e *executor) beat(writer http.ResponseWriter, request *http.Request) {
	e.log.Info("心跳检测")
	_, _ = writer.Write(util.ReturnGeneral())
}

// 忙碌检测
func (e *executor) idleBeat(writer http.ResponseWriter, request *http.Request) {
	e.mu.Lock()
	defer e.mu.Unlock()
	req, _ := ioutil.ReadAll(request.Body)
	param := &model.IdleBeatReq{}
	err := json.Unmarshal(req, &param)
	if err != nil {
		_, _ = writer.Write(util.ReturnIdleBeat(model.FailureCode))
		e.log.Errorf("参数解析错误:%v", string(req))
		return
	}
	if e.runList.Exists(util.Int64ToStr(param.JobID)) {
		_, _ = writer.Write(util.ReturnIdleBeat(model.FailureCode))
		e.log.Errorf("idleBeat任务[%d]正在运行", param.JobID)
		return
	}
	e.log.Infof("忙碌检测任务参数:%+v", param)
	_, _ = writer.Write(util.ReturnGeneral())
}

//注册执行器到调度中心
func (e *executor) registry() {

	t := time.NewTimer(time.Second * 0) //初始立即执行
	defer t.Stop()
	req := &model.Registry{
		RegistryGroup: "EXECUTOR",
		RegistryKey:   e.opts.RegistryKey,
		RegistryValue: "http://" + e.address,
	}
	param, err := json.Marshal(req)
	if err != nil {
		log.Fatal("执行器注册信息解析失败:" + err.Error())
	}
	for {
		<-t.C
		t.Reset(time.Second * time.Duration(20)) //20秒心跳防止过期
		func() {
			result, err := e.post("/api/registry", string(param))
			if err != nil {
				e.log.Errorf("执行器注册失败1:%s", err.Error())
				return
			}
			defer result.Body.Close()
			body, err := ioutil.ReadAll(result.Body)
			if err != nil {
				e.log.Errorf("执行器注册失败2:%s", err.Error())
				return
			}
			res := &model.Res{}
			_ = json.Unmarshal(body, &res)
			if res.Code != model.SuccessCode {
				e.log.Errorf("执行器注册失败3:%s", string(body))
				return
			}
			e.log.Infof("执行器注册成功:%s", string(body))	//这里从注册改成了心跳, log没改
		}()

	}
}

//执行器注册摘除
func (e *executor) registryRemove() {
	t := time.NewTimer(time.Second * 0) //初始立即执行
	defer t.Stop()
	req := &model.Registry{
		RegistryGroup: "EXECUTOR",
		RegistryKey:   e.opts.RegistryKey,
		RegistryValue: "http://" + e.address,
	}
	param, err := json.Marshal(req)
	if err != nil {
		e.log.Errorf("执行器摘除失败:%s", err.Error())
		return
	}
	res, err := e.post("/api/registryRemove", string(param))
	if err != nil {
		e.log.Errorf("执行器摘除失败:%s", err.Error())
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	e.log.Infof("执行器摘除成功:%s", string(body))
	_ = res.Body.Close()
}

//回调任务列表
func (e *executor) callback(task *Task, code int64, msg string) {
	e.runList.Del(util.Int64ToStr(task.Id))
	task.log.Close()
	res, err := e.post("/api/callback", string(util.ReturnCall(task.Param, code, msg)))
	if err != nil {
		e.log.Errorf("callback err : %s", err.Error())
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		e.log.Errorf("callback ReadAll err : %s", err.Error())
		return
	}
	e.log.Infof("任务回调成功: %s", string(body))
}

//post
func (e *executor) post(action, body string) (resp *http.Response, err error) {
	request, err := http.NewRequest("POST", e.opts.ServerAddr+action, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	request.Header.Set("XXL-JOB-ACCESS-TOKEN", e.opts.AccessToken)
	client := http.Client{
		Timeout: e.opts.Timeout,
	}
	return client.Do(request)
}

// RunTask 运行任务
func (e *executor) RunTask(writer http.ResponseWriter, request *http.Request) {
	e.runTask(writer, request)
}

// KillTask 删除任务
func (e *executor) KillTask(writer http.ResponseWriter, request *http.Request) {
	e.killTask(writer, request)
}

// TaskLog 任务日志
func (e *executor) TaskLog(writer http.ResponseWriter, request *http.Request) {
	e.taskLog(writer, request)
}

// Beat 心跳检测
func (e *executor) Beat(writer http.ResponseWriter, request *http.Request) {
	e.beat(writer, request)
}

// IdleBeat 忙碌检测
func (e *executor) IdleBeat(writer http.ResponseWriter, request *http.Request) {
	e.idleBeat(writer, request)
}

