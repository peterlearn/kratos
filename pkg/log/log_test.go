package log

import (
	"context"
	"fmt"
	"testing"

	"github.com/pll/kratos/pkg/net/metadata"

	"github.com/stretchr/testify/assert"
)

func initStdout() {
	conf := &Config{
		Stdout: true,
	}
	Init(conf)
}

func initFile() {
	conf := &Config{
		Dir: "/tmp",
		// VLevel:  2,
		Module: map[string]int32{"log_test": 1},
	}
	Init(conf)
}

type TestLog struct {
	A string
	B int
	C string
	D string
}

func testLog(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		Error("hello %s", "world")
		Errorv(context.Background(), KV("key", 2222222), KV("test2", "test"))
		Errorc(context.Background(), "keys: %s %s...", "key1", "key2")
	})
	t.Run("Warn", func(t *testing.T) {
		Warn("hello %s", "world")
		Warnv(context.Background(), KV("key", 2222222), KV("test2", "test"))
		Warnc(context.Background(), "keys: %s %s...", "key1", "key2")
	})
	t.Run("Info", func(t *testing.T) {
		Info("hello %s", "world")
		Infov(context.Background(), KV("key", 2222222), KV("test2", "test"))
		Infoc(context.Background(), "keys: %s %s...", "key1", "key2")
	})
}

func TestFile(t *testing.T) {
	initFile()
	testLog(t)
	assert.Equal(t, nil, Close())
}

func TestStdout(t *testing.T) {
	initStdout()
	testLog(t)
	assert.Equal(t, nil, Close())
}

func TestLogW(t *testing.T) {
	D := logw([]interface{}{"i", "like", "a", "dog"})
	if len(D) != 2 || D[0].Key != "i" || D[0].Value != "like" || D[1].Key != "a" || D[1].Value != "dog" {
		t.Fatalf("logw out put should be ' {i like} {a dog}'")
	}
	D = logw([]interface{}{"i", "like", "dog"})
	if len(D) != 1 || D[0].Key != "i" || D[0].Value != "like" {
		t.Fatalf("logw out put should be ' {i like}'")
	}
}

func TestLogWithMirror(t *testing.T) {
	Info("test log")
	mdcontext := metadata.NewContext(context.Background(), metadata.MD{metadata.Mirror: "true"})
	Infov(mdcontext, KV("key1", "val1"), KV("key2", ""), KV("log", "log content"), KV("msg", "msg content"))

	mdcontext = metadata.NewContext(context.Background(), metadata.MD{metadata.Mirror: "***"})
	Infov(mdcontext, KV("key1", "val1"), KV("key2", ""), KV("log", "log content"), KV("msg", "msg content"))

	Infov(context.Background(), KV("key1", "val1"), KV("key2", ""), KV("log", "log content"), KV("msg", "msg content"))

}

func TestOverwriteSouce(t *testing.T) {
	ctx := context.Background()
	t.Run("test source kv string", func(t *testing.T) {
		Infov(ctx, KVString("source", "test"))
	})
	t.Run("test source kv string", func(t *testing.T) {
		Infov(ctx, KV("source", "test"))
	})
}

func BenchmarkLog(b *testing.B) {
	ctx := context.Background()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Infov(ctx, KVString("test", "hello"), KV("int", 34), KV("hhh", "hhhh"))
		}
	})
}

func TestFile2(t *testing.T) {
	initFile()
	Infov(context.Background(), KV(KeyTS, 121), KV(KeyRet, 857), KV(KeyReqLen, 12), KV(KeyRspLen, 789))
	assert.Equal(t, nil, Close())
}

func TestMap(t *testing.T) {
	mm := make(map[string]int32, 15)
	addMap(mm)
	addMap2(mm)
	fmt.Println(mm)
}

func addMap(mm map[string]int32) {
	mm["1"] = 2
	mm["2"] = 3
}
func addMap2(mm map[string]int32) {
	mm["3"] = 4
	mm["4"] = 5
}
