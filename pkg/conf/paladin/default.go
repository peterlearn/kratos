package paladin

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/peterlearn/kratos/v1/pkg/log"
)

var (
	// DefaultClient default client.
	DefaultClient Client
	confPath      string
)

func init() {
	flag.StringVar(&confPath, "conf", "", "default config path")
}

// Init init config client.
// If confPath is set, it inits file client by default
// Otherwise we could pass args to init remote client
// args[0]: driver name, string type
func Init(args ...interface{}) (err error) {
	if confPath != "" {
		DefaultClient, err = NewFile(confPath)
	} else {
		var (
			driver Driver
		)
		argsLackErr := errors.New("lack of remote config center args")
		if len(args) == 0 {
			panic(argsLackErr.Error())
		}
		argsInvalidErr := errors.New("invalid remote config center args")
		driverName, ok := args[0].(string)
		if !ok {
			panic(argsInvalidErr.Error())
		}
		driver, err = GetDriver(driverName)
		if err != nil {
			return
		}
		DefaultClient, err = driver.New()
	}
	if err != nil {
		return
	}
	return
}

func IsExistStringArray(m string, a []string) bool {
	for _, v := range a {
		if v == m {
			return true
		}
	}
	return false
}

// Watch watch on a key. The configuration implements the setter interface, which is invoked when the configuration changes.
func Watch(keys []string, mm map[string]Setter) {
	var err error
	ll := len(keys)
	// 优先初始化的文件
	for idx := 0; idx < ll; idx++ {
		k := keys[idx]
		err = SetConf(k, mm[k])
		if err != nil {
			panic(fmt.Sprintf("file:%s, err:%v", k, err))
		}
	}

	// 其他文件初始化
	for k, _ := range mm {
		if IsExistStringArray(k, keys) {
			continue
		}
		err = SetConf(k, mm[k])
		if err != nil {
			panic(fmt.Sprintf("file:%s, err:%v", k, err))
		}
	}

	go func(mm map[string]Setter) {
		for event := range WatchEvent(context.TODO(), []string{}...) {
			m, ok := mm[event.Key]
			if !ok {
				continue
			}
			log.Debug("Watch file event :%s, size:%d", event.Key, len(event.Value))
			err = m.Set(event.Value)
			if err != nil {
				log.Error("Watch Set key:%s err:%v", event.Key, err)
			}
		}
	}(mm)
	return
}

func SetConf(key string, setter Setter) (err error) {
	v := DefaultClient.Get(key)
	if v == nil {
		return ErrNotExist
	}
	err = setter.Set(v.raw)
	if err != nil {
		return
	}
	log.Debug("SetConf :%s, size:%d", key, len(v.raw))
	return
}

// WatchEvent watch on multi keys. Events are returned when the configuration changes.
func WatchEvent(ctx context.Context, keys ...string) <-chan Event {
	return DefaultClient.WatchEvent(ctx, keys...)
}

// Get return value by key.
func Get(key string) *Value {
	return DefaultClient.Get(key)
}

// GetAll return all config map.
func GetAll() *Map {
	return DefaultClient.GetAll()
}

// Keys return values key.
func Keys() []string {
	return DefaultClient.GetAll().Keys()
}

// Close close watcher.
func Close() error {
	return DefaultClient.Close()
}
