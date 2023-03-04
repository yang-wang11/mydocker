package cgroups

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type CpusetSubSystem struct {
	CPUSet           string
	SubsysCgroupPath string
}

const cpusetFile string = "cpuset.cpus"

func (s *CpusetSubSystem) Set(cgroupPath string, res interface{}) error {

	// get cgroup path
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		s.SubsysCgroupPath = subsysCgroupPath
	} else {
		return err
	}

	// set parameter
	if cpuSet, ok := res.(string); ok && cpuSet != "" {
		if err := ioutil.WriteFile(path.Join(s.SubsysCgroupPath, cpusetFile), []byte(cpuSet), 0644); err != nil {
			return fmt.Errorf("Set: cgroup cpuset %s setup failed. %v", s.CPUSet, err)
		} else {
			s.CPUSet = cpuSet
			log.Debugf("Set: cgroup cpuset %s setup successfully. ", s.CPUSet)
		}
	} else {
		log.Debugf("Set: cgroup cpuset ignored!! ")
	}

	return nil

}

func (s *CpusetSubSystem) Remove(cgroupPath string) error {
	return os.RemoveAll(s.SubsysCgroupPath)
}

func (s *CpusetSubSystem) Apply(cgroupPath string, pid int) error {
	if s.CPUSet != "" {
		if err := ioutil.WriteFile(path.Join(s.SubsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("Apply: cgroup cpuset %s setup failed. %v", s.CPUSet, err)
		} else {
			log.Debugf("Apply: cgroup cpuset %s setup successfully. ", s.CPUSet)
			return nil
		}
	} else {
		log.Debugf("Apply: cgroup cpuset ignored!! ")
		return nil
	}
}

func (s *CpusetSubSystem) Name() subSysType {
  return SUBSYSCPUSET
}
