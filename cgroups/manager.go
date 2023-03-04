package cgroups

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/yang-wang11/mydocker/common"
)

type CgroupManager struct {
	Path     string                 // hierarchy path of cgroup
	Resource *common.ResourceConfig // resource setting
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

// apply cgroup setting
func (c *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range SubsystemsIns {
		err := subSysIns.Apply(c.Path, pid)
		if err != nil {
			logrus.Warnf("%s call apply function failed, %v ", subSysIns.Name(), err.Error())
		}
	}
	return nil
}

// cgroup restriction setting
func (c *CgroupManager) Set(res *common.ResourceConfig) error {
	var err error
	for _, subSysIns := range SubsystemsIns {
		switch subSysIns.Name() {
		case SUBSYSCPU:
			err = subSysIns.Set(c.Path, res.CpuShare)
		case SUBSYSCPUSET:
			err = subSysIns.Set(c.Path, res.CpuSet)
		case SUBSYSMEMORY:
			err = subSysIns.Set(c.Path, res.MemoryLimit)
		default:
			panic(errors.New(fmt.Sprintf("unknown %s call Set function. ", subSysIns.Name())))
		}
		if err != nil {
			logrus.Warnf("%s call Set function failed, %v ", subSysIns.Name(), err.Error())
		}
	}
	return nil
}

// revoke cgroup setting
func (c *CgroupManager) Destroy() error {
	for _, subSysIns := range SubsystemsIns {
		if err := subSysIns.Remove(c.Path); err != nil {
			logrus.Warnf("%s call remove function failed, %v ", subSysIns.Name(), err.Error())
		}
	}
	return nil
}
