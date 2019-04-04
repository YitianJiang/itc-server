gotqs
=======
[TQS](http://tqs.byted.org) go SDK

## 使用指南
### 1. 申请app
在tqs网站申请app，得到appId/appKey/userName
### 2. 拉取代码
```bash
cd $GOPATH/src/code.byted.org && git clone git@code.byted.org:dp/gotqs.git
```
### 3. 创建tqs client
```go
tqsClient, err := MakeTqsClient(ctx, appId, appKey, userName, timeout, cluster)
```
`appId`/`appKey`/`userName`为步骤1申请得到的。
`timeout`为程序执行同步查询的超时时间，超时SyncQuery函数会返回error。
`cluster`参考`consts/const.go`:
```
//tos api cluster
const (
	CLUSTER_TEST      = "test"
	CLUSTER_DEFAULT   = "default"
	CLUSTER_AWS       = "va_aws"
	CLUSTER_MALIVA    = "maliva"
	CLUSTER_VA_TEST   = "va_test"
	CLUSTER_CN_PRIEST = "cn_priest"
	CLUSTER_VA_PRIEST = "va_priest"
)
const (
	TQS_API_SERVER_PSM_TEST      = "data.olap.tqs_test.service.lf.byted.org"
	TQS_API_SERVER_PSM_CN        = "data.olap.tqs.service.lf.byted.org"
	TQS_API_SERVER_PSM_AWS       = "data.olap.tqs_va_aws.service.maliva.byted.org"
	TQS_API_SERVER_PSM_TEST_VA   = "data.olap.tqs_test_va.service.maliva.byted.org"
	TQS_API_SERVER_PSM_CN_PRIEST = "data.olap.tqs_cn_priest.service.lf.byted.org"
	TQS_API_SERVER_PSM_VA_PRIEST = "data.olap.tqs_va_priest.service.maliva.byted.org"
)
```

### 4. 执行查询
```go
status, jobPreview, err := SyncQuery(ctx, tqsClient, hql)
```
`status`为任务最终状态，含义见gotqs.StatusMsg。
`jobPreview`为任务预览数据，数据结构如下：
```go
type JobPreview struct{
	Rows [][]string `json:"rows"`
	Url string `json:"url"`
}
```
其中：`Rows`为任务预览数据，`Url`为结果下载地址。

具体使用可以参考`tqs_test.go`。
