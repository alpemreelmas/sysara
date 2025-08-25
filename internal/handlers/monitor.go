package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/alpemreelmas/sysara/internal/models"
	templ "github.com/alpemreelmas/sysara/templ"
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

// ShowMonitor displays the system monitoring dashboard
func (h *MonitorHandler) ShowMonitor(c *gin.Context) {
	currentUser, _ := c.Get("current_user")
	userModel, ok := currentUser.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
		return
	}

	data := templ.MonitorData{
		AuthData: templ.AuthData{
			Title:       "System Monitor - Sysara",
			PageTitle:   "System Monitor",
			CurrentUser: *userModel,
		},
	}
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)
	templ.Monitor(data).Render(c.Request.Context(), c.Writer)
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
		statsData := templ.SystemStatsData{
			Stats: *stats,
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusOK)
		templ.SystemStatsPartial(statsData).Render(c.Request.Context(), c.Writer)
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
		processData := templ.ProcessListData{
			Processes: processes,
		}
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusOK)
		templ.ProcessListPartial(processData).Render(c.Request.Context(), c.Writer)
		return
	}

	c.JSON(http.StatusOK, gin.H{"processes": processes})
}

// collectSystemStats gathers system statistics
func (h *MonitorHandler) collectSystemStats() (*templ.SystemStats, error) {
	stats := &templ.SystemStats{}

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
	// Use C:\ for Windows, / for Unix-like systems
	diskPath := "/"
	if runtime.GOOS == "windows" {
		diskPath = "C:\\"
	}
	diskStats, err := disk.Usage(diskPath)
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
func (h *MonitorHandler) collectProcessInfo() ([]templ.ProcessInfo, error) {
	pids, err := process.Pids()
	if err != nil {
		return nil, err
	}

	var processes []templ.ProcessInfo
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

		processes = append(processes, templ.ProcessInfo{
			PID:        pid,
			Name:       name,
			CPUPercent: cpuPercent,
			Memory:     memory,
			Status:     statusStr,
		})
	}

	return processes, nil
}
