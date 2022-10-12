package gin

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/peterlearn/kratos/v1/pkg/log"
	"github.com/peterlearn/kratos/v1/pkg/net/http/config"
	"net/http"
	"os"
	"time"
)

var (
	httpconfig      *config.ServerConfig
	httpDSN         string
	httpServerPoint map[*gin.Engine]*http.Server
)

func init() {
	AddFlag(flag.CommandLine)
	httpServerPoint = make(map[*gin.Engine]*http.Server)
}

func AddFlag(fs *flag.FlagSet) {
	v := os.Getenv("HTTP")
	if v == "" {
		v = "tcp://0.0.0.0:8000/?timeout=1s"
	}
	fs.StringVar(&httpDSN, "gin.http", v, "listen http dsn, or use HTTP env variable.")
}

func DefaultServer(conf *config.ServerConfig) *gin.Engine {
	if conf == nil {
		if !flag.Parsed() {
			fmt.Fprint(os.Stderr, "[gin] please call flag.Parse() before Init gin server, some configure may not effect.\n")
		}
		conf = config.ParseDSN(httpDSN)
	}

	httpconfig = conf

	engine := NewServer()
	engine.Use(gin.Recovery(), kginlog(), ServerReqMetric())
	return engine
}

func NewServer() *gin.Engine {
	engine := gin.New()
	pprof.Register(engine)
	engine.GET("/metrics", monitor())
	engine.GET("/healthcheck", healthcheck)
	//engine.GET("/metadata", engine.metadata())
	engine.NoRoute(func(c *gin.Context) {
		c.Data(404, "text/plain", config.Default404Body)
		c.Abort()
	})
	engine.NoMethod(func(c *gin.Context) {
		c.Data(405, "text/plain", config.Default405Body)
		c.Abort()
	})
	return engine
}

func Start(engine *gin.Engine) {
	if httpconfig != nil {
		httpserver := &http.Server{
			Addr:         httpconfig.Addr,
			Handler:      engine,
			ReadTimeout:  time.Duration(httpconfig.ReadTimeout),
			WriteTimeout: time.Duration(httpconfig.WriteTimeout),
		}
		log.Info("Gin: start HTTP listen addr: %v", httpconfig.Addr)
		httpServerPoint[engine] = httpserver
		go httpserver.ListenAndServe()
	} else {
		go engine.Run()
	}
}

func Shutdown(engine *gin.Engine, ctx context.Context) error {
	if srv, ok := httpServerPoint[engine]; ok {
		err := srv.Shutdown(ctx)
		return err
	} else {
		return errors.New("Invalid gin.Engine")
	}
}
