package main

import (
	"flag"

	"github.com/coxlong/eureka/internal/app"
	"github.com/coxlong/eureka/internal/pkg/config"
)

func main() {
	var conf string
	flag.StringVar(&conf, "conf", "./configs/dev.toml", "应用配置文件，默认为\"./configs/dev.toml\"")
	flag.Parse()

	cfg, err := config.LoadConf(conf)
	if err != nil {
		panic((err))
	}

	engine, err := app.Bootstrap(cfg)
	if err != nil {
		panic(err)
	}
	if cfg.Server.HTTPS {
		if err := engine.RunTLS(cfg.Server.Addr, cfg.Server.CrtFile, cfg.Server.KeyFile); err != nil {
			panic(err)
		}
	} else {
		if err := engine.Run(cfg.Server.Addr); err != nil {
			panic(err)
		}
	}
}
