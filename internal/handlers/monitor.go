package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// MonitorHandler handles system monitoring operations
type MonitorHandler struct{}

// NewMonitorHandler creates a new monitor handler
func NewMonitorHandler() *MonitorHandler {
	return &MonitorHandler{}
}

// SystemStats represents system statistics
type SystemStats struct {
	CPU     CPUStats     `json:"cpu"`
	Memory  MemoryStats  `json:"memory"`
	Disk    DiskStats    `json:"disk"`
	Network NetworkStats `json:"network"`
	Host    HostStats    `json:"host"`
}

type CPUStats struct {
	Usage     float64 `json:"usage"`
	Cores     int     `json:"cores"`
	ModelName string  `json:"model_name"`
}

type MemoryStats struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

type DiskStats struct {
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

type NetworkStats struct {
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
}

type HostStats struct {
	Hostname        string `json:"hostname"`
	Uptime          uint64 `json:"uptime"`
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	KernelVersion   string `json:"kernel_version"`
}

type ProcessInfo struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	CPUPercent float64 `json:"cpu_percent"`
	Memory     uint64  `json:"memory"`
	Status     string  `json:"status"`
}

// ShowMonitor displays the system monitoring dashboard
func (h *MonitorHandler) ShowMonitor(c *gin.Context) {
	currentUser, _ := c.Get("current_user")

	c.HTML(http.StatusOK, "pages/monitor/dashboard.html", gin.H{
		"Title":       "System Monitor - Sysara",
		"CurrentUser": currentUser,
	})
}

// GetSystemStats returns current system statistics (HTMX endpoint)
func (h *MonitorHandler) GetSystemStats(c *gin.Context) {
	stats, err := h.collectSystemStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to collect system stats"})
		return
	}

	// For HTMX requests, return HTML partial
	if c.GetHeader("HX-Request") == "true" {
		c.HTML(http.StatusOK, "partials/system-stats.html", gin.H{
			"Stats": stats,
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetProcesses returns current running processes (HTMX endpoint)
func (h *MonitorHandler) GetProcesses(c *gin.Context) {
	processes, err := h.collectProcessInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to collect process info"})
		return
	}

	// For HTMX requests, return HTML partial
	if c.GetHeader("HX-Request") == "true" {
		c.HTML(http.StatusOK, "partials/process-list.html", gin.H{
			"Processes": processes,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"processes": processes})
}

// collectSystemStats gathers system statistics
func (h *MonitorHandler) collectSystemStats() (*SystemStats, error) {
	stats := &SystemStats{}

	// CPU stats
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		stats.CPU.Usage = cpuPercent[0]
	}
	stats.CPU.Cores = runtime.NumCPU()

	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		stats.CPU.ModelName = cpuInfo[0].ModelName
	}

	// Memory stats
	memStats, err := mem.VirtualMemory()
	if err == nil {
		stats.Memory.Total = memStats.Total
		stats.Memory.Available = memStats.Available
		stats.Memory.Used = memStats.Used
		stats.Memory.UsedPercent = memStats.UsedPercent
	}

	// Disk stats (root partition)
	diskStats, err := disk.Usage("/")
	if err == nil {
		stats.Disk.Total = diskStats.Total
		stats.Disk.Free = diskStats.Free
		stats.Disk.Used = diskStats.Used
		stats.Disk.UsedPercent = diskStats.UsedPercent
	}

	// Network stats
	netStats, err := net.IOCounters(false)
	if err == nil && len(netStats) > 0 {
		stats.Network.BytesSent = netStats[0].BytesSent
		stats.Network.BytesRecv = netStats[0].BytesRecv
		stats.Network.PacketsSent = netStats[0].PacketsSent
		stats.Network.PacketsRecv = netStats[0].PacketsRecv
	}

	// Host stats
	hostInfo, err := host.Info()
	if err == nil {
		stats.Host.Hostname = hostInfo.Hostname
		stats.Host.Uptime = hostInfo.Uptime
		stats.Host.OS = hostInfo.OS
		stats.Host.Platform = hostInfo.Platform
		stats.Host.PlatformVersion = hostInfo.PlatformVersion
		stats.Host.KernelVersion = hostInfo.KernelVersion
	}

	return stats, nil
}

// collectProcessInfo gathers information about running processes
func (h *MonitorHandler) collectProcessInfo() ([]ProcessInfo, error) {
	pids, err := process.Pids()
	if err != nil {
		return nil, err
	}

	var processes []ProcessInfo
	maxProcesses := 20 // Limit to top 20 processes

	for i, pid := range pids {
		if i >= maxProcesses {
			break
		}

		proc, err := process.NewProcess(pid)
		if err != nil {
			continue
		}

		name, err := proc.Name()
		if err != nil {
			name = "Unknown"
		}

		cpuPercent, err := proc.CPUPercent()
		if err != nil {
			cpuPercent = 0
		}

		memInfo, err := proc.MemoryInfo()
		var memory uint64 = 0
		if err == nil {
			memory = memInfo.RSS
		}

		status, err := proc.Status()
		if err != nil {
			status = []string{"Unknown"}
		}

		statusStr := "Unknown"
		if len(status) > 0 {
			statusStr = status[0]
		}

		processes = append(processes, ProcessInfo{
			PID:        pid,
			Name:       name,
			CPUPercent: cpuPercent,
			Memory:     memory,
			Status:     statusStr,
		})
	}

	return processes, nil
}
