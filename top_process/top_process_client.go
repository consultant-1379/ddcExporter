package top_process

import (
	"ddcExporter/common"
	ssh_utils "ddcExporter/ssh-utils"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type Client struct {
}

func NewPsClient() *Client {
	return &Client{}
}

func (c *Client) Gather(sshConfig ssh_utils.SSHParams, ch chan<- prometheus.Metric, collector *psCollector) {
	var wg sync.WaitGroup
	for _, host := range sshConfig.Hosts {
		if strings.Contains(host, "amos") || strings.Contains(host, "scp") {
			wg.Add(1)
			go func(config ssh_utils.SSHParams, host string, ch chan<- prometheus.Metric, collector *psCollector) {
				addr := host + ":" + strconv.Itoa(config.Port)
				output, err := ssh_utils.GetOutput(config.User, addr, config.Command, config.Key)
				if err != nil {
					log.Println("ps_client.Gather: ", host, err)
					wg.Done()
					return
				}
				output = strings.TrimSpace(output)
				ParseAndAddMetrics(output, ch, collector)
				wg.Done()
			}(sshConfig, host, ch, collector)
		}
		wg.Wait()
	}
}

// Parse the ps command output according to mem / cpu metric type.
func ParseAndAddMetrics(sData string, ch chan<- prometheus.Metric, collector *psCollector) {
	csvTransform := regexp.MustCompile(`(?m)[\t\f\v ]+`).ReplaceAllString(sData, ",") // Replace all blocks of whitespace with comma.
	splitData := regexp.MustCompile(`(?m)^\n`).Split(csvTransform, -1)                // Split data by empty lines.

	hostname := strings.TrimSpace(splitData[0])

	for _, metrics := range splitData {
		if strings.Contains(metrics, "CPU") {
			rows := common.GetRows(metrics)
			if rows == nil {
				continue
			}
			addMetrics(rows, hostname, ch, collector, "CPU")
		} else if strings.Contains(metrics, "MEM") {
			rows := common.GetRows(metrics)
			if rows == nil {
				continue
			}
			addMetrics(rows, hostname, ch, collector, "MEM")
		}
	}

}

// Parse and add CPU or Memory metrics.
func addMetrics(rows []map[string]string, hostname string, ch chan<- prometheus.Metric, collector *psCollector, metricType string) {
	for _, row := range rows {
		command := row["COMMAND"]
		user := row["USER"]
		pid := row["PID"]
		tty := row["TT"]
		if command != "" && user != "" && pid != "" && tty != "" {
			if metricType == "CPU" {
				ch <- prometheus.MustNewConstMetric(collector.topCPUProcess, prometheus.GaugeValue, common.String2float(row["%CPU"]), hostname, command, user, pid, tty)
			} else if metricType == "MEM" {
				ch <- prometheus.MustNewConstMetric(collector.topMemProcess, prometheus.GaugeValue, common.String2float(row["%MEM"]), hostname, command, user, pid, tty)
			}
		} else {
			log.Println("Top Process collection received nil label values for " + hostname)
		}
	}
}
