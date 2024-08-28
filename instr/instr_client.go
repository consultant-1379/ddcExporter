package instr

import (
    "ddcExporter/common"
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "time"

    "github.com/prometheus/client_golang/prometheus"
)

type Client struct {
}

func NewInstrClient() *Client {
    return &Client{}
}

func (c *Client) Gather(ch chan<- prometheus.Metric, collector *InstrCollector) {
    currentDate := time.Now().Format("020106")
    dirs, _ := filepath.Glob(common.DDCLocation + "*")

    for _, dir := range dirs {
        fi, _ := os.Lstat(dir)
        if fi.IsDir() {
            folders := strings.Split(dir, "/")
            hostName := strings.Replace(folders[len(folders)-1], "_TOR", "", -1)
            path := fmt.Sprintf("%s/%s/%s", dir, currentDate, "instr.txt")
            foundSubMetrics := map[string]bool{}
            lines := common.Tail(path, 50)
            for _, line := range lines {
                if !strings.Contains(line, "CFG-") {
                    collect(ch, collector, line, hostName, foundSubMetrics)
                }
            }
        }
    }
}

func collect(ch chan<- prometheus.Metric, collector *InstrCollector, line string, hostName string, found map[string]bool) {
    for _, metric := range collector.metrics {
        _, ok := found[metric.displayName]
        if !ok && strings.Contains(line, metric.Name) {
            found[metric.displayName] = true
            fields := strings.Split(line, " ")
            matchAllowed, _ := regexp.MatchString(metric.AllowedHostPattern, hostName)
            matchNotAllowed, _ := regexp.MatchString(metric.NotAllowedHostPattern, hostName)
            if len(metric.NotAllowedHostPattern) == 0 {
                matchNotAllowed = false
            }
            if matchAllowed && !matchNotAllowed {
                for _, att := range metric.attributes {
                    ch <- prometheus.MustNewConstMetric(att.prometheusDesc, prometheus.GaugeValue,
                        common.String2float(fields[att.instrIdx]), hostName)
                }
            }
        }
    }
}
