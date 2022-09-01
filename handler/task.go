package handler

import (
	"context"
	"executor-go/joblog"
	"executor-go/model"
	"fmt"
	"runtime/debug"
)

// TaskFunc 任务执行函数
type TaskFunc func(cxt context.Context, param *model.RunReq, log *joblog.JobLog) string

// Task 任务
type Task struct {
	Id        int64
	Name      string
	Ext       context.Context
	Param     *model.RunReq
	fn        TaskFunc
	Cancel    context.CancelFunc
	StartTime int64
	EndTime   int64
	//日志
	log *joblog.JobLog
}

// Run 运行任务
func (t *Task) Run(callback func(code int64, msg string)) {
	defer func(cancel func()) {
		if err := recover(); err != nil {
			t.log.Info(t.Info()+" panic: %v", err)
			debug.PrintStack() //堆栈跟踪
			callback(model.FailureCode, fmt.Sprintf("task panic:%v", err))
			cancel()
		}
	}(t.Cancel)
	msg := t.fn(t.Ext, t.Param, t.log)
	callback(model.SuccessCode, msg)
	return
}

// Info 任务信息
func (t *Task) Info() string {
	return fmt.Sprintf("任务ID[%d]任务名称[%s]参数:%s", t.Id, t.Name, t.Param.ExecutorParams)
}
