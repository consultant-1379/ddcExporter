package main

import (
    "ddcExporter/amos"
    "ddcExporter/instr"
    "ddcExporter/sar"
    ssh_utils "ddcExporter/ssh-utils"
    "ddcExporter/top_process"
    "flag"
    "io/ioutil"
    "log"
    "net/http"
    "path/filepath"
    "strconv"
    "strings"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/spf13/viper"
)

var (
    configLocation string
)

func init() {
    flag.StringVar(&configLocation, "config", "config.json", "Config File Location")
    flag.Parse()
    initViperConfig()
}

func initViperConfig() {
    dir, file := filepath.Split(configLocation)
    viper.SetConfigName(file)
    viper.SetConfigType("json")
    viper.AddConfigPath(dir)
    viper.AddConfigPath(".")
    viper.SetDefault("ssh.user", "cloud-user")
    viper.SetDefault("ssh.port", 22)
    viper.SetDefault("ssh.key", "./key.pem")
    viper.SetDefault("ssh.hosts", "./hosts.txt")
    viper.SetDefault("http.port", 9010)
    viper.SetDefault("instr.cpu.AllowedHostPattern", ".*")
    viper.SetDefault("instr.cpu.NotAllowedHostPattern", "")
    viper.SetDefault("instr.jvmMemory.AllowedHostPattern", ".*")
    viper.SetDefault("instr.jvmMemory.NotAllowedHostPattern", "")
    viper.SetDefault("instr.pmFileCollection.AllowedHostPattern", ".*pmserv.*")
    viper.SetDefault("instr.pmFileCollection.NotAllowedHostPattern", "")
    viper.SetDefault("instr.dpMediation.AllowedHostPattern", ".*dpmediation.*")
    viper.SetDefault("instr.dpMediation.NotAllowedHostPattern", "")
    viper.SetDefault("instr.domainProxy.AllowedHostPattern", ".*domainproxy.*")
    viper.SetDefault("instr.domainProxy.NotAllowedHostPattern", "")
    viper.SetDefault("metrics.sar", true)
    viper.SetDefault("metrics.amos", true)
    viper.SetDefault("metrics.instr", true)
    viper.SetDefault("metrics.ps", true)
    viper.SetDefault("metrics.pm", true)
    viper.SetDefault("metrics.dpMediation", true)
    viper.SetDefault("metrics.domainProxy", true)

    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            log.Fatal("main.initViperConfig - Config file not found: ", err)
        } else {
            log.Fatal("main.initViperConfig - Error reading config file: ", err)
        }
    }

}

func getInstrPermission() map[string]instr.Metric {
    var permissions map[string]instr.Metric
    err := viper.UnmarshalKey("instr", &permissions)
    if err != nil {
        log.Fatal("main.main - Error parsing instr config json: ", err)
    }
    return permissions
}

func registerSarCollector(hosts []string, keyFileBuffer []byte) {
    if viper.GetBool("metrics.sar") {

        // Get hostname as well because sar file can give host as cloudyone.localdomain in certain conditions.
        command := "hostname && echo && LC_TIME=\"C\" sar -A -p -f /var/log/sa/sa$(date +\"%d\") -s $(date +\"%T\" -d \"2 mins ago\") -e $(date +\"%T\")"
        sshConfig := ssh_utils.SSHParams{
            User:    viper.GetString("ssh.user"),
            Key:     keyFileBuffer,
            Hosts:   hosts,
            Port:    viper.GetInt("ssh.port"),
            Command: command,
        }

        sarCol := sar.NewSarCollector(sshConfig)
        prometheus.MustRegister(sarCol)

    }
}

func registerPsCollector(hosts []string, keyFileBuffer []byte) {
    if viper.GetBool("metrics.ps") {
        // Get hostname and use custom sort for memory as the ps command doesn't sort properly in scp and amos VMs - weird !
        command := "hostname && echo && ps -Ao user:32,comm,pid,tty,pcpu --sort -pcpu | head -n 10 && echo && ps -Ao user:32,comm,pid,tty,pmem | sed -e1\\!b -e'w /dev/fd/2' -ed |sort -n -r -k5,5 | head -n 10"
        sshConfig := ssh_utils.SSHParams{
            User:    viper.GetString("ssh.user"),
            Key:     keyFileBuffer,
            Hosts:   hosts,
            Port:    viper.GetInt("ssh.port"),
            Command: command,
        }

        psCol := top_process.NewPsCollector(sshConfig)
        prometheus.MustRegister(psCol)

    }
}

func registerAmosCollector() {
    if viper.GetBool("metrics.amos") {
        amosCol := amos.NewAmosCollector()
        prometheus.MustRegister(amosCol)
    }
}

func registerInstrCollector() {
    if viper.GetBool("metrics.instr") {
        instrCol := instr.NewInstrCollector(getInstrPermission())
        prometheus.MustRegister(instrCol)
    }
}

func main() {

    HostFileBuffer, err := ioutil.ReadFile(viper.GetString("ssh.hosts"))
    if err != nil {
        log.Fatal("main.main - Error reading host file: ", err)
    }
    hosts := strings.Split(strings.TrimSuffix(string(HostFileBuffer), "\n"), "\n")

    KeyFileBuffer, err := ioutil.ReadFile(viper.GetString("ssh.key"))
    if err != nil {
        log.Fatal("main.main - Error reading key file: ", err)
    }
    registerSarCollector(hosts, KeyFileBuffer)
    registerAmosCollector()
    registerInstrCollector()
    registerPsCollector(hosts, KeyFileBuffer)

    http.Handle("/metrics", promhttp.Handler())

    listenAddr := ":" + strconv.Itoa(viper.GetInt("http.port"))
    log.Println("Starting server on port: " + listenAddr)
    log.Fatal(http.ListenAndServe(listenAddr, nil))

}
