package env

import "sync/atomic"

const (
	// UnknownRegion .
	UnknownRegion = "-"
	R_CN          = "CN"
	R_SG          = "SG"
	R_US          = "US"
	R_MALIVA      = "MALIVA"
	R_ALISG       = "ALISG" // Singapore Aliyun
	R_CA          = "CA"    // West America
	R_BOE         = "BOE"
)

var (
	region     atomic.Value
	regionIDCs = map[string][]string{
		R_CN:     []string{DC_HY, DC_LF, DC_HL},
		R_SG:     []string{DC_SG},
		R_US:     []string{DC_VA},
		R_MALIVA: []string{DC_MALIVA},
		R_CA:     []string{DC_CA},
		R_ALISG:  []string{DC_ALISG},
		R_BOE:    []string{DC_BOE},
	}
)

// Region .
func Region() string {
	if v := region.Load(); v != nil {
		return v.(string)
	}

	idc := IDC()
	regionResult := GetRegionFromIDC(idc)
	region.Store(regionResult)
	return regionResult
}

func GetRegionFromIDC(idc string) string {
	for r, idcs := range regionIDCs {
		for _, dc := range idcs {
			if idc == dc {
				return r
			}
		}
	}

	return UnknownRegion
}
