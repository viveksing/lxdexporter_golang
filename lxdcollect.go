package main

import (
	lxd "github.com/lxc/lxd/client"
	"github.com/prometheus/client_golang/prometheus"
)

//Define a struct for you collector that contains pointers
//to prometheus descriptors for each metric you wish to expose.
//Note you can also include fields of other types if they provide utility
//but we just won't be exposing them as metrics.
type lxdcollector struct {
	cpuUsage               *prometheus.Desc
	memUsage               *prometheus.Desc
	memUsagePeak           *prometheus.Desc
	swapUsage              *prometheus.Desc
	swapUsagePeak          *prometheus.Desc
	processCount           *prometheus.Desc
	diskUsage              *prometheus.Desc
	containerPid           *prometheus.Desc
	networkUsage           *prometheus.Desc
	containerRunningStatus *prometheus.Desc
}

//You must create a constructor for you collector that
//initializes every descriptor and returns a pointer to the collector
func newLxdCollector() *lxdcollector {
	return &lxdcollector{
		cpuUsage: prometheus.NewDesc("lxd_container_cpu_usage",
			"Container Cpu Usage in Seconds",
			[]string{"container_name"}, nil,
		),
		memUsage: prometheus.NewDesc("lxd_container_mem_usage",
			"Container Memory Usage",
			[]string{"container_name"}, nil,
		),
		memUsagePeak: prometheus.NewDesc("lxd_container_mem_usage_peak",
			"Container Memory Usage Peak",
			[]string{"container_name"}, nil,
		),
		swapUsage: prometheus.NewDesc("lxd_container_swap_usage",
			"Container Swap Usage",
			[]string{"container_name"}, nil,
		),
		swapUsagePeak: prometheus.NewDesc("lxd_container_swap_usage_peak",
			"Container Swap Usage Peak",
			[]string{"container_name"}, nil,
		),
		processCount: prometheus.NewDesc("lxd_container_process_count",
			"Container number of process Running",
			[]string{"container_name"}, nil,
		),
		diskUsage: prometheus.NewDesc("lxd_container_disk_usage",
			"Container Disk Usage",
			[]string{"container_name", "disk_device"}, nil,
		),
		containerPid: prometheus.NewDesc("lxd_container_pid",
			"Container PID",
			[]string{"container_name"}, nil,
		),
		networkUsage: prometheus.NewDesc("lxd_container_network_usage",
			"Container Network Usage",
			[]string{"container_name", "interface", "operation"}, nil,
		),
		containerRunningStatus: prometheus.NewDesc("lxd_container_running_status",
			"Container Running Status",
			[]string{"container_name"}, nil,
		),
	}
}

//Each and every collector must implement the Describe function.
//It essentially writes all descriptors to the prometheus desc channel.
func (collector *lxdcollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.cpuUsage
	ch <- collector.memUsage
	ch <- collector.memUsagePeak
	ch <- collector.swapUsage
	ch <- collector.swapUsagePeak
	ch <- collector.processCount
	ch <- collector.diskUsage
	ch <- collector.containerPid
	ch <- collector.networkUsage
	ch <- collector.containerRunningStatus

}

//Collect implements required collect function for all promehteus collectors
func (collector *lxdcollector) Collect(ch chan<- prometheus.Metric) {

	//Implement logic here to determine proper metric value to return to prometheus
	//for each descriptor or call other functions that do so.
	connection, _ := lxd.ConnectLXDUnix("", nil)
	names, _ := connection.GetContainerNames()
	var (
		cpuUsageValue      float64
		memUsageValue      float64
		memUsagePeakValue  float64
		swapUsageValue     float64
		swapUsagePeakValue float64
		processCountValue  float64
		diskUsageValue     float64
		containerPidValue  float64
	)

	for _, name := range names {
		state, _, _ := connection.GetContainerState(name)
		//	fmt.Println(state)
		cpuUsageValue = float64(state.CPU.Usage)
		memUsageValue = float64(state.Memory.Usage)
		memUsagePeakValue = float64(state.Memory.UsagePeak)
		swapUsageValue = float64(state.Memory.SwapUsage)
		swapUsagePeakValue = float64(state.Memory.SwapUsagePeak)
		processCountValue = float64(state.Processes)
		containerPidValue = float64(state.Pid)
		//	containerRunningStatusValue = float64(state.Status)
		//Write latest value for each metric in the prometheus metric channel.
		//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
		ch <- prometheus.MustNewConstMetric(collector.cpuUsage, prometheus.GaugeValue, cpuUsageValue, name)
		ch <- prometheus.MustNewConstMetric(collector.memUsage, prometheus.GaugeValue, memUsageValue, name)
		ch <- prometheus.MustNewConstMetric(collector.memUsagePeak, prometheus.GaugeValue, memUsagePeakValue, name)
		ch <- prometheus.MustNewConstMetric(collector.swapUsage, prometheus.GaugeValue, swapUsageValue, name)
		ch <- prometheus.MustNewConstMetric(collector.swapUsagePeak, prometheus.GaugeValue, swapUsagePeakValue, name)
		ch <- prometheus.MustNewConstMetric(collector.processCount, prometheus.GaugeValue, processCountValue, name)
		ch <- prometheus.MustNewConstMetric(collector.containerPid, prometheus.GaugeValue, containerPidValue, name)

		for key, value := range state.Disk {
			diskUsageValue = float64(value.Usage)
			ch <- prometheus.MustNewConstMetric(collector.diskUsage, prometheus.GaugeValue, diskUsageValue, name, key)
		}

		for key, value := range state.Network {
			operations := map[string]float64{
				"BytesReceived":   float64(value.Counters.BytesReceived),
				"BytesSent":       float64(value.Counters.BytesSent),
				"PacketsReceived": float64(value.Counters.PacketsReceived),
				"PacketsSent":     float64(value.Counters.PacketsSent),
			}

			for i, j := range operations {
				ch <- prometheus.MustNewConstMetric(collector.networkUsage, prometheus.GaugeValue, j, name, key, i)
			}
		}

		if state.Status == "Running" {
			ch <- prometheus.MustNewConstMetric(collector.containerRunningStatus, prometheus.GaugeValue, 1, name)
		}
	}
}
