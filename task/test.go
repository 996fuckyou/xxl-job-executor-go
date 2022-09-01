package task

import (
	"context"
	"executor-go/joblog"
	"executor-go/model"
	"executor-go/util"
)

func Test(cxt context.Context, param *model.RunReq, log *joblog.JobLog) (msg string) {
	log.Info("test one task" + param.ExecutorHandler + " param：" + param.ExecutorParams + " log_id:" + util.Int64ToStr(param.LogID))
	return "test done"
}
