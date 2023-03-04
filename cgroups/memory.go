package cgroups

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type MemorySubSystem struct {
	Memory           string
	SubsysCgroupPath string
}

const memFile string = "memory.limit_in_bytes"
const oomKill string = "memory.oom_control"

func (s *MemorySubSystem) Set(cgroupPath string, res interface{}) error {

	// get cgroup path
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err != nil {
		return err
	} else {
		s.SubsysCgroupPath = subsysCgroupPath
	}

	// set memory limit
	if memoryLimit, ok := res.(string); ok && memoryLimit != "" {
		fullmemFile := path.Join(s.SubsysCgroupPath, memFile)
		if err := ioutil.WriteFile(fullmemFile, []byte(memoryLimit), 0644); err != nil {
			return fmt.Errorf("Set: cgroup memory %s setup failed. %v", s.Memory, err)
		} else {
			s.Memory = memoryLimit
			log.Debugf("Set: cgroup memory %s setup successfully. ", s.Memory)
		}

		// avoid oom killer
		fullOomKillerFile := path.Join(s.SubsysCgroupPath, oomKill)
		if err := ioutil.WriteFile(fullOomKillerFile, []byte("1"), 0644); err != nil {
			return fmt.Errorf("Set: cgroup memory disable oom killer failed. %v", err)
		} else {
			s.Memory = memoryLimit
			log.Debugf("Set: cgroup memory disable oom killer successfully. ")
		}

	} else {
		log.Debugf("Set: cgroup memory ignored!! ")
	}

	return nil

}

func (s *MemorySubSystem) Remove(cgroupPath string) error {
	return os.RemoveAll(s.SubsysCgroupPath)
}

func (s *MemorySubSystem) Apply(cgroupPath string, pid int) error {
	if s.Memory != "" {
		if err := ioutil.WriteFile(path.Join(s.SubsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("Apply: cgroup memory %s setup failed. %v", s.Memory, err)
		} else {
			log.Debugf("Apply: cgroup memory %s setup successfully. ", s.Memory)
			return nil
		}
	} else {
		log.Debugf("Apply: cgroup memory ignored!! ")
		return nil
	}
}

func (s *MemorySubSystem) Name() subSysType {
	return SUBSYSMEMORY
}
