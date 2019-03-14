configstorer
======
Use asyncache to cache configs stored by ETCD.

# Usage
```go
import "code.byted.org/microservice/configstorer"

configstorer.InitStorerWithBlocking(10 * time.Millesecond)
//or
configstorer.InitStorer()
  
configstore.SetGetterTimeout(50 * time.Millesecond) //default 20ms
  
//Get config key, err == asyncache.EmptyErr if the key is not present
val, err := configstorer.Get(key)  
  
//Get config key, return default value if nonexist
val, err := configstorer.GetOrDefault(key, "default")  
  
if configstorer.IsKeyNonexist(err) {
	//Key not exist
}

```
