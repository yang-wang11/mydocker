package main

import (
	"github.com/yang-wang11/mydocker/network"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = "A simple container runtime implementation."
	app.Version = "1.0.0"

	app.Commands = []cli.Command{
		containersCommand,
		imagesCommand,
		networkCommand,
	}

	app.Before = func(context *cli.Context) error {

		// log as JSON instead of the default ASCII formatter.
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(log.DebugLevel)

		network.Initbridge()

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
