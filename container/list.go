package container

import (
	. "docker/mydocker/util"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"text/tabwriter"
	"time"
)

// generate n-length integer as ID
func GeneratorID(n int) string {

	bytesList := "0123456789"

	Ids := make([]byte, n)

	// setup random seed
	rand.Seed(time.Now().Unix())

	for i := 0; i < n; i++ {
		Ids[i] = bytesList[rand.Intn(len(bytesList))]
	}

	return string(Ids)

}

func PersistContainerInfo(c *Container) error {

	//log.Debugf("Container info %v ", c)

	ContainerInfo, err := json.Marshal(c);
	if err != nil {
		return NewError("Marshal container %s info failed.", c.ContainerName)
	}

	// persist container info to storage
	ContainerRuntimePath := path.Join(ContainerRuntimeBaseFolder, c.ContainerName)
	ContainerRuntimeConf := path.Join(ContainerRuntimePath, ContainerConfigName)
	if !PathExists(ContainerRuntimePath) {
		if err := os.MkdirAll(ContainerRuntimePath, 0644); err != nil {
			return NewError("create folder %s failed.", ContainerRuntimePath)
		}
	}

	if err := ioutil.WriteFile(ContainerRuntimeConf, []byte(ContainerInfo), 0644); err != nil {
		return NewError("Persist container %s info failed.", c.ContainerName)
	}

	log.Debugf("Persist container %s successfully.", c.ContainerName)
	return nil

}

func ListContainers(output bool) (cons []string) {

	var containers []Container

	// list the files from containers folder
	files, err := ioutil.ReadDir(ContainerRuntimeBaseFolder);
	if err != nil {
		log.Errorf("list container folder failed.")
		return nil
	}

	for _, file := range files {
		containers  = append(containers, GetContinerInfoByName(file.Name()))
		cons = append(cons, file.Name())
	}

	if output {

		// print containers
		w := tabwriter.NewWriter(os.Stdout, 9, 1, 2, ' ', 0)
		fmt.Fprint(w, "ID\tNAME\tIMAGE\tPID\tIP\tPORTMAPPING\tSTATUS\tCOMMAND\tCREATED\n")
		for _, container := range containers {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					container.Id, container.ContainerName, container.ImageName, container.Pid, container.IP,
					container.PortMapping, container.Status, container.Command, container.CreateTime)
		}
		if err := w.Flush(); err != nil {
			log.Errorf("Flush failed, %v", err)
			return
		}

	}

  return cons

}
