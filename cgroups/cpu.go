package cgroups

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	log "github.com/Sirupsen/logrus"
)

type CpuSubSystem struct{
	CPUShare string
	SubsysCgroupPath string
}

const cpuFile string = "cpu.shares"

func (s *CpuSubSystem) Set(cgroupPath string, res interface{}) error {

	// get cgroup path
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err != nil {
		return err
	} else {
		s.SubsysCgroupPath = subsysCgroupPath
	}

	// set parameter
	if cpuShare, ok := res.(string); ok && cpuShare != "" {
		if err := ioutil.WriteFile(path.Join(s.SubsysCgroupPath, cpuFile), []byte(cpuShare), 0644); err != nil {
			return fmt.Errorf("Set: cgroup cpushare %s setup failed. %v", cpuShare, err)
		} else {
			s.CPUShare = cpuShare
			log.Debugf("Set: cgroup cpushare %s setup successfully. ", cpuShare)
		}
	}  else {
		log.Debugf("Set: cgroup cpushare ignored!! ")
	}

	return nil

}

func (s *CpuSubSystem) Remove(cgroupPath string) error {
	return os.RemoveAll(s.SubsysCgroupPath)
}

func (s *CpuSubSystem) Apply(cgroupPath string, pid int) error {
	if s.CPUShare != "" {
		if err := ioutil.WriteFile(path.Join(s.SubsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("Apply: cgroup cpushare %s setup failed. %v", s.CPUShare, err)
		} else {
			log.Debugf("Apply: cgroup cpushare %s setup successfully. ", s.CPUShare)
			return nil
		}
	} else {
		log.Debugf("Apply: cgroup cpushare ignored!! ")
		return nil
	}

}

func (s *CpuSubSystem) Name() subSysType {
	return SUBSYSCPU
}
