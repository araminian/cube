package worker

import (
	"log"

	"github.com/c9s/goprocinfo/linux"
)

type Stats struct {
	MemStats  *linux.MemInfo
	DiskStats *linux.Disk
	CpuStats  *linux.CPUStat
	LoadStats *linux.LoadAvg
	TaskCount int
}

func getStats() Stats {
	return Stats{
		MemStats:  GetMemoryInfo(),
		DiskStats: GetDiskInfo(),
		CpuStats:  GetCpuStat(),
		LoadStats: GetLoadAvg(),
		TaskCount: 0,
	}
}

func GetMemoryInfo() *linux.MemInfo {
	memstats, err := linux.ReadMemInfo("/proc/meminfo")
	if err != nil {
		log.Println("Failed to read memory info:", err)
		return &linux.MemInfo{}
	}
	return memstats
}

func GetDiskInfo() *linux.Disk {
	diskstats, err := linux.ReadDisk("/")
	if err != nil {
		log.Println("Failed to read disk info:", err)
		return &linux.Disk{}
	}
	return diskstats
}

func GetCpuStat() *linux.CPUStat {
	stats, err := linux.ReadStat("/proc/stat")
	if err != nil {
		log.Println("Failed to read cpu stat:", err)
		return &linux.CPUStat{}
	}
	return &stats.CPUStatAll
}

func GetLoadAvg() *linux.LoadAvg {
	loadAvg, err := linux.ReadLoadAvg("/proc/loadavg")
	if err != nil {
		log.Println("Failed to read load avg:", err)
		return &linux.LoadAvg{}
	}
	return loadAvg
}

func (s *Stats) MemTotalKb() uint64 {
	return s.MemStats.MemTotal
}

func (s *Stats) MemAvailableKb() uint64 {
	return s.MemStats.MemAvailable
}

func (s *Stats) MemUsedKb() uint64 {
	return s.MemStats.MemTotal - s.MemStats.MemAvailable
}

func (s *Stats) MemUsedPercent() uint64 {
	return s.MemStats.MemAvailable / s.MemStats.MemTotal
}

func (s *Stats) DiskTotal() uint64 {
	return s.DiskStats.All
}

func (s *Stats) DiskUsed() uint64 {
	return s.DiskStats.Used
}

func (s *Stats) DiskFree() uint64 {
	return s.DiskStats.Free
}

func (s *Stats) CpuUsage() float64 {
	idle := s.CpuStats.Idle + s.CpuStats.IOWait
	nonIdle := s.CpuStats.User + s.CpuStats.Nice + s.CpuStats.System + s.CpuStats.IRQ + s.CpuStats.SoftIRQ + s.CpuStats.Steal
	total := idle + nonIdle
	if total == 0 {
		return 0.00
	}
	return (float64(total) - float64(idle)) / float64(total)
}
