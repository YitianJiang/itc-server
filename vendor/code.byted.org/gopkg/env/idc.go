package env

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
)

const (
	UnknownIDC = "-"
	DC_HY      = "hy"
	DC_LF      = "lf"
	DC_HL      = "hl"
	DC_WJ      = "wj"
	DC_VA      = "va"
	DC_SG      = "sg"
	DC_BR      = "br"
	DC_JP      = "jp"
	DC_IN      = "in"
	DC_CA      = "ca" // West America
	DC_FRAWS   = "fraws"
	DC_ALISG   = "alisg" // Singapore Aliyun
	DC_ALIVA   = "aliva"
	DC_MALIVA  = "maliva"
	DC_ALINC2  = "alinc2" //aliyun north
	DC_MAWSJP  = "mawsjp" //musical.ly 东京老机房
	DC_BOE     = "boe"    // bytedance offline environment
)

var (
	idc       atomic.Value
	idcPrefix = map[string][]string{
		DC_HY:     {"10.4."},
		DC_LF:     {"10.2.", "10.3.", "10.6.", "10.8.", "10.9.", "10.10.", "10.11.", "10.12.", "10.13.", "10.14.", "10.15.", "10.16.", "10.17.", "10.18.", "10.1."},
		DC_HL:     {"10.20.", "10.21.", "10.22.", "10.23.", "10.24.", "10.25."},
		DC_VA:     {"10.100."},
		DC_SG:     {"10.101."},
		DC_CA:     {"10.106."},
		DC_ALISG:  {"10.115."},
		DC_ALIVA:  {},
		DC_MALIVA: {"10.110.0.", "10.110.255."},
		DC_ALINC2: {"10.108"},
		DC_MAWSJP: {},
		DC_WJ:     {"10.31."},
		DC_BR:     {"10.102."},
		DC_JP:     {"10.103."},
		DC_IN:     {"10.109."},
		DC_FRAWS:  {"10.105."},
		DC_BOE:    {"10.225."},
	}
	FixedIDCList = []string{ // NOTE: new added idc must be append to the end
		UnknownIDC, DC_HY, DC_LF, DC_HL, DC_VA, DC_SG, DC_CA, DC_ALISG, DC_ALIVA, DC_MALIVA, DC_ALINC2, DC_MAWSJP,
		DC_WJ, DC_JP, DC_IN, DC_FRAWS, DC_BOE,
	}
)

// IDC .
func IDC() string {
	if v := idc.Load(); v != nil {
		return v.(string)
	}

	if dc := os.Getenv("RUNTIME_IDC_NAME"); dc != "" {
		idc.Store(dc)
		return dc
	}

	b, err := ioutil.ReadFile("/opt/tmp/consul_agent/datacenter")
	if err == nil {
		if dc := strings.TrimSpace(string(b)); dc != "" {
			idc.Store(dc)
			return dc
		}
	}

	cmd0 := exec.Command("/opt/tiger/consul_deploy/bin/determine_dc.sh")
	output0, err := cmd0.Output()
	if err == nil {
		dc := strings.TrimSpace(string(output0))
		if _, ok := idcPrefix[dc]; ok {
			idc.Store(dc)
			return dc
		}
	}

	cmd := exec.Command(`bash`, `-c`, `sd report|grep "Data center"|awk '{print $3}'`)
	output, err := cmd.Output()
	if err == nil {
		dc := strings.TrimSpace(string(output))
		if _, ok := idcPrefix[dc]; ok {
			idc.Store(dc)
			return dc
		}
	}

	idc.Store(UnknownIDC)
	return UnknownIDC
}

func GetIDCFromHost(ip string) string {
	for dc, prefixes := range idcPrefix {
		for _, prefix := range prefixes {
			if strings.HasPrefix(ip, prefix) {
				return dc
			}
		}
	}
	return UnknownIDC
}

func GetIDCList() []string {
	idcList := []string{}
	for key, _ := range idcPrefix {
		idcList = append(idcList, key)
	}
	return idcList
}

func SetIDC(v string) {
	idc.Store(v)
}
