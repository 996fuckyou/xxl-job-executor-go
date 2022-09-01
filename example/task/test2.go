package task

import (
	"context"
	"executor-go/joblog"
	"executor-go/model"
	"time"
)

func Test2(cxt context.Context, param *model.RunReq, log *joblog.JobLog) (msg string) {
	num := 1
	for {

		select {
		case <-cxt.Done():
			log.Info("task" + param.ExecutorHandler + "被手动终止")
			return
		default:
			num++
			time.Sleep(10 * time.Second)
			log.Info("test one task"+param.ExecutorHandler+" param："+param.ExecutorParams+"执行行", num)
			if num > 10 {
				log.Info("test one task" + param.ExecutorHandler + " param：" + param.ExecutorParams + "执行完毕！")
				return
			}
		}
	}

}
