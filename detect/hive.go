package detect

import (
	"code.byted.org/dp/gotqs"
	"code.byted.org/dp/gotqs/client"
	"code.byted.org/gopkg/logs"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/json"
	"net/http"
	"reflect"
	"time"
)
var ctx = context.Background()
func HiveQuery(c *gin.Context){
	type initArgs struct {
		ctx      context.Context
		appId    string
		appKey   string
		userName string
		timeout  time.Duration
		cluster  string
	}
	initAg := initArgs{
		ctx:      ctx,
		appId:    "BiQpawVCgDlNDorFLkQbxtEWdpWzf0mPI1WgltFC2fXBSD3P",
		appKey:   "VTStyv3g7fSS2mkiIBjYACxCPcW20UfIEGdstuC6xVGdCauS",
		userName: "kanghuaisong",
		timeout:  time.Minute * 20,
		cluster:  "hibis",
	}
	tqsClient, err := gotqs.MakeTqsClient(initAg.ctx, initAg.appId, initAg.appKey, initAg.userName, initAg.timeout, initAg.cluster)
	if err != nil {
		fmt.Println("MakeTqsClient error, err: %v", err)
		return
	}
	type args struct {
		ctx       context.Context
		tqsClient *client.TqsClient
		query     string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   *client.JobPreview
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				ctx:       ctx,
				tqsClient: tqsClient,
				query: `SELECT
    					device_id
						FROM
						origin_log.event_log_hourly
						where
						date > "20190409"
    					and date < "20190411"
    					and app = "news_article"
    					AND os_name='ios'
    					AND event='test_invitation_tfapp_check'
    					and params ['install_testflight'] = "1" limit 50`,
			},
			wantErr: false,
		},
	}
	got, got1, err := gotqs.SyncQuery(tests[0].args.ctx, tqsClient, tests[0].args.query)
	if (err != nil) != tests[0].wantErr {
		logs.Info("SyncQuery() error = %v, wantErr %v", err, tests[0].wantErr)
		return
	}
	if got != tests[0].want {
		logs.Info("SyncQuery() got = %v, want %v", got, tests[0].want)
	}
	if !reflect.DeepEqual(got1, tests[0].want1) {
		logs.Info("SyncQuery() got1 = %v, want %v", got1, tests[0].want1)
	}
	PrintGotJson(got1)
	time.Sleep(time.Second)
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : "success",
	})
}
func PrintGotJson(got interface{}) {
	jsGot, _ := json.Marshal(got)
	logs.Info("got: %v", string(jsGot))
}
