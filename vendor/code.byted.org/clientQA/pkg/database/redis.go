package database

import (
	"code.byted.org/gopkg/logs"
	"code.byted.org/kv/goredis"
	"strconv"
	"time"
)

var redisClient *goredis.Client

func InitRedis() {
	var err error
	redisClient, err = goredis.NewClient("toutiao.redis.pkg")

	if err != nil {
		logs.Error("connect redis error : %v", err)
		return
	}
}

/*
设置某个三位版本下的五位版本号，到redis中
*/
func SetFiveVersion(aid string, platform string, outerVersion string, fiveVersion string) (bool, error) {
	//首先先给版本号区分一下？用三位版本号好了
	if outerVersion == "" {
		if len(fiveVersion) < 4 {
			logs.Error("%v", "五位(四位)版本号不正确")
			return false, nil
		} else {
			//这里需要提取五位版本号前三位
			outerVersion = convertVersionCodes(fiveVersion)
		}
	}
	res, err := redisClient.HSet(aid+platform, outerVersion, fiveVersion).Result()
	if err != nil {
		logs.Error("%v", err)
		//这事情比较特殊，需要给我发消息了
		//http_util.LarkDingOneInnerV2("maoyu", "hset5位版本号失败，错误是" + err.Error())
	}
	return res, nil
}

/*
删除Redis里面的冗余数据
*/
func DelOldFiveVersion(aid string, platform string, maxVersion int) {
	//首先先给版本号区分一下？用三位版本号好了
	outerVersion := convertVersionCodes(strconv.Itoa(maxVersion - 500))
	//res, err := redisClient.HDel(aid+platform, outerVersion).Result()
	redisClient.HDel(aid+platform, outerVersion).Result()
	//if err != nil {
	//	logs.Error("%v", err)
	//	//这事情比较特殊，需要给我发消息了
	//	//http_util.LarkDingOneInnerV2("maoyu", "hset5位版本号失败，错误是" + err.Error())
	//}
	//return res, nil
}

/*
设置某个产品线下的所有action的最高版本号
*/
func SetOuterVersion(aid string, platform string, content string, outerVersion string, fiveVersion string) (string, error) {
	//首先先给版本号区分一下？用三位版本号好了
	if outerVersion == "" {
		if len(fiveVersion) < 4 {
			logs.Error("%v", "五位(四位)版本号不正确")
			return "", nil
		} else {
			//这里需要提取五位版本号前三位
			outerVersion = convertVersionCodes(fiveVersion)
		}
	}
	str, err := redisClient.Set(aid+platform+content, outerVersion, 0).Result()
	if err != nil {
		logs.Error("%v", err)
	}
	return str, nil
}

/*
	5位版本号不含.（长度为5），分割成x.x.x
*/
func convertVersionCodes(version string) string {
	outerVersion := ""

	for ind, elem := range version {
		if ind >= 3 {
			break
		} else if ind == 2 {
			outerVersion += string(rune(elem))
		} else {
			outerVersion += string(rune(elem)) + "."
		}
	}

	return outerVersion
}

func GetFiveVersions(aid string, platform string) map[string]string {
	return redisClient.HGetAll(aid + platform).Val()
}

func GetOuterVersions(aid string, platform string) string {
	return redisClient.Get(aid + platform + "official").Val()
}

/*
在Redis中存储用户的lark id
因为该id不会变化，因此map存储即可
*/
func SetLarkID(user string, larkID string) (bool, error) {
	res, err := redisClient.HSet("LarkID", user, larkID).Result()
	if err != nil {
		logs.Error("%v", err)
	}
	return res, nil
}

/*
在Redis中获取用户的lark id
*/
func GetLarkID(name string) string {
	return redisClient.HGet("LarkID", name).Val()
}

/*
在Redis中存储和用户聊天的chart id
这个chart可能会过期，因此设置一天的过期时间即可
*/
func SetLarkChannel(name string, channel string) (string, error) {
	str, err := redisClient.Set(name, channel, time.Hour*24).Result()
	if err != nil {
		logs.Error("%v", err)
	}
	return str, nil
}

/*
在Redis中获取和用户聊天的chart id
*/
func GetLarkChannel(name string) string {
	return redisClient.Get(name).Val()
}

//func SetBranchInfoToRedis(branch string, time time.Time, info interface{}) error {
//	//var rediskey string
//	rediskey := PKG_BRANCH_INFO
//	_ , err := redisClient.Do("HSet", rediskey+branch, time.String(), info)
//	if err != nil {
//		logs.Error("%v", err)
//		return fmt.Errorf("redis set error : branch:%s, time:%s, err:%s", branch, time, err)
//	}
//	return err
//}
//
//func GetAllBranchInfo(branch string) (map[string]string, error)  {
//	ret, err := redisClient.Do("HGetAll", branch)
//	if err != nil {
//		logs.Error("%v", err)
//		return nil, fmt.Errorf("redis hgetall error : branch:%s, err:%s", branch, err)
//	}
//	res, err := redis.StringMap(ret, nil)
//	if err != nil {
//		logs.Error("%v", err)
//		return nil, err
//	}
//	return res, nil
//}

//func GetUserPluginConfig(appId int, deviceId int64) (map[string]string, error) {
//	ret, err := redisClient.Do("hgetall", key(appId, deviceId))
//	if err != nil {
//		logs.Error("%v", err)
//		return nil, fmt.Errorf("redis hgetall error : aid:%s, did:%s, err:%s", appId, deviceId, err)
//	}
//	res, err := redis.StringMap(ret, nil)
//	if err != nil {
//		logs.Error("%v", err)
//		return nil, err
//	}
//	return res, nil
//}
//
//func SetSessionToRedis(k string, v string) error {
//	_, err := redisClient.Do("SET", fmt.Sprintf("c:%s:%s", k, v))
//	if err != nil {
//		logs.Error("%v", err)
//		return fmt.Errorf("redis hgetall error : aid:%s, did:%s, err:%s", k, v, err)
//	}
//	return nil
//}
