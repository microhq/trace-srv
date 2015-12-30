package main

import (
	"github.com/codegangsta/cli"
	log "github.com/golang/glog"
	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/server"

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
	cmd.Init()

	// Add for MySQL configuration
	cmd.Flags = append(cmd.Flags,
		cli.StringFlag{
			Name:   "database_url",
			EnvVar: "DATABASE_URL",
			Usage:  "The database URL e.g root@tcp(127.0.0.1:3306)/trace",
		},
	)

	cmd.Actions = append(cmd.Actions, func(c *cli.Context) {
		mysql.Url = c.String("database_url")
	})

	server.Init(
		server.Name("go.micro.srv.trace"),
	)

	proto.RegisterTraceHandler(server.DefaultServer, new(handler.Trace))

	server.Subscribe(
		server.NewSubscriber(trace.TraceTopic, trace.ProcessSpan),
	)

	if err := db.Init(); err != nil {
		log.Fatal(err)
	}

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
