package util

import (
	"encoding/json"
	"executor-go/model"
	"strconv"
)

// Int64ToStr int64 to str
func Int64ToStr(i int64) string {
	return strconv.FormatInt(i, 10)
}

//执行任务回调
func ReturnCall(req *model.RunReq, code int64, msg string) []byte {
	data := model.Call{
		&model.CallElement{
			LogID:      req.LogID,
			LogDateTim: req.LogDateTime,
			ExecuteResult: &model.ExecuteResult{
				Code: code,
				Msg:  msg,
			},
			HandleCode: int(code),
			HandleMsg:  msg,
		},
	}
	str, _ := json.Marshal(data)
	return str
}

//杀死任务返回
func ReturnKill(req *model.KillReq, code int64) []byte {
	msg := ""
	if code != model.SuccessCode {
		msg = "Task kill err"
	}
	data := model.Res{
		Code: code,
		Msg:  msg,
	}
	str, _ := json.Marshal(data)
	return str
}

//忙碌返回
func ReturnIdleBeat(code int64) []byte {
	msg := ""
	if code != model.SuccessCode {
		msg = "Task is busy"
	}
	data := model.Res{
		Code: code,
		Msg:  msg,
	}
	str, _ := json.Marshal(data)
	return str
}

//通用返回
func ReturnGeneral() []byte {
	data := &model.Res{
		Code: model.SuccessCode,
		Msg:  "",
	}
	str, _ := json.Marshal(data)
	return str
}
