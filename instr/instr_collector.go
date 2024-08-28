package instr

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/spf13/viper"
)

type Attribute struct {
    instrIdx       int
    prometheusDesc *prometheus.Desc
}

type Metric struct {
    AllowedHostPattern    string
    NotAllowedHostPattern string
    Name                  string
    displayName           string
    attributes            []Attribute
}

type InstrCollector struct {
    metrics map[string]Metric
    client  *Client
}

func NewInstrCollector(permissions map[string]Metric) *InstrCollector {
    labels := []string{"hostname"}
    jvmMemory := make([]Attribute, 0)
    cpu := make([]Attribute, 0)
    pmFileCollection := make([]Attribute, 0)
    dpMediationFileCollection := make([]Attribute, 0)
    dpProxyFileCollection := make([]Attribute, 0)

    jvmMemory = append(jvmMemory,
        Attribute{4, prometheus.NewDesc("jvm_memory_heap_memory_usage_committed", "JVM Heap Memory Commited usage from INSTR file.", labels, nil)},
        Attribute{7, prometheus.NewDesc("jvm_memory_heap_memory_usage_used", "JVM Heap Memory Used usage from INSTR file.", labels, nil)},
    )
    cpu = append(cpu,
        Attribute{5, prometheus.NewDesc("cpu_process_cpu_load", "CPU Process CUP Load from INSTR file.", labels, nil)},
    )

    jvmMemoryANP, _ := permissions["jvmmemory"]
    cpuANP, _ := permissions["cpu"]
    metrics := map[string]Metric{
        "jvmMemory": {jvmMemoryANP.AllowedHostPattern, jvmMemoryANP.NotAllowedHostPattern, "-jvm-memory", "jvm-memory", jvmMemory},
        "cpu":       {cpuANP.AllowedHostPattern, cpuANP.NotAllowedHostPattern, "-os", "cpu", cpu},
    }

    if viper.GetBool("metrics.pm") {
        pmFileCollection = append(pmFileCollection,
            Attribute{7, prometheus.NewDesc("pmserv_files_collected", "pmserv file collected in current rop from INSTR file.", labels, nil)},
            Attribute{9, prometheus.NewDesc("pmserv_files_failed", "pmserv file failed in current rop from INSTR file.", labels, nil)},
            Attribute{11, prometheus.NewDesc("pmserv_rop_collection_time", "pmserv rop collection time from INSTR file.", labels, nil)},
        )
        pmFileCollectionANP, _ := permissions["pmFileCollection"]
        metrics["pmFileCollection"] = Metric{pmFileCollectionANP.AllowedHostPattern, pmFileCollectionANP.NotAllowedHostPattern, "=FileCollectionInstrumentation", "pmFileCollection", pmFileCollection}
    }

    if viper.GetBool("metrics.dpmediation") {
        dpMediationFileCollection = append(dpMediationFileCollection,
            Attribute{4, prometheus.NewDesc("dp_failed_attempts_with_sas", "numberOfFailedAttempsWithSas from INSTR file.", labels, nil)},
            Attribute{6, prometheus.NewDesc("dp_total_number_of_hb_to_sas", "totalNumberOfHbToSAS Total number of attempts in contacting SAS from INSTR file.", labels, nil)},
        )
        dpMediationFileCollectionANP, _ := permissions["dpmediationfilecollection"]
        metrics["dpMediationFileCollection"] = Metric{dpMediationFileCollectionANP.AllowedHostPattern, dpMediationFileCollectionANP.NotAllowedHostPattern, "=DpmediationSasClientInstrumentationBean", "dpMediationFileCollection", dpMediationFileCollection}
    }

    if viper.GetBool("metrics.domainproxy") {
        dpProxyFileCollection = append(dpProxyFileCollection,
            Attribute{25, prometheus.NewDesc("dp_grant_request_count", "grantRequestsCount from INSTR file.", labels, nil)},
            Attribute{27, prometheus.NewDesc("dp_hb_request_count", "heartbeatRequestsCount from INSTR file.", labels, nil)},
            Attribute{36, prometheus.NewDesc("dp_active_cells_count", "numberOfActiveCellsCount from INSTR file.", labels, nil)},
            Attribute{40, prometheus.NewDesc("dp_failed_connection_attempts_with_sas_incremental", "numberOfFailedConnectionAttemptsWithSasIncremental from INSTR file.", labels, nil)},
            Attribute{41, prometheus.NewDesc("dp_inactive_cells_count", "numberOfInactiveCellsCount from INSTR file.", labels, nil)},
            Attribute{42, prometheus.NewDesc("dp_maintained_grants_count", "numberOfMaintainedGrantsCount from INSTR file.", labels, nil)},
            Attribute{68, prometheus.NewDesc("dp_registration_requests_count", "registrationRequestsCount from INSTR file.", labels, nil)},
            Attribute{72, prometheus.NewDesc("dp_relinquishment_requests_count", "relinquishmentRequestsCount from INSTR file.", labels, nil)},
            Attribute{78, prometheus.NewDesc("dp_spectrum_inquiry_request_count", "spectrumInquiryRequestsCount from INSTR file.", labels, nil)},
        )
        dpProxyFileCollectionANP, _ := permissions["dpproxyfilecollection"]
        metrics["dpProxyFileCollection"] = Metric{dpProxyFileCollectionANP.AllowedHostPattern, dpProxyFileCollectionANP.NotAllowedHostPattern, "ericsson.oss.sas.common.metrics.domain-proxy-service:type=DomainProxyInstrumentation", "dpProxyFileCollection", dpProxyFileCollection}
    }

    return &InstrCollector{
        metrics: metrics,
    }

}

func (i *InstrCollector) Describe(ch chan<- *prometheus.Desc) {
    for _, met := range i.metrics {
        for _, att := range met.attributes {
            ch <- att.prometheusDesc
        }
    }
}

func (i *InstrCollector) Collect(ch chan<- prometheus.Metric) {
    i.client.Gather(ch, i)
}
