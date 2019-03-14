## 平台

[动态配置中心](http://cloud.bytedance.net/tcc/all)


## WIKI

[动态配置中心介绍](https://wiki.bytedance.net/pages/viewpage.action?pageId=219247405)

## USAGE

### 基础用法

```
import "code.byted.org/gopkg/tccclient"

var (
	client *tccclient.Client
)

func init() {
	serviceName := "toutiao.test.test"
	config := tccclient.NewConfig()
	var err error
	client, err = tccclient.NewClient(serviceName, config)
	if err != nil {
		panic(err)
	}
}

func main() {
    // Get gets value by config key, may return error if the cnofig doesn't exist
	value, err := client.Get(key)
    if err != nil {
        println(value)
    }
}
```

### 带缓存的自定义解析。
当value有更新才会触发重新解析，避免无效的解析成本。

1. 实现一个 TCCParser
```
type TCCParser func(value string, err error, cacheResult interface{}) (interface{}, error)
```
2. 使用 GetWithParser 获取解析后的结果
```
result, err := client.GetWithParser(key, YourTCCParser)
```

- 每次 value 发生变动，或获取失败的时候会触发 TCCParser 函数。
- TCCParser 返回 err == nil 则会将结果 cache。
- TCCParser 的返回结果会被透传回 GetWithParser。
