#DDC SAR Exporter

DDC currently collects and pushes metrics every 4 hours. This exporter will collect those metrics on demand and expose it in a prometheus friendly way.

For CPU/Mem/Storage/Network we SSH into the VMs given in the host file and run the SAR command to fetch data.
For Top CPU/Mem consuming process we SSH only into the "amos and scp" VMs given in the host file and run the ps command to fetch data. The host file must have amos and scp VMs in the list for collection to work.
For AMOS we read the instr.txt file for AMOS VMs.

This is supposed to run inside an ENM VM with external networking (e.g EMP/SCP).
###Usage
```shell script
$ ./ddcExporter --help
```
    Usage of ddcExporter:
      -config string
            Config File Location (default "config.json")
    
### Example Config File
These values represent the default values in case a field in missing from the config file. The pm attribute in metrics atrribute of the config file provides the capability to toggle PM File collection START and STOP. 
```json
{
  "ssh": {
    "user": "cloud-user",
    "port": 22,
    "key": "./key.pem",
    "hosts": "./hosts.txt" 
  },
  "http": {
    "port": 9010
  },
  "instr": {
    "cpu": {
      "AllowedHostPattern": ".*",
      "NotAllowedHostPattern": ""
    },
    "jvmMemory": {
      "AllowedHostPattern": ".*",
      "NotAllowedHostPattern": ""
    }
  },
  "metrics": {
    "sar": true,
    "amos": true,
    "instr": true,
    "ps": true,
    "pm": true,
    "dpMediation": true,
    "domainProxy": true
  }
}
```
| Key          | Description                                             |
|--------------|---------------------------------------------------------|
| ssh.user     | SSH User name                                           |
| ssh.port     | SSH Port                                                |
| ssh.key      | Path to SSH key file                                    |
| ssh.hosts    | Path to file containing new line separated hostnames/IP |
| http.port    | Listen port for the exporter                            |
| instr        | See detailed description below                          |
| metrics.*    | Enable/Diable metric collection for that type           |
####Instr Json config
Instr config will be used to control relevant VMs for metrics.
```AllowedHostPattern``` and ```NotAllowedHostPattern``` are mandatory fields. Empty strings are allowed for these fields.

#####Instr Config examples
```json
{
 "cpu": {
  "AllowedHostPattern": ".*",
  "NotAllowedHostPattern": ""
 },
 "jvmMemory": {
  "AllowedHostPattern": ".*",
  "NotAllowedHostPattern": ""
 }
}
```
Exposed cpu and jvm_memory for all VMs. 

```json
{
 "cpu": {
  "AllowedHostPattern": "(amos.*|fmx.*)",
  "NotAllowedHostPattern": ""
 },
...
}
```
Exposed cpu metrics only for ```amos``` and ```fmx``` VMS

```json
{
 "cpu": {
  "AllowedHostPattern": ".*",
  "NotAllowedHostPattern": "(amos.*|fmx.*)"
 },
...
}
```
Exposed cpu metrics only for every VMs except ```amos``` and ```fmx``` VMS

```json
...
},
  "metrics": {
    ...
    "pm": true
  }
}
```
Start File Collection

```json
...
},
  "metrics": {
    ...
    "pm": false
  }
}
```
Stop File Collection


###Metrics Exposed
#### CPU
    cpu_usage_idle - Idle CPU usage from SARS file.
	cpu_usage_user - User CPU usage from SARS file.
	cpu_usage_sys - Sys CPU usage from SARS file.
	cpu_usage_wait - Wait CPU usage from SARS file.
	cpu_usage_steal - Steal CPU usage from SARS file.
#### Memory
	mem_used_kb - Memory used in kb from SARs file
	mem_free_kb - Memory free in kb from SARs file
	mem_used_percent - Memory used in percent from SARs file
#### Storage
    storage_await - Storage await from SARs file
    storage_queue - Storage queue from SARs file
    storage_read_seconds - Storage read/second from SARs file
    storage_sector_average - Storage sector average from SARs file
    storage_service_time - Storage service time from SARs file
    storage_transaction_seconds - Storage transactions/second from SARs file
    storage_utils_percent - Storage utilization percentage from SARs file
    storage_write_seconds - Storage write/second from SARs file
#### Network
    network_rx_packets - Network receive packets per second  from SARs file
    network_tx_packets - Network transmit packets per second from SARs file
    network_rx_kb - Network receive kb per second from SARs file
    network_tx_kb - Network transmit kb per second from SARs file
    network_rx_compressed - Network receive compressed per second from SARs file
    network_tx_compressed - Network transmit compressed per second from SARs file
    network_rx_multicast - Network receive multicast per second from SARs file

#### AMOS
	amos_cpu_used - AMOS CPU usage from INSTR file
	amos_mem_used - AMOS MEM usage from INSTR file
	amos_sessions - AMOS sessions from INSTR file

#### JBoss JVM Memory
    jvm_memory_heap_memory_usage_committed       - JVM Heap Memory Commited usage from INSTR file
    jvm_memory_heap_memory_usage_used            - JVM Heap Memory Used usage from INSTR file

#### JBoss CPU
    cpu_process_cpu_load            - CPU Process CUP Load from INSTR file

#### Top CPU/Memory consuming Process
    top_cpu_process                 - Top CPU consuming process
    top_mem_process                 - Top Memory consuming process

#### Pm file collection
    pmserv_files_collected          - pmserv file collected in current rop
    pmserv_files_failed             - pmserv file failed in current rop
    pmserv_rop_collection_time      - pmserv rop collection time

#### DPMediation
    dp_failed_attempts_with_sas     - numberOfFailedAttempsWithSas from INSTR file
    dp_total_number_of_hb_to_sas    - totalNumberOfHbToSAS Total number of attempts in contacting SAS from INSTR file.

####Domain Proxy
    dp_grant_request_count                                      - grantRequestsCount from INSTR file.
    dp_hb_request_count                                         - heartbeatRequestsCount from INSTR file.
    dp_active_cells_count                                       - numberOfActiveCellsCount from INSTR file.
    dp_failed_connection_attempts_with_sas_incremental          - numberOfFailedConnectionAttemptsWithSasIncremental from INSTR file.
    dp_inactive_cells_count                                     - numberOfInactiveCellsCount from INSTR file.
    dp_maintained_grants_count                                  - numberOfMaintainedGrantsCount from INSTR file.
    dp_registration_requests_count                              - registrationRequestsCount from INSTR file.
    dp_spectrum_inquiry_request_count                           - spectrumInquiryRequestsCount from INSTR file.
    dp_relinquishment_requests_count                            - relinquishmentRequestsCount from INSTR file.

###Build Instructions
```shell script
$ go build
```
####Cross compile for Linux on Windows
```shell script
$ SET GOOS=linux
$ go build
```
