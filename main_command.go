package main

import (
	. "docker/mydocker/container"
	. "docker/mydocker/network"
	. "docker/mydocker/util"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var containersCommand = cli.Command{
	Name:  "containers",
	Usage: "container commands",
	Subcommands: []cli.Command{

		cli.Command{
			Name:  "init",
			Usage: "Init container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name",
					Usage: "container name",
				},
			},
			Action: func(context *cli.Context) error {

				var err error

				log.Debugf("start to invoke init.")

				// container name
				containerName := context.String("name")

				// init logger
				if containerName != "" {
					RuntimeLogger = GetPersistentLogger(containerName)
					RuntimeLogger.Println("init setup persistent logger")
				}

				if err = RunContainerInitProcess(containerName); err != nil {
					RuntimeLogger.Printf("run RunContainerInitProcess failed, %v", err)
					UpdateContainerStatusByName(containerName, ContainerFailed)
				}

				return err
			},
		},

		cli.Command{
			Name:  "run",
			Usage: `Create a container`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "it",
					Usage: "enable tty",
				},
				cli.BoolFlag{
					Name:  "d",
					Usage: "detach mode",
				},
				cli.StringFlag{
					Name:  "m",
					Usage: "memory limit",
				},
				cli.StringFlag{
					Name:  "cpushare",
					Usage: "cpushare limit",
				},
				cli.StringFlag{
					Name:  "cpuset",
					Usage: "cpuset limit",
				},
				cli.StringFlag{
					Name:  "v",
					Usage: "volume mapping",
				},
				cli.StringFlag{
					Name:  "i",
					Usage: "image name",
					Value: "busybox",
				},
				cli.StringFlag{
					Name:  "name",
					Usage: "container name",
				},
				// -e bird1=1 -e bird2=2
				cli.StringSliceFlag{
					Name:  "e",
					Usage: "container environment",
				},
				cli.StringFlag{
					Name:  "net",
					Usage: "container network",
				},
				cli.StringSliceFlag{
					Name:  "p",
					Usage: "port mapping",
				},
			},
			Action: func(context *cli.Context) error {

				log.Debugf("start to invoke run.")

				imageName := context.String("i")
				if imageName == "" {
					return NewError("image name not set")
				}

				if !CheckImage(imageName) {
					return NewError("image %s not support", imageName)
				} else {
					log.Debugf("start to use image %s", imageName)
				}

				// split commands into array
				var cmdArray []string
				for _, arg := range context.Args() {
					cmdArray = append(cmdArray, arg)
				}
				log.Debugf("call run with args: %v ", context.Args())

				// container name
				containerName := context.String("name")
				// check container
				if CheckContainer(containerName) {
					return NewError("container %s already exist", containerName)
				}

				// env
				containerEnv := context.StringSlice("e")

				// network
				network := context.String("net")
				if network == "" {
					network = DefNetworkName
				}

				// port map
				portMapping := context.StringSlice("p")

				// tty
				ttyMode := context.Bool("it")

				// detach
				detachMode := context.Bool("d")

				// volumn
				volumeMap := context.String("v")

				// avoid enable both tty and detach mode
				if detachMode && ttyMode {
					return NewError("shouldn't set -d and -it together.")
				}

				// cgroup
				resConf := &ResourceConfig{
					MemoryLimit: context.String("m"),
					CpuSet:      context.String("cpuset"),
					CpuShare:    context.String("cpushare"),
				}

				// init container
				GlobalContainer = NewContainerInfo(
					0,
					imageName,
					&containerName,
					network,
					volumeMap,
					ttyMode,
					strings.Join(cmdArray, " "),
					containerEnv,
					portMapping,
					resConf,
				)

				// init logger
				if containerName != "" {
					RuntimeLogger = GetPersistentLogger(containerName)
					RuntimeLogger.Debugln("run setup persistent logger")

					if err := PersistContainerInfo(&GlobalContainer); err != nil {
						log.Errorf("runCommand failed, %v", err)
					}
				}

				Run(&GlobalContainer)

				return nil

			},
		},

		cli.Command{
			Name:  "commit",
			Usage: "Create a new image from a container's changes",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "i",
					Usage: "image name",
				},
			},
			Action: func(context *cli.Context) error {

				log.Debugf("start to invoke commit.")
				// get container name
				if len(context.Args()) < 1 {
					return NewError("container name not set")
				}

				// check image name
				imageName := context.String("i")
				if imageName == "" {
					return NewError("image name not set")
				}

				if CheckImage(imageName) {
					return NewError("image name already exist.")
				}

				// check container name
				ContainerName := context.Args().Get(0)
				if !CheckContainer(ContainerName) {
					return NewError("container %s not exist. please check again", ContainerName)
				}

				CommitContainer(ContainerName, imageName)

				return nil
			},
		},

		cli.Command{
			Name:  "ps",
			Usage: "List containers",
			Action: func(context *cli.Context) error {
				log.Debugf("start to invoke ps.")
				ListContainers(true)
				return nil
			},
		},

		cli.Command{
			Name:  "logs",
			Usage: "Fetch the logs of a container",
			Action: func(context *cli.Context) error {
				log.Debugf("start to invoke log.")
				// get container name
				if len(context.Args()) < 1 {
					return NewError("container name didn't set")
				}
				// check container name
				containerName := context.Args().Get(0)
				if !CheckContainer(containerName) {
					return NewError("container %s not exist. please check again", containerName)
				}
				GrabContainerLog(containerName)
				return nil
			},
		},

		cli.Command{
			Name:  "rm",
			Usage: "delete a container",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "a",
					Usage: "all container",
				},
			},
			Action: func(context *cli.Context) error {

				log.Debugf("start to invoke rm.")

				var containerName string

				// is that mean delete all instances?
				delAll := context.Bool("a")

				// get container name
				if len(context.Args()) < 1 && !delAll {
					return NewError("container name didn't set")
				}

				// if only delete on container then check container name
				if len(context.Args()) == 1 {
					// get container name from first variable
					containerName = context.Args().Get(0)
					// check!!
					if !CheckContainer(containerName) {
						return NewError("container %s not exist. please check again", containerName)
					}
					if GetContinerInfoByName(containerName).Status == ContainerRunning {
						log.Errorf("Shouldn't delete running container %s", containerName)
						return nil
					}

				}

				DeleteContainer(containerName, delAll)

				return nil
			},
		},

		cli.Command{
			Name:  "stop",
			Usage: "stop a container",
			Action: func(context *cli.Context) error {
				log.Debugf("start to invoke stop.")
				// get container name
				if len(context.Args()) < 1 {
					return fmt.Errorf("container name didn't set")
				}
				containerName := context.Args().Get(0)
				if GetContinerInfoByName(containerName).Status == ContainerStopped {
					return NewError("container %s already stopped", containerName)
				}
				StopContainer(containerName)
				return nil
			},
		},

		cli.Command{
			Name:  "exec",
			Usage: "Run a command in a running container",
			Action: func(context *cli.Context) error {
				log.Debugf("start to invoke exec.")

				// -it/-d command
				if len(context.Args()) < 2 {
					return NewError("miss container name & command")
				}

				// check container name
				containerName := context.Args().Get(0)
				if !CheckContainer(containerName) {
					return NewError("container %s not exist. please check again", containerName)
				}
				command := context.Args().Get(1)

				ExecContainer(containerName, command)
				return nil
			},
		},
	},
}

var imagesCommand = cli.Command{
	Name:  "images",
	Usage: "image commands",
	Subcommands: []cli.Command{
		{
			Name:  "ls",
			Usage: "list supported images",
			Action: func(context *cli.Context) error {

				log.Debugf("start to list images.")

				ListImages(true)
				return nil
			},
		},

		{
			Name:  "rm",
			Usage: "remove supported images",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "i",
					Usage: "image name",
				},
			},
			Action: func(context *cli.Context) error {

				log.Debugf("start to remove image.")

				imageName := context.String("i")
				if imageName == "" {
					return NewError("image name not set")
				}

				RemoveImages(imageName)
				return nil
			},
		},
	},
}

var networkCommand = cli.Command{
	Name:  "network",
	Usage: "network commands",
	Subcommands: []cli.Command{
		{
			Name:  "create",
			Usage: "create a container network",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "driver",
					Usage: "network driver",
				},
				cli.StringFlag{
					Name:  "subnet",
					Usage: "subnet cidr",
				},
			},
			Action: func(context *cli.Context) error {

				// get network name
				if len(context.Args()) < 1 {
					return NewError("network name didn't set")
				}
				networkName := context.Args()[0]

				// get subnet
				subnet := context.String("subnet")
				if subnet == "" {
					return NewError("subnet didn't set")
				} else {
					log.Debugf("subetnet set to %s", subnet)
				}

				// get network driver
				driverType := context.String("driver")
				if driverType == "" {
					return NewError("network driver didn't set")
				}
				if !ValidNetworkDriver(driverType) {
					log.Errorf("driver %s not support", driverType)
				}

				// create network
				err := CreateNetwork(driverType, subnet, networkName)
				if err != nil {
					return NewError("create network %s, %+v", networkName, err)
				}

				log.Debugf("create network %s successfully.", networkName)
				return nil
			},
		},
		{
			Name:  "ls",
			Usage: "list container network",
			Action: func(context *cli.Context) error {
				ListNetwork()
				return nil
			},
		},
		{
			Name:  "rm",
			Usage: "remove container network",
			Action: func(context *cli.Context) error {

				// get network name
				if len(context.Args()) < 1 {
					return NewError("network name didn't set")
				}
				networkName := context.Args()[0]

				err := DeleteNetwork(networkName)
				if err != nil {
					return NewError("remove network %s, %+v", networkName, err)
				}

				return nil
			},
		},
	},
}
