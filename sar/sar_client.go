package sar

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

func NewSarClient() *Client {
	return &Client{}
}

func (c *Client) Gather(sshConfig ssh_utils.SSHParams, ch chan<- prometheus.Metric, collector *sarCollector) {
	var wg sync.WaitGroup
	for _, host := range sshConfig.Hosts {
		wg.Add(1)
		go func(config ssh_utils.SSHParams, host string, ch chan<- prometheus.Metric, collector *sarCollector) {
			addr := host + ":" + strconv.Itoa(config.Port)
			output, err := ssh_utils.GetOutput(config.User, addr, config.Command, config.Key)
			if err != nil {
				log.Println("sar_client.Gather: ", host, err)
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

// Parse the sar file collected from the VM and add metric to Prometheus
func ParseAndAddMetrics(sData string, ch chan<- prometheus.Metric, collector *sarCollector) {
	csvTransform := regexp.MustCompile(`(?m)[\t\f\v ]+`).ReplaceAllString(sData, ",") // Replace all blocks of whitespace with comma.
	splitData := regexp.MustCompile(`(?m)^\n`).Split(csvTransform, -1)                // Split data by empty lines.

	hostname := strings.TrimSpace(splitData[0])
	for _, metrics := range splitData {
		if strings.Contains(metrics, "CPU") {
			rows := common.GetRows(metrics)
			if rows == nil {
				continue
			}
			addCPUMetrics(rows, hostname, ch, collector)
		} else if strings.Contains(metrics, "%commit") {
			rows := common.GetRows(metrics)
			if rows == nil {
				continue
			}
			addMemoryMetrics(rows, hostname, ch, collector)
		} else if strings.Contains(metrics, "rd_sec/s") {
			rows := common.GetRows(metrics)
			if rows == nil {
				continue
			}
			addStorageMetrics(rows, hostname, ch, collector)
		} else if strings.Contains(metrics, "rxpck/s") {
			rows := common.GetRows(metrics)
			if rows == nil {
				continue
			}
			addNetworkMetrics(rows, hostname, ch, collector)
		}
	}

}

// Parse and add CPU metrics.
func addCPUMetrics(rows []map[string]string, hostname string, ch chan<- prometheus.Metric, collector *sarCollector) {
	for _, row := range rows {
		// Not taking the average row , only individual times AND not taking individual cpu core values.
		if strings.Contains(row["CPU"], "all") && row["time"] != "Average:" {
			ch <- prometheus.MustNewConstMetric(collector.cpuIdle,
				prometheus.GaugeValue, common.String2float(row["%idle"]), hostname)
			ch <- prometheus.MustNewConstMetric(collector.cpuWait,
				prometheus.GaugeValue, common.String2float(row["%iowait"]), hostname)
			ch <- prometheus.MustNewConstMetric(collector.cpuUser,
				prometheus.GaugeValue, common.String2float(row["%usr"]), hostname)
			ch <- prometheus.MustNewConstMetric(collector.cpuSys,
				prometheus.GaugeValue, common.String2float(row["%sys"]), hostname)
			ch <- prometheus.MustNewConstMetric(collector.cpuSteal,
				prometheus.GaugeValue, common.String2float(row["%steal"]), hostname)

			break // exit after 1st metric as we cant have anymore. There should only ever be the 1 metric,
			// but in case there isn't break will save the exporter from throwing an error
		}
	}

}

// Parse and add memory metrics.
func addMemoryMetrics(rows []map[string]string, hostname string, ch chan<- prometheus.Metric, collector *sarCollector) {
	for _, row := range rows {
		if row["time"] != "Average:" { // Not taking the average row, only individual times.
			ch <- prometheus.MustNewConstMetric(collector.memFree,
				prometheus.GaugeValue, common.String2float(row["kbmemfree"]), hostname)
			ch <- prometheus.MustNewConstMetric(collector.memUsed,
				prometheus.GaugeValue, common.String2float(row["kbmemused"]), hostname)
			ch <- prometheus.MustNewConstMetric(collector.memPer,
				prometheus.GaugeValue, common.String2float(row["%memused"]), hostname)
			ch <- prometheus.MustNewConstMetric(collector.memBuffer,
				prometheus.GaugeValue, common.String2float(row["kbbuffers"]), hostname)
			ch <- prometheus.MustNewConstMetric(collector.memCache,
				prometheus.GaugeValue, common.String2float(row["kbcached"]), hostname)

			break // exit after 1st metric as we cant have anymore. There should only ever be the 1 metric,
			// but in case there isn't break will save the exporter from throwing an error

		}
	}

}

// Parse and add storage metrics.
func addStorageMetrics(rows []map[string]string, hostname string, ch chan<- prometheus.Metric, collector *sarCollector) {
	deviceList := make([]string, len(rows))
	for _, row := range rows {
		// Not taking the average row , only individual times .
		if row["time"] != "Average:" {
			deviceName := row["DEV"]
			if !common.Contains(deviceList, deviceName) {
				deviceList = append(deviceList, deviceName)
				ch <- prometheus.MustNewConstMetric(collector.storageTPS,
					prometheus.GaugeValue, common.String2float(row["tps"]), hostname, deviceName)
				ch <- prometheus.MustNewConstMetric(collector.storageRead,
					prometheus.GaugeValue, common.String2float(row["rd_sec/s"]), hostname, deviceName)
				ch <- prometheus.MustNewConstMetric(collector.storageWrite,
					prometheus.GaugeValue, common.String2float(row["wr_sec/s"]), hostname, deviceName)
				ch <- prometheus.MustNewConstMetric(collector.storageSector,
					prometheus.GaugeValue, common.String2float(row["avgrq-sz"]), hostname, deviceName)
				ch <- prometheus.MustNewConstMetric(collector.storageQueue,
					prometheus.GaugeValue, common.String2float(row["avgqu-sz"]), hostname, deviceName)
				ch <- prometheus.MustNewConstMetric(collector.storageAwait,
					prometheus.GaugeValue, common.String2float(row["await"]), hostname, deviceName)
				ch <- prometheus.MustNewConstMetric(collector.storageServiceTime,
					prometheus.GaugeValue, common.String2float(row["svctm"]), hostname, deviceName)
				ch <- prometheus.MustNewConstMetric(collector.storageUtilization,
					prometheus.GaugeValue, common.String2float(row["%util"]), hostname, deviceName)

			}

		}
	}
}

// Parse and add network metrics.
func addNetworkMetrics(rows []map[string]string, hostname string, ch chan<- prometheus.Metric, collector *sarCollector) {
	interfaceList := make([]string, len(rows))
	for _, row := range rows {
		// Not taking the average row , only individual times
		if row["time"] != "Average:" {
			networkInterface := row["IFACE"]
			if !common.Contains(interfaceList, networkInterface) {
				interfaceList = append(interfaceList, networkInterface)
				ch <- prometheus.MustNewConstMetric(collector.networkTxPacket,
					prometheus.GaugeValue, common.String2float(row["txpck/s"]), hostname, networkInterface)
				ch <- prometheus.MustNewConstMetric(collector.networkRxPacket,
					prometheus.GaugeValue, common.String2float(row["rxpck/s"]), hostname, networkInterface)
				ch <- prometheus.MustNewConstMetric(collector.networkRxKb,
					prometheus.GaugeValue, common.String2float(row["rxkB/s"]), hostname, networkInterface)
				ch <- prometheus.MustNewConstMetric(collector.networkTxKb,
					prometheus.GaugeValue, common.String2float(row["txkB/s"]), hostname, networkInterface)
				ch <- prometheus.MustNewConstMetric(collector.networkRxCompressed,
					prometheus.GaugeValue, common.String2float(row["rxcmp/s"]), hostname, networkInterface)
				ch <- prometheus.MustNewConstMetric(collector.networkTxCompressed,
					prometheus.GaugeValue, common.String2float(row["txcmp/s"]), hostname, networkInterface)
				ch <- prometheus.MustNewConstMetric(collector.networkRxMulticast,
					prometheus.GaugeValue, common.String2float(row["rxmcst/s"]), hostname, networkInterface)

			}
		}
	}
}
