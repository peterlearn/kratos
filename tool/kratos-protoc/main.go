package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "protoc"
	app.Usage = "protobuf生成工具"
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "bm",
			Usage:       "whether to use BM for generation",
			Destination: &withBM,
		},
		&cli.BoolFlag{
			Name:        "gin",
			Usage:       "whether to use BM for generation",
			Destination: &withGin,
		},
		&cli.BoolFlag{
			Name:        "grpc",
			Usage:       "whether to use gRPC for generation",
			Destination: &withGRPC,
		},
		&cli.BoolFlag{
			Name:        "ws",
			Usage:       "whether to use ws for generation",
			Destination: &withWS,
		},
		&cli.BoolFlag{
			Name:        "tcp",
			Usage:       "whether to use tcp for generation",
			Destination: &withTcp,
		},
		&cli.BoolFlag{
			Name:        "tcp_loop",
			Usage:       "whether to use tcp_loop for generation",
			Destination: &withTcpLoop,
		},
		&cli.BoolFlag{
			Name:        "comet",
			Usage:       "whether to use comet for generation",
			Destination: &withComet,
		},
		&cli.BoolFlag{
			Name:        "gomsg",
			Usage:       "whether to use gomsg for generation",
			Destination: &withGomsg,
		},
		&cli.BoolFlag{
			Name:        "gomsg_loop",
			Usage:       "whether to use gomsg loop for generation",
			Destination: &withGomsgLoop,
		},
		&cli.BoolFlag{
			Name:        "swagger",
			Usage:       "whether to use swagger for generation",
			Destination: &withSwagger,
		},
		&cli.BoolFlag{
			Name:        "ecode",
			Usage:       "whether to use ecode for generation",
			Destination: &withEcode,
		},
	}
	app.Action = func(c *cli.Context) error {
		return protocAction(c)
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
