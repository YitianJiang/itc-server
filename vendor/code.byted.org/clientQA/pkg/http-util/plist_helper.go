package http_util

import (
	"code.byted.org/clientQA/pkg/const"
	"code.byted.org/gopkg/logs"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

/*
使用pkgUrl生成plist文件
*/
func CreateDownloadUrl(pkgUrl string, branch string, bundleID string, content string, proj string) string {
	bytes := []byte(branch + bundleID)
	h := md5.New()
	h.Write(bytes)
	branch_md5 := hex.EncodeToString(h.Sum(nil))

	bytesProj := []byte(proj)
	h.Reset()
	h.Write(bytesProj)
	projMd5 := hex.EncodeToString(h.Sum(nil))

	resp, err := http.PostForm(_const.Plist_url, url.Values{"job": {projMd5 + branch_md5}, "durl": {pkgUrl}, "bundleid": {bundleID}, "displaname": {content}})
	if err != nil {
		logs.Error("%v", "generate plist failed")
		return ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("%v", "read plist failed")
		return ""
	}

	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err == nil {
		logs.Debug("%v", "==============json str 转map=======================")
		logs.Debug("%v", dat)
		logs.Debug("%v", dat["host"])
	}

	durl := fmt.Sprint(dat["data"])
	logs.Debug("%v", durl)
	return "itms-services://?action=download-manifest&url=" + durl
}
