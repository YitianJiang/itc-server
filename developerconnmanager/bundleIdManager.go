package developerconnmanager

import (
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

func Test64DecodeToString(c *gin.Context){
	logs.Info("开始测试pp文件content转化成bin文件")
	ppContentString := "MIIhqgYJKoZIhvcNAQcCoIIhmzCCIZcCAQExCzAJBgUrDgMCGgUAMIIRPgYJKoZIhvcNAQcBoIIRLwSCESs8P3htbCB2ZXJzaW9uPSIxLjAiIGVuY29kaW5nPSJVVEYtOCI/Pgo8IURPQ1RZUEUgcGxpc3QgUFVCTElDICItLy9BcHBsZS8vRFREIFBMSVNUIDEuMC8vRU4iICJodHRwOi8vd3d3LmFwcGxlLmNvbS9EVERzL1Byb3BlcnR5TGlzdC0xLjAuZHRkIj4KPHBsaXN0IHZlcnNpb249IjEuMCI+CjxkaWN0PgoJPGtleT5BcHBJRE5hbWU8L2tleT4KCTxzdHJpbmc+TmV3c1NvY2lhbDwvc3RyaW5nPgoJPGtleT5BcHBsaWNhdGlvbklkZW50aWZpZXJQcmVmaXg8L2tleT4KCTxhcnJheT4KCTxzdHJpbmc+VTlKRVk2Nk42QTwvc3RyaW5nPgoJPC9hcnJheT4KCTxrZXk+Q3JlYXRpb25EYXRlPC9rZXk+Cgk8ZGF0ZT4yMDE5LTA1LTEwVDExOjM3OjE4WjwvZGF0ZT4KCTxrZXk+UGxhdGZvcm08L2tleT4KCTxhcnJheT4KCQk8c3RyaW5nPmlPUzwvc3RyaW5nPgoJPC9hcnJheT4KCTxrZXk+SXNYY29kZU1hbmFnZWQ8L2tleT4KCTxmYWxzZS8+Cgk8a2V5PkRldmVsb3BlckNlcnRpZmljYXRlczwva2V5PgoJPGFycmF5PgoJCTxkYXRhPk1JSUZ6ekNDQkxlZ0F3SUJBZ0lJRVQzUUJrV0ZHVjB3RFFZSktvWklodmNOQVFFTEJRQXdnWll4Q3pBSkJnTlZCQVlUQWxWVE1STXdFUVlEVlFRS0RBcEJjSEJzWlNCSmJtTXVNU3d3S2dZRFZRUUxEQ05CY0hCc1pTQlhiM0pzWkhkcFpHVWdSR1YyWld4dmNHVnlJRkpsYkdGMGFXOXVjekZFTUVJR0ExVUVBd3c3UVhCd2JHVWdWMjl5YkdSM2FXUmxJRVJsZG1Wc2IzQmxjaUJTWld4aGRHbHZibk1nUTJWeWRHbG1hV05oZEdsdmJpQkJkWFJvYjNKcGRIa3dIaGNOTVRrd05ERTRNRGN6TlRNeFdoY05NakF3TkRFM01EY3pOVE14V2pDQndqRWFNQmdHQ2dtU0pvbVQ4aXhrQVFFTUNsVTVTa1ZaTmpaT05rRXhVVEJQQmdOVkJBTU1TR2xRYUc5dVpTQkVhWE4wY21saWRYUnBiMjQ2SUVKbGFXcHBibWNnUW5sMFpXUmhibU5sSUZSbFkyaHViMnh2WjNrZ1EyOHVMQ0JNZEdRdUlDaFZPVXBGV1RZMlRqWkJLVEVUTUJFR0ExVUVDd3dLVlRsS1JWazJOazQyUVRFdk1DMEdBMVVFQ2d3bVFtVnBhbWx1WnlCQ2VYUmxaR0Z1WTJVZ1ZHVmphRzV2Ykc5bmVTQkRieTRzSUV4MFpDNHhDekFKQmdOVkJBWVRBa05PTUlJQklqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FROEFNSUlCQ2dLQ0FRRUFzMk1hOG5qT1RaZ0pORFdtdTgwcTd6OXQ5bjFkVmZvc2hsaHBmbUMrK3dCbE5LUVpnelJGN1hWcjJweXBqQjgveHN3Rk4yTXkxcGJrWnpDKzRydDRWUEFaak5COTFtTFBpWFowc25BYlMwemkvNSt5MmJQTTJiK01hRllLWlFyUGxONktDSlVNYXFFa0tOZ09ab3VaTm11dUdITXNhR1RkdVZFKzFvN3orNVpJblVHM1RsYVBLNUdoSXZVdDBzc1orSjd5enY3Zk9KWHloRFRiZVdERlRLQXVTNmFCVGpXRy9saUtyWGpIcHVTRnZodzhBS25vdnZRL1JGN00wdlR3L2V6cldrTGc0enUzT1dYeG9iVHBNaW9pOSszc3dtMG4yYmhGZCt0RFNLdlVMbnY1bGZxQXhRSTV0YWJhcUJCd1E4eUpXcmh1RndPdis0VjFmY1I0aVFJREFRQUJvNElCOFRDQ0FlMHdEQVlEVlIwVEFRSC9CQUl3QURBZkJnTlZIU01FR0RBV2dCU0lKeGNKcWJZWVlJdnM2N3IyUjFuRlVsU2p0ekEvQmdnckJnRUZCUWNCQVFRek1ERXdMd1lJS3dZQkJRVUhNQUdHSTJoMGRIQTZMeTl2WTNOd0xtRndjR3hsTG1OdmJTOXZZM053TURNdGQzZGtjakV4TUlJQkhRWURWUjBnQklJQkZEQ0NBUkF3Z2dFTUJna3Foa2lHOTJOa0JRRXdnZjR3Z2NNR0NDc0dBUVVGQndJQ01JRzJESUd6VW1Wc2FXRnVZMlVnYjI0Z2RHaHBjeUJqWlhKMGFXWnBZMkYwWlNCaWVTQmhibmtnY0dGeWRIa2dZWE56ZFcxbGN5QmhZMk5sY0hSaGJtTmxJRzltSUhSb1pTQjBhR1Z1SUdGd2NHeHBZMkZpYkdVZ2MzUmhibVJoY21RZ2RHVnliWE1nWVc1a0lHTnZibVJwZEdsdmJuTWdiMllnZFhObExDQmpaWEowYVdacFkyRjBaU0J3YjJ4cFkza2dZVzVrSUdObGNuUnBabWxqWVhScGIyNGdjSEpoWTNScFkyVWdjM1JoZEdWdFpXNTBjeTR3TmdZSUt3WUJCUVVIQWdFV0ttaDBkSEE2THk5M2QzY3VZWEJ3YkdVdVkyOXRMMk5sY25ScFptbGpZWFJsWVhWMGFHOXlhWFI1THpBV0JnTlZIU1VCQWY4RUREQUtCZ2dyQmdFRkJRY0RBekFkQmdOVkhRNEVGZ1FVRnhFNEZDaSt2RG5TWU9SU2dZd2t5Y1haM1Jvd0RnWURWUjBQQVFIL0JBUURBZ2VBTUJNR0NpcUdTSWIzWTJRR0FRUUJBZjhFQWdVQU1BMEdDU3FHU0liM0RRRUJDd1VBQTRJQkFRQ2puY3FycjBNRk13UWxFRVVqS0NCUWl3Z2JLQThLcldwbmtHc2NKY0Y4dXJxaHJJRzhpZXJ5WTZlV0VmY3pPUFRXYzhpbXNCd3Q3cTJ2anFFNk5xT2F0KzIrVmhtSnY4ZHY3anhUSitnQnBQYnkxN2s0Z1A1azNwcklFVGhzWlQ0Y3FPcEJOM05iWWdTQ1BCWkxTYWpLR0hWTTV2SXZ0aU02VkhvNml4a3ZaNDdmVlpEUzh0ek5CNEFBQkM3YmM1NzArUWVyYitFNXZkL0ZHOHBXVkN2cUZsZGRHOUc2MFpNYXcrUnZTajYzRnkrUUtML3ZnQ2themdRUks1a1JoM1gvVTFSS1lXck45NWUrandGSzVuQlI5VXVZYUZqbzZ0Y1pUcjZUY1RadStYaWdKWjhqSzZNQ0xvbEM4T2NYNXpnVS96bm5kejc2b2FidDljaXQ3QTlVPC9kYXRhPgoJPC9hcnJheT4KCgkJCQkJCQkJCQkJCQkJCQkJCQkJCQkJCQkJCQkJCQkJCQkJCQkJCQkJCQkJCQkJCQkJCQkJCQkJCQkJCQoJPGtleT5FbnRpdGxlbWVudHM8L2tleT4KCTxkaWN0PgoJCTxrZXk+YmV0YS1yZXBvcnRzLWFjdGl2ZTwva2V5PgoJCTx0cnVlLz4KCQkJCQkJPGtleT5hcHMtZW52aXJvbm1lbnQ8L2tleT4KCQk8c3RyaW5nPnByb2R1Y3Rpb248L3N0cmluZz4KCQkJCQkJPGtleT5jb20uYXBwbGUuZGV2ZWxvcGVyLnViaXF1aXR5LWt2c3RvcmUtaWRlbnRpZmllcjwva2V5PgoJCTxzdHJpbmc+VTlKRVk2Nk42QS4qPC9zdHJpbmc+CgkJCQkJCTxrZXk+Y29tLmFwcGxlLmRldmVsb3Blci5pY2xvdWQtc2VydmljZXM8L2tleT4KCQk8c3RyaW5nPio8L3N0cmluZz4KCQkJCQkJPGtleT5jb20uYXBwbGUuZGV2ZWxvcGVyLmljbG91ZC1jb250YWluZXItaWRlbnRpZmllcnM8L2tleT4KCQk8YXJyYXk+PC9hcnJheT4KCQkJCQkJPGtleT5jb20uYXBwbGUuZGV2ZWxvcGVyLmljbG91ZC1jb250YWluZXItZGV2ZWxvcG1lbnQtY29udGFpbmVyLWlkZW50aWZpZXJzPC9rZXk+CgkJPGFycmF5PjwvYXJyYXk+CgkJCQkJCTxrZXk+Y29tLmFwcGxlLmRldmVsb3Blci51YmlxdWl0eS1jb250YWluZXItaWRlbnRpZmllcnM8L2tleT4KCQk8YXJyYXk+PC9hcnJheT4KCQkJCQkJPGtleT5jb20uYXBwbGUuZGV2ZWxvcGVyLmFzc29jaWF0ZWQtZG9tYWluczwva2V5PgoJCTxzdHJpbmc+Kjwvc3RyaW5nPgoJCQkJCQk8a2V5PmNvbS5hcHBsZS5kZXZlbG9wZXIubmV0d29ya2luZy53aWZpLWluZm88L2tleT4KCQk8dHJ1ZS8+CgkJCQkJCTxrZXk+Y29tLmFwcGxlLmRldmVsb3Blci5zaXJpPC9rZXk+CgkJPHRydWUvPgoJCQkJCQk8a2V5PmNvbS5hcHBsZS5kZXZlbG9wZXIucGFzcy10eXBlLWlkZW50aWZpZXJzPC9rZXk+CgkJPGFycmF5PgoJCQkJPHN0cmluZz5VOUpFWTY2TjZBLio8L3N0cmluZz4KCQk8L2FycmF5PgoJCQkJCQk8a2V5PmNvbS5hcHBsZS5zZWN1cml0eS5hcHBsaWNhdGlvbi1ncm91cHM8L2tleT4KCQk8YXJyYXk+CgkJCQk8c3RyaW5nPmdyb3VwLnRvZGF5RXh0ZW5zdGlvblNoYXJlRGVmYXVsdHM8L3N0cmluZz4KCQk8L2FycmF5PgoJCQkJCQk8a2V5PmFwcGxpY2F0aW9uLWlkZW50aWZpZXI8L2tleT4KCQk8c3RyaW5nPlU5SkVZNjZONkEuY29tLnNzLmlwaG9uZS5hcnRpY2xlLk5ld3NTb2NpYWw8L3N0cmluZz4KCQkJCQkJPGtleT5rZXljaGFpbi1hY2Nlc3MtZ3JvdXBzPC9rZXk+CgkJPGFycmF5PgoJCQkJPHN0cmluZz5VOUpFWTY2TjZBLio8L3N0cmluZz4KCQk8L2FycmF5PgoJCQkJCQk8a2V5PmdldC10YXNrLWFsbG93PC9rZXk+CgkJPGZhbHNlLz4KCQkJCQkJPGtleT5jb20uYXBwbGUuZGV2ZWxvcGVyLnRlYW0taWRlbnRpZmllcjwva2V5PgoJCTxzdHJpbmc+VTlKRVk2Nk42QTwvc3RyaW5nPgoKCTwvZGljdD4KCTxrZXk+RXhwaXJhdGlvbkRhdGU8L2tleT4KCTxkYXRlPjIwMjAtMDQtMTdUMDc6MzU6MzFaPC9kYXRlPgoJPGtleT5OYW1lPC9rZXk+Cgk8c3RyaW5nPk5ld3NTb2NpYWxEaXN0PC9zdHJpbmc+Cgk8a2V5PlRlYW1JZGVudGlmaWVyPC9rZXk+Cgk8YXJyYXk+CgkJPHN0cmluZz5VOUpFWTY2TjZBPC9zdHJpbmc+Cgk8L2FycmF5PgoJPGtleT5UZWFtTmFtZTwva2V5PgoJPHN0cmluZz5CZWlqaW5nIEJ5dGVkYW5jZSBUZWNobm9sb2d5IENvLiwgTHRkLjwvc3RyaW5nPgoJPGtleT5UaW1lVG9MaXZlPC9rZXk+Cgk8aW50ZWdlcj4zNDI8L2ludGVnZXI+Cgk8a2V5PlVVSUQ8L2tleT4KCTxzdHJpbmc+ZjgwYjMwNzktN2MwYi00ODZiLThmZmQtM2U0Y2YwZDc2NjI5PC9zdHJpbmc+Cgk8a2V5PlZlcnNpb248L2tleT4KCTxpbnRlZ2VyPjE8L2ludGVnZXI+CjwvZGljdD4KPC9wbGlzdD6ggg2xMIID8zCCAtugAwIBAgIBFzANBgkqhkiG9w0BAQUFADBiMQswCQYD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"

	decoded, err := base64.StdEncoding.DecodeString(ppContentString)
	if err != nil{
		logs.Info("decode64位加密失败，pp文件content内容格式错误")
		c.JSON(http.StatusOK, gin.H{
			"message":   "error",
			"errorCode": 1,
			"data": "decode64位加密失败，pp文件content内容格式错误",
		})
	}else {
		err := ioutil.WriteFile("NewsSocial.mobileprovision", decoded, 0666)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"message":   "error",
				"errorCode": 2,
				"data": "文件写入失败",
			})
		}else {
			c.JSON(http.StatusOK, gin.H{
				"message":   "success",
				"errorCode": 0,
				"data": "文件成功生成",
			})
		}
	}
}

//func TestAskBundleId(c *gin.Context)  {
//	logs.Info("开始测试Bundle ID逻辑")
//	nameBundle := c.Query("nameBundle")
//	var bundleIdObj dal.BundleIdManager
//	bundleIdObj.BundleId = nameBundle
//
//	if dal.InsertBundleId(bundleIdObj) {
//		c.JSON(http.StatusOK, gin.H{
//			"message":   "success",
//			"errorCode": 0,
//			"data": "插入BundleID正常",
//			"name": nameBundle,
//		})
//	} else {
//		c.JSON(http.StatusBadRequest, gin.H{
//			"errorCode": -1,
//			"message":   "数据库存储错误，新建BundleID失败！",
//		})
//	}
//}

func GetBunldIdsObj(c *gin.Context){
	logs.Info("返回数据库中所有的Bundle ID")
	BundleIdsObjResponse,boolResult := dal.SearchBundleIds()
	if boolResult{
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"errorCode": 0,
			"data": BundleIdsObjResponse,
		})
	}else {
		c.JSON(http.StatusOK, gin.H{
			"message":   "fail DB",
			"errorCode": 1,
		})
	}
}

func CreateP8DBInfoToTable(c *gin.Context) {
	logs.Info("接收P8文件，并转化string字符串存入DB")
	nameBundle := c.PostForm("nameBundle")
	file, header, _ := c.Request.FormFile("P8file")
	logs.Info("打印File Name：" + header.Filename)
	p8ByteInfo,err := ioutil.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message":   "error read p8 file",
			"errorCode": 0,
			"data": err,
		})
		return
	}
	var p8info dal.P8fileManager
	p8info.BundleId = nameBundle
	p8info.P8StringInfo = string(p8ByteInfo)
	if dal.InsertBundleIdAndP8(p8info) {
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"errorCode": 0,
			"data": "插入BundleID正常",
			"name": nameBundle,
			"p8infostring": string(p8ByteInfo),
		})
		return
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode": -1,
			"message":   "数据库存储错误，新建BundleID失败！",
		})
		return
	}
}

func ParsePrivateKey(c *gin.Context) {
	tokenString := GetTokenString()
	if tokenString == "not find p8String"{
		c.JSON(http.StatusOK, gin.H{
			"message":   "error",
			"errorCode": 1,
			"data": tokenString,
		})
	}else {
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"errorCode": 0,
			"data": tokenString,
		})
	}
}