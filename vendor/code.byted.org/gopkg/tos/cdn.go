package tos

import (
	"crypto/md5"
	"sort"
	"strconv"
)

type ProductName string

/*
 * 访问国内的资源
 */
const (
	ToutiaoCDN    ProductName = "TT"  // 今日头条
	XiGuaCDN      ProductName = "XG"  // 西瓜视频
	HuoshanCDN    ProductName = "HS"  // 火山小视频
	WuKongCDN     ProductName = "WK"  // 悟空问答
	DouYinCDN     ProductName = "DY"  // 抖音
	FlipgramCDN   ProductName = "FG"  // Flipgram
	TuchongCDN    ProductName = "TC"  // 图虫
	NeiHanCDN     ProductName = "NH"  // 内涵段子
	GuaGuaLongCDN ProductName = "GGL" // 瓜瓜龙
	MayaCDN       ProductName = "MY"  // 快闪
	OtherCDN      ProductName = "TT"  // 其他(复用头条CDN)

	maxDomain = 3
)

/*
 * 访问SG的资源
 */
const (
	SGHypstarCDN ProductName = "HY" // Hypstar
	SGTickTokCDN ProductName = "TK" // TickTok
	SGTopBuzzCDN ProductName = "TB" // TopBuzz
)

/*
 * 访问VA的资源
 */
const (
	VAMusicallyCDN ProductName = "MS" // Musically
)

/*
使用说明：
一、必须指定当前请求上下文的业务，以便作CDN财务审计;
二、GetDomains 接口会根据 product & uri 返回一组域名，按顺序用作fallback使用；
三、访问路径为 http[s]?://{DOMAIN}/obj/{BUCKET}/{KEY}
*/
func GetDomains(product ProductName, uri string) []string {
	ret := make([]string, 0, maxDomain)
	v := md5uint64(uri)
	nodes := sfCDNDomains[product]
	if len(nodes) == 0 {
		return ret
	}
	idx := sort.Search(len(nodes), func(i int) bool { return nodes[i].Value >= v })
NEXT_NODE:
	for i := 0; i < 1000; i++ {
		name := nodes[(idx+i)%len(nodes)].Name
		for _, n := range ret {
			if n == name {
				continue NEXT_NODE
			}
		}
		ret = append(ret, name)
		if len(ret) == cap(ret) {
			return ret
		}
	}
	return ret
}

// 相关较大的文件专用域名 (平均20M以上)
func GetDomainsForLargeFile(product ProductName, uri string) []string {
	ret := make([]string, 0, maxDomain)
	v := md5uint64(uri)
	nodes := lfCDNDomains[product]
	if len(nodes) == 0 {
		return ret
	}
	idx := sort.Search(len(nodes), func(i int) bool { return nodes[i].Value >= v })
NEXT_NODE:
	for i := 0; i < 1000; i++ {
		name := nodes[(idx+i)%len(nodes)].Name
		for _, n := range ret {
			if n == name {
				continue NEXT_NODE
			}
		}
		ret = append(ret, name)
		if len(ret) == cap(ret) {
			return ret
		}
	}
	return ret
}

func GetSGDomains(product ProductName, uri string) []string {
	ret := make([]string, 0, maxDomain)
	v := md5uint64(uri)
	nodes := sfSGCDNDomains[product]
	if len(nodes) == 0 {
		return ret
	}
	idx := sort.Search(len(nodes), func(i int) bool { return nodes[i].Value >= v })
NEXT_NODE:
	for i := 0; i < 1000; i++ {
		name := nodes[(idx+i)%len(nodes)].Name
		for _, n := range ret {
			if n == name {
				continue NEXT_NODE
			}
		}
		ret = append(ret, name)
		if len(ret) == cap(ret) {
			return ret
		}
	}
	return ret
}

func GetSGDomainsForLargeFile(product ProductName, uri string) []string {
	ret := make([]string, 0, maxDomain)
	v := md5uint64(uri)
	nodes := lfSGCDNDomains[product]
	if len(nodes) == 0 {
		return ret
	}
	idx := sort.Search(len(nodes), func(i int) bool { return nodes[i].Value >= v })
NEXT_NODE:
	for i := 0; i < 1000; i++ {
		name := nodes[(idx+i)%len(nodes)].Name
		for _, n := range ret {
			if n == name {
				continue NEXT_NODE
			}
		}
		ret = append(ret, name)
		if len(ret) == cap(ret) {
			return ret
		}
	}
	return ret
}

func GetVADomains(product ProductName, uri string) []string {
	ret := make([]string, 0, maxDomain)
	v := md5uint64(uri)
	nodes := sfVACDNDomains[product]
	if len(nodes) == 0 {
		return ret
	}
	idx := sort.Search(len(nodes), func(i int) bool { return nodes[i].Value >= v })
NEXT_NODE:
	for i := 0; i < 1000; i++ {
		name := nodes[(idx+i)%len(nodes)].Name
		for _, n := range ret {
			if n == name {
				continue NEXT_NODE
			}
		}
		ret = append(ret, name)
		if len(ret) == cap(ret) {
			return ret
		}
	}
	return ret
}

func GetVADomainsForLargeFile(product ProductName, uri string) []string {
	ret := make([]string, 0, maxDomain)
	v := md5uint64(uri)
	nodes := lfVACDNDomains[product]
	if len(nodes) == 0 {
		return ret
	}
	idx := sort.Search(len(nodes), func(i int) bool { return nodes[i].Value >= v })
NEXT_NODE:
	for i := 0; i < 1000; i++ {
		name := nodes[(idx+i)%len(nodes)].Name
		for _, n := range ret {
			if n == name {
				continue NEXT_NODE
			}
		}
		ret = append(ret, name)
		if len(ret) == cap(ret) {
			return ret
		}
	}
	return ret
}

func md5uint64(s string) uint64 { // BigEndian
	b := md5.Sum([]byte(s))
	return uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
		uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
}

type consistentNode struct {
	Value uint64
	Name  string
}

// product => consistent hashing list
var sfCDNDomains map[ProductName][]consistentNode
var lfCDNDomains map[ProductName][]consistentNode
var sfSGCDNDomains map[ProductName][]consistentNode
var lfSGCDNDomains map[ProductName][]consistentNode
var sfVACDNDomains map[ProductName][]consistentNode
var lfVACDNDomains map[ProductName][]consistentNode

type domain struct {
	Product ProductName
	Name    string
	Weight  int
}

var sfCDNs = []domain{ // TODO: mv to etcd
	{ToutiaoCDN, "sf1-ttcdn-tos.pstatp.com", 100},
	{ToutiaoCDN, "sf3-ttcdn-tos.pstatp.com", 100},
	{ToutiaoCDN, "sf6-ttcdn-tos.pstatp.com", 100},
	{XiGuaCDN, "sf1-xgcdn-tos.pstatp.com", 100},
	{XiGuaCDN, "sf3-xgcdn-tos.pstatp.com", 100},
	{XiGuaCDN, "sf6-xgcdn-tos.pstatp.com", 100},
	{HuoshanCDN, "sf1-hscdn-tos.pstatp.com", 100},
	{HuoshanCDN, "sf3-hscdn-tos.pstatp.com", 100},
	{HuoshanCDN, "sf6-hscdn-tos.pstatp.com", 100},
	{WuKongCDN, "sf1-wkcdn-tos.pstatp.com", 100},
	{WuKongCDN, "sf3-wkcdn-tos.pstatp.com", 100},
	{WuKongCDN, "sf6-wkcdn-tos.pstatp.com", 100},
	{DouYinCDN, "sf1-dycdn-tos.pstatp.com", 100},
	{DouYinCDN, "sf3-dycdn-tos.pstatp.com", 100},
	{DouYinCDN, "sf6-dycdn-tos.pstatp.com", 100},
	{FlipgramCDN, "sf1-fgcdn-tos.pstatp.com", 100},
	{FlipgramCDN, "sf3-fgcdn-tos.pstatp.com", 100},
	{FlipgramCDN, "sf6-fgcdn-tos.pstatp.com", 100},
	{TuchongCDN, "sf1-tccdn-tos.pstatp.com", 100},
	{TuchongCDN, "sf3-tccdn-tos.pstatp.com", 100},
	{TuchongCDN, "sf6-tccdn-tos.pstatp.com", 100},
	{NeiHanCDN, "sf1-nhcdn-tos.pstatp.com", 100},
	{NeiHanCDN, "sf3-nhcdn-tos.pstatp.com", 100},
	{NeiHanCDN, "sf6-nhcdn-tos.pstatp.com", 100},
	{GuaGuaLongCDN, "p6.985gm.com", 100},
	{GuaGuaLongCDN, "p3.985gm.com", 100},
	{MayaCDN, "p1.ppkankan01.com", 100},
	{MayaCDN, "p3.ppkankan01.com", 100},
}

var lfCDNs = []domain{ // TODO: mv to etcd
	{ToutiaoCDN, "lf1-ttcdn-tos.pstatp.com", 100},
	{ToutiaoCDN, "lf3-ttcdn-tos.pstatp.com", 100},
	{ToutiaoCDN, "lf6-ttcdn-tos.pstatp.com", 100},
	{XiGuaCDN, "lf1-xgcdn-tos.pstatp.com", 100},
	{XiGuaCDN, "lf3-xgcdn-tos.pstatp.com", 100},
	{XiGuaCDN, "lf6-xgcdn-tos.pstatp.com", 100},
	{HuoshanCDN, "lf1-hscdn-tos.pstatp.com", 100},
	{HuoshanCDN, "lf3-hscdn-tos.pstatp.com", 100},
	{HuoshanCDN, "lf6-hscdn-tos.pstatp.com", 100},
	{WuKongCDN, "lf1-wkcdn-tos.pstatp.com", 100},
	{WuKongCDN, "lf3-wkcdn-tos.pstatp.com", 100},
	{WuKongCDN, "lf6-wkcdn-tos.pstatp.com", 100},
	{DouYinCDN, "lf1-dycdn-tos.pstatp.com", 100},
	{DouYinCDN, "lf3-dycdn-tos.pstatp.com", 100},
	{DouYinCDN, "lf6-dycdn-tos.pstatp.com", 100},
	{FlipgramCDN, "lf1-fgcdn-tos.pstatp.com", 100},
	{FlipgramCDN, "lf3-fgcdn-tos.pstatp.com", 100},
	{FlipgramCDN, "lf6-fgcdn-tos.pstatp.com", 100},
	{TuchongCDN, "lf1-tccdn-tos.pstatp.com", 100},
	{TuchongCDN, "lf3-tccdn-tos.pstatp.com", 100},
	{TuchongCDN, "lf6-tccdn-tos.pstatp.com", 100},
	{NeiHanCDN, "lf1-nhcdn-tos.pstatp.com", 100},
	{NeiHanCDN, "lf3-nhcdn-tos.pstatp.com", 100},
	{NeiHanCDN, "lf6-nhcdn-tos.pstatp.com", 100},
}

var sfSGCDNs = []domain{ // TODO: mv to etcd
	{SGHypstarCDN, "sf-hs-sg.ibytedtos.com", 100},
	{SGTickTokCDN, "sf-tk-sg.ibytedtos.com", 100},
	{SGTopBuzzCDN, "sf-tb-sg.ibytedtos.com", 100},
}

var lfSGCDNs = []domain{ // TODO: mv to etcd
	{SGHypstarCDN, "lf-hs-sg.ibytedtos.com", 100},
	{SGTickTokCDN, "lf-tk-sg.ibytedtos.com", 100},
	{SGTopBuzzCDN, "lf-tb-sg.ibytedtos.com", 100},
}

var sfVACDNs = []domain{ // TODO: mv to etcd
	{VAMusicallyCDN, "sf16-muse-va.ibytedtos.com", 100},
}

var lfVACDNs = []domain{ // TODO: mv to etcd
	{VAMusicallyCDN, "lf16-muse-va.ibyttedtos.com", 100},
}

type byValue []consistentNode

func (a byValue) Len() int           { return len(a) }
func (a byValue) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byValue) Less(i, j int) bool { return a[i].Value < a[j].Value }

func init() {
	// china
	sfCDNDomains = make(map[ProductName][]consistentNode)
	for _, n := range sfCDNs {
		nodes := sfCDNDomains[n.Product]
		for i := 0; i < n.Weight; i++ {
			s := n.Name + "|" + strconv.Itoa(i)
			v := md5uint64(s)
			nodes = append(nodes, consistentNode{Value: v, Name: n.Name})
		}
		sfCDNDomains[n.Product] = nodes
	}
	for _, nodes := range sfCDNDomains {
		sort.Sort(byValue(nodes))
	}
	lfCDNDomains = make(map[ProductName][]consistentNode)
	for _, n := range lfCDNs {
		nodes := lfCDNDomains[n.Product]
		for i := 0; i < n.Weight; i++ {
			s := n.Name + "|" + strconv.Itoa(i)
			v := md5uint64(s)
			nodes = append(nodes, consistentNode{Value: v, Name: n.Name})
		}
		lfCDNDomains[n.Product] = nodes
	}
	for _, nodes := range lfCDNDomains {
		sort.Sort(byValue(nodes))
	}

	// sg
	sfSGCDNDomains = make(map[ProductName][]consistentNode)
	for _, n := range sfSGCDNs {
		nodes := sfSGCDNDomains[n.Product]
		for i := 0; i < n.Weight; i++ {
			s := n.Name + "|" + strconv.Itoa(i)
			v := md5uint64(s)
			nodes = append(nodes, consistentNode{Value: v, Name: n.Name})
		}
		sfSGCDNDomains[n.Product] = nodes
	}
	for _, nodes := range sfSGCDNDomains {
		sort.Sort(byValue(nodes))
	}
	lfSGCDNDomains = make(map[ProductName][]consistentNode)
	for _, n := range lfSGCDNs {
		nodes := lfSGCDNDomains[n.Product]
		for i := 0; i < n.Weight; i++ {
			s := n.Name + "|" + strconv.Itoa(i)
			v := md5uint64(s)
			nodes = append(nodes, consistentNode{Value: v, Name: n.Name})
		}
		lfSGCDNDomains[n.Product] = nodes
	}
	for _, nodes := range lfSGCDNDomains {
		sort.Sort(byValue(nodes))
	}

	// va
	sfVACDNDomains = make(map[ProductName][]consistentNode)
	for _, n := range sfVACDNs {
		nodes := sfVACDNDomains[n.Product]
		for i := 0; i < n.Weight; i++ {
			s := n.Name + "|" + strconv.Itoa(i)
			v := md5uint64(s)
			nodes = append(nodes, consistentNode{Value: v, Name: n.Name})
		}
		sfVACDNDomains[n.Product] = nodes
	}
	for _, nodes := range sfVACDNDomains {
		sort.Sort(byValue(nodes))
	}
	lfVACDNDomains = make(map[ProductName][]consistentNode)
	for _, n := range lfVACDNs {
		nodes := lfVACDNDomains[n.Product]
		for i := 0; i < n.Weight; i++ {
			s := n.Name + "|" + strconv.Itoa(i)
			v := md5uint64(s)
			nodes = append(nodes, consistentNode{Value: v, Name: n.Name})
		}
		lfVACDNDomains[n.Product] = nodes
	}
	for _, nodes := range lfVACDNDomains {
		sort.Sort(byValue(nodes))
	}
}
