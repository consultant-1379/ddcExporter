package amos

import (
	"ddcExporter/common"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Client struct {
}

func NewAmosClient() *Client {
	return &Client{}
}

func (c *Client) Gather(ch chan<- prometheus.Metric, collector *amosCollector) {
	currentDate := time.Now().Format("020106")
	amosDirs, _ := filepath.Glob(common.AMOSGlobPattern)
	t := time.Now()
	tz, _ := t.Zone()

	for _, dir := range amosDirs {
		path := fmt.Sprintf("%s/%s/%s", dir, currentDate, "instr.txt")
		lines := common.Tail(path, 50)
		for _, line := range lines {
			if strings.Contains(line, "terminal-websocket:type=VmMonitoringBean") {
				//08-06-20 11:13:00.023 ieatenmenv1-amos-0-com.ericsson.oss.presentation.server.terminal.vm.monitoring.terminal-websocket:type=VmMonitoringBean 0 14 0
				fields := strings.Split(line, " ")
				timestamp, err := time.Parse("02-01-06 15:04:05.000 MST", fields[0]+" "+fields[1]+" "+tz)
				if err != nil {
					log.Println("amos.client.gather: ", err)
					continue
				}
				hostname := strings.Replace(fields[2], "-com.ericsson.oss.presentation.server.terminal.vm.monitoring.terminal-websocket:type=VmMonitoringBean", "", -1)
				cpu := fields[3]
				memory := fields[4]
				sessions := fields[5]
				timestamp = timestamp.UTC() // Convert to UTC to avoid record too old exception in prometheus.
				ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(collector.memUsed,
					prometheus.GaugeValue, common.String2float(memory), hostname, timestamp.String()))
				ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(collector.cpuUsed,
					prometheus.GaugeValue, common.String2float(cpu), hostname, timestamp.String()))
				ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(collector.sessions,
					prometheus.GaugeValue, common.String2float(sessions), hostname, timestamp.String()))

				break
			}

		}

	}
}
