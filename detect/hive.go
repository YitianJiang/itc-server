package detect

import (
	"code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/dp/gotqs"
	"code.byted.org/dp/gotqs/client"
	"code.byted.org/dp/gotqs/consts"
	"code.byted.org/gopkg/logs"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"time"
)
var ctx = context.Background()
/*func HiveQuery(c *gin.Context){

}*/

func InitCron(){
	logs.Info("init cron job")
	/*c := cron.New()
	spec := "0 5 15 * * ?"
	c.AddFunc(spec, cronJob)
	c.Start()*/
}

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
		ctx     :   ctx,
		appId	:   "BiQpawVCgDlNDorFLkQbxtEWdpWzf0mPI1WgltFC2fXBSD3P",
		appKey	:   "VTStyv3g7fSS2mkiIBjYACxCPcW20UfIEGdstuC6xVGdCauS",
		userName: 	"kanghuaisong",
		timeout	:   time.Minute * 30,
		//cluster	:   "hibis",
		cluster	:   consts.CLUSTER_CN_PRIEST,
	}
	timeEnd := time.Now()
	timeStart := timeEnd.AddDate(0, 0, -2)
	ts := timeStart.Format("20060102")
	te := timeEnd.Format("20060102")
	tqsClient, err := gotqs.MakeTqsClient(initAg.ctx, initAg.appId, initAg.appKey, initAg.userName, initAg.timeout, initAg.cluster)
	if err != nil {
		logs.Error("MakeTqsClient error, err: %v", err)
		return
	}
	type args struct {
		ctx       context.Context
		tqsClient *client.TqsClient
		query     string
	}
	query := []struct {
		name    string
		args    args
		want    string
		want1   *client.JobPreview
		wantErr bool
	}{
		{
			args: args{
			ctx			: ctx,
			tqsClient	: tqsClient,
			query		: `SELECT
    							device_id
							FROM  origin_log.event_log_hourly
							where
								date > "` + ts + `"
    							and date < "` + te + `"
    							and app = "news_article"
    							AND os_name='ios'
    							AND event='test_invitation_tfapp_check'
    							and params ['install_testflight'] = "1" limit 50`,
			},
			wantErr: false,
		},
	}
	got, got1, err := gotqs.SyncQuery(query[0].args.ctx, tqsClient, query[0].args.query)
	if (err != nil) != query[0].wantErr {
		fmt.Println("SyncQuery() error = %v, wantErr %v", err, query[0].wantErr)
		return
	}
	if got != query[0].want {
		fmt.Println("SyncQuery() got = %v, want %v", got, query[0].want)
	}
	if !reflect.DeepEqual(got1, query[0].want1) {
		fmt.Println("SyncQuery() got1 = %v, want %v", *got1, query[0].want1)
	}
	//PrintGotJson(t, got1)
	rows := (*got1).Rows
	//将数据存至数据库
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
	}
	defer connection.Close()
	db := connection.LogMode(_const.DB_LOG_MODE)
	if rows != nil {
		for i:=1; i<len(rows); i++ {
			//获取到did
			did := rows[i][0]
			sql := "insert ignore into tb_installed_tf (did) values ('" + did + "')"
			db.Exec(sql)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : "success",
	})
}
