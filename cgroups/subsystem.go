package cgroups

// subsystem function
type Subsystem interface {
	Name() subSysType
	Set(path string, res interface{}) error
	Apply(path string, pid int) error
	Remove(path string) error
}

type subSysType string ;

const (
	SUBSYSCPU    subSysType = "cpu";
	SUBSYSCPUSET subSysType = "cpuset";
	SUBSYSMEMORY subSysType = "memory";
)


var (
	SubsystemsIns = []Subsystem{
		&CpusetSubSystem{},
		&MemorySubSystem{},
		&CpuSubSystem{},
	}
)
