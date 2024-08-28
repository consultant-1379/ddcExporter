package sar

import (
	ssh_utils "ddcExporter/ssh-utils"

	"github.com/prometheus/client_golang/prometheus"
)

type sarCollector struct {
	cpuIdle  *prometheus.Desc
	cpuSys   *prometheus.Desc
	cpuUser  *prometheus.Desc
	cpuWait  *prometheus.Desc
	cpuSteal *prometheus.Desc

	memUsed   *prometheus.Desc
	memFree   *prometheus.Desc
	memBuffer *prometheus.Desc
	memCache  *prometheus.Desc
	memPer    *prometheus.Desc

	storageTPS         *prometheus.Desc
	storageRead        *prometheus.Desc
	storageWrite       *prometheus.Desc
	storageSector      *prometheus.Desc
	storageQueue       *prometheus.Desc
	storageAwait       *prometheus.Desc
	storageServiceTime *prometheus.Desc
	storageUtilization *prometheus.Desc

	networkRxPacket     *prometheus.Desc
	networkTxPacket     *prometheus.Desc
	networkRxKb         *prometheus.Desc
	networkTxKb         *prometheus.Desc
	networkRxCompressed *prometheus.Desc
	networkTxCompressed *prometheus.Desc
	networkRxMulticast  *prometheus.Desc

	client    *Client
	sshConfig ssh_utils.SSHParams
}

func NewSarCollector(sshConfig ssh_utils.SSHParams) *sarCollector {
	labels := []string{"hostname"}
	labelsWithInterface := append(labels, "interface")
	labelsWithDevice := append(labels, "device")
	return &sarCollector{
		// CPU
		cpuIdle:  prometheus.NewDesc("cpu_usage_idle", "Idle CPU usage from SARS file.", labels, nil),
		cpuUser:  prometheus.NewDesc("cpu_usage_user", "User CPU usage from SARS file.", labels, nil),
		cpuSys:   prometheus.NewDesc("cpu_usage_sys", "Sys CPU usage from SARS file.", labels, nil),
		cpuWait:  prometheus.NewDesc("cpu_usage_wait", "Wait CPU usage from SARS file.", labels, nil),
		cpuSteal: prometheus.NewDesc("cpu_usage_steal", "Steal CPU usage from SARS file.", labels, nil),

		//MEM
		memUsed:   prometheus.NewDesc("mem_used_kb", "Memory used in kb from SARs file", labels, nil),
		memFree:   prometheus.NewDesc("mem_free_kb", "Memory free in kb from SARs file", labels, nil),
		memBuffer: prometheus.NewDesc("mem_buffers_kb", "Memory buffers in kb from SARs file", labels, nil),
		memCache:  prometheus.NewDesc("mem_cached_kb", "Memory cached in kb from SARs file", labels, nil),
		memPer:    prometheus.NewDesc("mem_used_percent", "Memory used in percent from SARs file", labels, nil),

		//STORAGE
		storageAwait:       prometheus.NewDesc("storage_await", "Storage await from SARs file", labelsWithDevice, nil),
		storageQueue:       prometheus.NewDesc("storage_queue", "Storage queue from SARs file", labelsWithDevice, nil),
		storageRead:        prometheus.NewDesc("storage_read_seconds", "Storage read/second from SARs file", labelsWithDevice, nil),
		storageSector:      prometheus.NewDesc("storage_sector_average", "Storage sector average from SARs file", labelsWithDevice, nil),
		storageServiceTime: prometheus.NewDesc("storage_service_time", "Storage service time from SARs file", labelsWithDevice, nil),
		storageTPS:         prometheus.NewDesc("storage_transaction_seconds", "Storage transactions/second from SARs file", labelsWithDevice, nil),
		storageUtilization: prometheus.NewDesc("storage_utils_percent", "Storage utilization percentage from SARs file", labelsWithDevice, nil),
		storageWrite:       prometheus.NewDesc("storage_write_seconds", "Storage write/second from SARs file", labelsWithDevice, nil),

		//NETWORK
		networkRxPacket:     prometheus.NewDesc("network_rx_packets", "Network receive packets per second  from SARs file", labelsWithInterface, nil),
		networkTxPacket:     prometheus.NewDesc("network_tx_packets", "Network transmit packets per second from SARs file", labelsWithInterface, nil),
		networkRxKb:         prometheus.NewDesc("network_rx_kb", "Network receive kb per second from SARs file", labelsWithInterface, nil),
		networkTxKb:         prometheus.NewDesc("network_tx_kb", "Network transmit kb per second from SARs file", labelsWithInterface, nil),
		networkRxCompressed: prometheus.NewDesc("network_rx_compressed", "Network receive compressed per second from SARs file", labelsWithInterface, nil),
		networkTxCompressed: prometheus.NewDesc("network_tx_compressed", "Network transmit compressed per second from SARs file", labelsWithInterface, nil),
		networkRxMulticast:  prometheus.NewDesc("network_rx_multicast", "Network receive multicast per second from SARs file", labelsWithInterface, nil),

		client:    NewSarClient(),
		sshConfig: sshConfig,
	}
}

func (c *sarCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.cpuIdle
	ch <- c.cpuSys
	ch <- c.cpuUser
	ch <- c.cpuWait

	ch <- c.memFree
	ch <- c.memUsed
	ch <- c.memPer

	ch <- c.storageAwait
	ch <- c.storageQueue
	ch <- c.storageTPS
	ch <- c.storageServiceTime
	ch <- c.storageSector
	ch <- c.storageRead
	ch <- c.storageUtilization
	ch <- c.storageWrite

	ch <- c.networkRxCompressed
	ch <- c.networkTxCompressed
	ch <- c.networkRxMulticast
	ch <- c.networkRxKb
	ch <- c.networkTxKb
	ch <- c.networkRxPacket
	ch <- c.networkTxPacket
}

func (c *sarCollector) Collect(ch chan<- prometheus.Metric) {
	c.client.Gather(c.sshConfig, ch, c)
}
