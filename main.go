package main

import (
	log "github.com/golang/glog"
	"github.com/micro/cli"
	micro "github.com/micro/go-micro"

	// trace
	"github.com/micro/trace-srv/handler"
	"github.com/micro/trace-srv/trace"

	// db
	"github.com/micro/trace-srv/db"
	"github.com/micro/trace-srv/db/mysql"

	// proto
	proto "github.com/micro/trace-srv/proto/trace"
)

func main() {
	service := micro.NewService(
		micro.Name("go.micro.srv.trace"),

		micro.Flags(
			cli.StringFlag{
				Name:   "database_url",
				EnvVar: "DATABASE_URL",
				Usage:  "The database URL e.g root@tcp(127.0.0.1:3306)/trace",
			},
		),
		// Add for MySQL configuration
		micro.Action(func(c *cli.Context) {
			if len(c.String("database_url")) > 0 {
				mysql.Url = c.String("database_url")
			}
		}),
	)

	service.Init()

	proto.RegisterTraceHandler(service.Server(), new(handler.Trace))

	service.Server().Subscribe(
		service.Server().NewSubscriber(trace.TraceTopic, trace.ProcessSpan),
	)

	if err := db.Init(); err != nil {
		log.Fatal(err)
	}

	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
