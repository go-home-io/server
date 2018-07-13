package main

import (
	"github.com/go-home-io/server/server"
	"github.com/go-home-io/server/settings"
	"github.com/go-home-io/server/worker"
	"github.com/jessevdk/go-flags"
)

func main() {
	options := &settings.StartUpOptions{}
	_, err := flags.Parse(options)
	if err != nil {
		panic(err)
	}

	s := settings.Load(options)
	s.SystemLogger().Info("Starting go-home server")

	if options.IsWorker {
		wkr, err := worker.NewWorker(s)
		if err != nil {
			s.SystemLogger().Fatal("Failed to start go-home worker", err)
		}

		wkr.Start()

	} else {
		srv, err := server.NewServer(s)
		if err != nil {
			s.SystemLogger().Fatal("Failed to start go-home server", err)
		}

		srv.Start()
	}
}
