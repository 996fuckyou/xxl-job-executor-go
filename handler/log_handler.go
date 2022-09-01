package handler

import (
	"encoding/json"
	"net/http"

	"executor-go/model"
)

/**
用来日志查询，显示到xxl-job-admin后台
*/

/*
和java执行器类似
每一个jobID都有一个log, 生命周期30天
通过配置文件决定log放置位置
每一次启动job将log写入log文件中
*/

type LogHandler func(req *model.LogReq) *model.LogRes

//默认返回
func defaultLogHandler(req *model.LogReq) *model.LogRes {
	return &model.LogRes{Code: model.SuccessCode, Msg: "", Content: model.LogResContent{
		FromLineNum: req.FromLineNum,
		ToLineNum:   2,
		LogContent:  "这是日志默认返回，说明没有设置LogHandler",
		IsEnd:       true,
	}}
}

//请求错误
func reqErrLogHandler(w http.ResponseWriter, req *model.LogReq, err error) {
	res := &model.LogRes{Code: model.FailureCode, Msg: err.Error(), Content: model.LogResContent{
		FromLineNum: req.FromLineNum,
		ToLineNum:   0,
		LogContent:  err.Error(),
		IsEnd:       true,
	}}
	str, _ := json.Marshal(res)
	_, _ = w.Write(str)
}
