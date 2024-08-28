package top_process

import (
	ssh_utils "ddcExporter/ssh-utils"

	"github.com/prometheus/client_golang/prometheus"
)

type psCollector struct {
	topCPUProcess *prometheus.Desc
	topMemProcess *prometheus.Desc

	client    *Client
	sshConfig ssh_utils.SSHParams
}

func NewPsCollector(sshConfig ssh_utils.SSHParams) *psCollector {
	labels := []string{"hostname", "command", "user", "pid", "tty"}
	return &psCollector{
		// CPU
		topCPUProcess: prometheus.NewDesc("top_cpu_process", "Top process consuming CPU.", labels, nil),
		topMemProcess: prometheus.NewDesc("top_mem_process", "Top process consuming Memory.", labels, nil),

		client:    NewPsClient(),
		sshConfig: sshConfig,
	}
}

func (c *psCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.topCPUProcess
	ch <- c.topMemProcess
}

func (c *psCollector) Collect(ch chan<- prometheus.Metric) {
	c.client.Gather(c.sshConfig, ch, c)
}
