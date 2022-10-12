package config

import (
	"github.com/pkg/errors"
	"github.com/peterlearn/kratos/pkg/conf/dsn"
	xtime "github.com/peterlearn/kratos/pkg/time"
)

// ServerConfig is the bm server config model
type ServerConfig struct {
	Network      string         `dsn:"network"`
	Addr         string         `dsn:"address"`
	Timeout      xtime.Duration `dsn:"query.timeout"`
	ReadTimeout  xtime.Duration `dsn:"query.readTimeout"`
	WriteTimeout xtime.Duration `dsn:"query.writeTimeout"`
}

var (
	Default405Body = []byte("405 method not allowed")
	Default404Body = []byte("404 page not found")
)

func ParseDSN(rawdsn string) *ServerConfig {
	conf := new(ServerConfig)
	d, err := dsn.Parse(rawdsn)
	if err != nil {
		panic(errors.Wrapf(err, "ServerConfig: invalid dsn: %s", rawdsn))
	}
	if _, err = d.Bind(conf); err != nil {
		panic(errors.Wrapf(err, "ServerConfig: invalid dsn: %s", rawdsn))
	}
	return conf
}
