package collector

import (
	"context"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/config"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/model"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

func collectSystem(ctx context.Context, cfg config.SystemConfig) (*model.SystemSample, []model.DiskSample, []model.NetworkSample, error) {
	var firstErr error
	setErr := func(err error) {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	system := &model.SystemSample{}
	if values, err := cpu.PercentWithContext(ctx, 0, false); err == nil && len(values) > 0 {
		system.CPUUsage = values[0]
	} else {
		setErr(err)
	}
	if vm, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		system.MemoryUsage = vm.UsedPercent
		system.MemoryUsed = vm.Used
		system.MemoryTotal = vm.Total
	} else {
		setErr(err)
	}
	if avg, err := load.AvgWithContext(ctx); err == nil {
		system.Load1 = avg.Load1
		system.Load5 = avg.Load5
		system.Load15 = avg.Load15
	} else {
		setErr(err)
	}
	if uptime, err := host.UptimeWithContext(ctx); err == nil {
		system.UptimeSeconds = uptime
	} else {
		setErr(err)
	}
	if info, err := host.InfoWithContext(ctx); err == nil {
		system.ProcessCount = info.Procs
		system.BootTime = info.BootTime
	} else {
		setErr(err)
	}

	disks := make([]model.DiskSample, 0, len(cfg.DiskMounts))
	for _, mount := range cfg.DiskMounts {
		usage, err := disk.UsageWithContext(ctx, mount)
		if err != nil {
			setErr(err)
			continue
		}
		disks = append(disks, model.DiskSample{
			Mount:      mount,
			Usage:      usage.UsedPercent,
			InodeUsage: usage.InodesUsedPercent,
			Used:       usage.Used,
			Total:      usage.Total,
		})
	}

	networks := []model.NetworkSample{}
	counters, err := net.IOCountersWithContext(ctx, true)
	if err != nil {
		setErr(err)
	} else {
		networks = make([]model.NetworkSample, 0, len(counters))
		for _, item := range counters {
			networks = append(networks, model.NetworkSample{
				Name:        item.Name,
				BytesSent:   item.BytesSent,
				BytesRecv:   item.BytesRecv,
				PacketsSent: item.PacketsSent,
				PacketsRecv: item.PacketsRecv,
			})
		}
	}
	return system, disks, networks, firstErr
}
