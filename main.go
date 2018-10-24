package main

import (
	"github.com/jessevdk/go-flags"
	"go-home.io/x/server/server"
	"go-home.io/x/server/settings"
	"go-home.io/x/server/worker"
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
