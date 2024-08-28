package amos

import "github.com/prometheus/client_golang/prometheus"

//08-06-20 00:00:12.980 CFG-ieatenmenv1-amos-0-com.ericsson.oss.presentation.server.terminal.vm.monitoring.terminal-websocket:type=VmMonitoringBean cpuUsed memoryUsed sessions
type amosCollector struct {
	cpuUsed  *prometheus.Desc
	memUsed  *prometheus.Desc
	sessions *prometheus.Desc
	client   *Client
}

func NewAmosCollector() *amosCollector {
	labels := []string{"hostname", "time"}
	return &amosCollector{
		cpuUsed:  prometheus.NewDesc("amos_cpu_used", "AMOS CPU usage from INSTR file.", labels, nil),
		memUsed:  prometheus.NewDesc("amos_mem_used", "AMOS MEM usage from INSTR file.", labels, nil),
		sessions: prometheus.NewDesc("amos_sessions", "AMOS sessions from INSTR file.", labels, nil),
	}

}

func (a *amosCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- a.sessions
	ch <- a.memUsed
	ch <- a.cpuUsed
}

func (a *amosCollector) Collect(ch chan<- prometheus.Metric) {
	a.client.Gather(ch, a)
}
