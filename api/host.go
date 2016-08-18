package api

type Host struct {
	Name          string
	Memory        uint64
	DiskSpace     uint64
	CpuUnits      uint64
	ReservedPorts []uint
	PortSelection []uint
}