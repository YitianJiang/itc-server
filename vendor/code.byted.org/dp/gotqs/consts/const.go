package consts

const (
	TQS_API_SERVER_PSM_TEST      = "data.olap.tqs_test.service.lf.byted.org"
	TQS_API_SERVER_PSM_CN        = "data.olap.tqs.service.lf.byted.org"
	TQS_API_SERVER_PSM_AWS       = "data.olap.tqs_va_aws.service.maliva.byted.org"
	TQS_API_SERVER_PSM_TEST_VA   = "data.olap.tqs_test_va.service.maliva.byted.org"
	TQS_API_SERVER_PSM_CN_PRIEST = "data.olap.tqs_cn_priest.service.lf.byted.org"
	TQS_API_SERVER_PSM_VA_PRIEST = "data.olap.tqs_va_priest.service.maliva.byted.org"
)

const (
	STATUS_COMPLETE      = "Completed"
	STATUS_CANCEL        = "Cancelled"
	STATUS_ANALYSIS_FAIL = "AnalysisFailed"
	STATUS_FAIL          = "Failed"
)

//client init status
const (
	CLIENT_INIT_STATUS_INIT    = 0
	CLIENT_INIT_STATUS_SUCCESS = 1
	CLIENT_INIT_STATUS_FAIL    = 2
)

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
