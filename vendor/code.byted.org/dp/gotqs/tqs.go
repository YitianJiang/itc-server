package gotqs

import (
	"code.byted.org/dp/gotqs/client"
	"code.byted.org/dp/gotqs/consts"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/pkg/errors"
	"context"
	"fmt"
	"time"
)

type Job struct {
	JobStatus  *client.JobStatus
	JobPreview *client.JobPreview
	err        error
}

//创建tqs client
func MakeTqsClient(ctx context.Context, appId string, appKey string, userName string, timeout time.Duration, cluster string) (*client.TqsClient, error) {
	if len(appId) == 0 || len(appKey) == 0 {
		logs.CtxError(ctx, "appId/appKey empty")
		return nil, errors.New("param error")
	}
	if len(userName) == 0 {
		logs.CtxError(ctx, "userName empty")
		return nil, errors.New("param error")
	}
	tqsClient, err := client.InitClient(ctx, appId, appKey, userName, timeout, cluster)
	if err != nil || tqsClient == nil {
		logs.CtxError(ctx, "InitClient error, err: %v", err)
		return nil, errors.New("InitClient error")
	}
	return tqsClient, nil
}

//同步查询
//return : 任务Status/任务数据预览(含下载url)/error
func SyncQuery(ctx context.Context, tqsClient *client.TqsClient, query string) (string, *client.JobPreview, error) {
	if nil == tqsClient {
		logs.CtxError(ctx, "tqsClient nil, use MakeTqsClient() to create one")
		return "", nil, errors.New("param error")
	}
	if consts.CLIENT_INIT_STATUS_SUCCESS != tqsClient.InitStatus || tqsClient.ClientConf == nil {
		logs.CtxError(ctx, "tqsClient init invalid, use MakeTqsClient() to re create")
		return "", nil, errors.New("init error")
	}
	//创建查询任务
	jobId, err := tqsClient.CreateQueryJob(ctx, query, false, false)
	if err != nil {
		logs.CtxError(ctx, "CreateQueryJob error, err: %v", err)
		return "", nil, errors.New("CreateQueryJob error")
	}
	timeout := tqsClient.ClientConf.Timeout

	logs.CtxDebug(ctx, "CreateQueryJob success, jobId: %v, timeout: %v", jobId, timeout)

	done := make(chan Job)
	tickStatus := make(chan string)

	//goroutine loop query
	go func(jobId int64) {
		slpIntv := time.Second * 3
		for {
			//查询任务状态
			jobStatus, err := tqsClient.GetQueryJobInfo(ctx, jobId)
			if err != nil || jobStatus == nil {
				logs.CtxWarn(ctx, "GetQueryJobInfo error, err: %v", err)
				time.Sleep(GetSleepInterval(slpIntv))
				continue
			}

			if jobStatus.Status == consts.STATUS_COMPLETE { //done
				//获取JobPreview
				jobPreview, err := tqsClient.PreviewQuery(ctx, jobId)
				doneJob := Job{
					JobStatus:  jobStatus,
					JobPreview: jobPreview,
				}
				if err != nil {
					logs.CtxError(ctx, "PreviewQuery error, err: %v", err)
					doneJob.err = errors.New(fmt.Sprintf("PreviewQuery error, err: %v", err))
				}
				done <- doneJob
				return
			} else if jobStatus.Status == consts.STATUS_ANALYSIS_FAIL || jobStatus.Status == consts.STATUS_CANCEL || jobStatus.Status == consts.STATUS_FAIL {
				doneJob := Job{
					JobStatus: jobStatus,
				}
				logs.CtxWarn(ctx, "Job failed, status: %v", jobStatus.Status)
				done <- doneJob
				//不返回Job.err，返回Status让父进程处理
				return
			} else {
				tickStatus <- jobStatus.Status
			}
			time.Sleep(GetSleepInterval(slpIntv))
		}

	}(jobId)

	for {
		select {
		case <-time.After(timeout): //test ok
			logs.CtxError(ctx, "query timeout: %v", timeout)
			return "", nil, errors.New("query time out")
		case job := <-done:
			if job.err != nil {
				logs.CtxError(ctx, "query error, err: %v", err)
				return "", nil, errors.New("query error")
			}
			if job.JobStatus == nil {
				logs.CtxError(ctx, "query unexpected error")
				return "", nil, errors.New("unexpected error")
			}
			if job.JobStatus.Status != consts.STATUS_COMPLETE {
				logs.CtxError(ctx, "query failed, status: %v", job.JobStatus.Status)
				return job.JobStatus.Status, nil, errors.New("query failed")
			}
			if job.JobPreview == nil {
				logs.CtxError(ctx, "query unexpected error")
				return "", nil, errors.New("unexpected error")
			}
			return job.JobStatus.Status, job.JobPreview, nil

		case st := <-tickStatus:
			strStatus := GetJobStatus(st)
			logs.CtxNotice(ctx, "[JOB STATUS] %s", strStatus)
			continue
		}
	}

}

func GetSleepInterval(duration time.Duration) time.Duration {
	if duration <= 0 {
		return time.Second * 1
	}
	if duration < time.Minute*1 {
		duration = duration * 2
	} else {
		duration = duration + time.Minute*1
	}
	return duration
}

var StatusMsg = map[string][2]string{
	"Created":           {"查询任务已经创建成功", "API Server"},
	"AnalysisCompleted": {"查询任务解析成功", "API Server"},
	"AnalysisFailed":    {"查询任务解析失败（一般有语法或语义问题）", "API Server"},
	"Dispatched":        {"查询任务已经投递给调度器", "API Server"},
	"Scheduled":         {"查询任务已经进入调度器", "Scheduler"},
	"Pending":           {"查询任务进入处理队列（排队）", "Scheduler"},
	"Matched":           {"查询任务已经完成资源和调度匹配，即将被送入", "Worker Scheduler"},
	"Processing":        {"查询任务正在被worker处理（已经开始查询）", "Worker "},
	"Completed":         {"查询任务已完成（对于explain_query解析成功之后也会进入这个状态）", "Worker"},
	"Cancelled":         {"查询任务在执行过程中被取消", "Worker"},
	"Failed":            {"查询任务执行失败", "Worker"},
}

//获取任务状态
func GetJobStatus(status string) string {
	if item, ok := StatusMsg[status]; ok {
		return fmt.Sprintf("%s => %s [STAGE] %s", status, item[0], item[1])
	}
	return fmt.Sprintf("%s => %s [STAGE] %s", status, "unknown", "unknown")
}
