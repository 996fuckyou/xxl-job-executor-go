package task

import (
	"context"
	"executor-go/joblog"
	"executor-go/model"

)

func Panic(cxt context.Context, param *model.RunReq, log *joblog.JobLog) (msg string) {
	panic("test panic")
	return
}
