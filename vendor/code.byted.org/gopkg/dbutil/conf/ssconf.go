package conf

const (
	Write     = "write"
	Read      = "read"
	Slave     = "slave"
	Offline   = "offline"
	BlackHole = "blackhole"
)

type DbConfKey struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

func GetDefaultConfKey(db string, cluster string) DbConfKey {
	key := db + "_" + cluster
	return DbConfKey{
		User:     "ss_" + key + "_user",
		Password: "ss_" + key + "_password",
		Host:     "ss_" + key + "_host",
		Port:     "ss_" + key + "_port",
		Name:     "ss_" + db + "_name",
	}
}

func GetConfKey(db string, cluster string) (ret DbConfKey) {
	key := db + "_" + cluster
	exceptData := map[string]DbConfKey{
		"": DbConfKey{},
	}
	ret, found := exceptData[key]
	if found == false {
		ret = GetDefaultConfKey(db, cluster)
	}
	return ret
}

func GetDbConf(conf map[string]string, db string, cluster string) DBOptional {
	key := GetConfKey(db, cluster)
	ret := GetDefaultDBOptional()
	ret.DBName = conf[key.Name]
	ret.User = conf[key.User]
	ret.Password = conf[key.Password]
	ret.DBHostname = conf[key.Host]
	ret.DBPort = conf[key.Port]
	return ret
}
