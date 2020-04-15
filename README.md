# check_snmp_status
Nagios/Icinga plugin for retrieving CPU/RAM/Disk/Network usage written in Go. It uses a [gosnmp library created by soniah](https://github.com/soniah/gosnmp)
for handling the network side of SNMP.

## CPU mode
The CPU mode will output performance data in percentage for all cores, and a seperate average percent of all cores.

## RAM mode
RAM mode will output performance data of physical memory in percentage and kilobytes.

## Disk mode
Disk will also output performance data of the specified partition in percentage and kilobytes.

## Interface mode
Interface mode will check if the interface is connected, and output the counter values for sent/received octets

## Usage
<details>
<summary>Command line examples</summary>

Retrieve disk usage of a linux system from a partition mounted on /

```
check_snmp_status.exe -host 10.0.0.1 -community public -mode disk -path / -os linux

DISK / OK - 3% used|'Disk'=1515832KB;;;0;57876188 'Disk %'=3%;100;100;0;100
```

Retrieve disk usage of a windows system from a drive mounted on E:\

```
check_snmp_status.exe -host 10.0.0.2 -community public -mode disk -path E -os windows

DISK E OK - 8% used|'Disk'=39304945664KB;;;0;500106784768 'Disk %'=8%;100;100;0;100
```

Retrieve RAM usage from a linux system

```
check_snmp_status.exe -host 10.0.0.1 -community public -mode ram -os linux

RAM OK - 78% used|'RAM'=9519624KB;;;0;12141468 'RAM %'=78%;100;100;0;100
```

Retrieve CPU usage from a windows system

```
check_snmp_status.exe -host 10.0.0.2 -community public -mode cpu -os windows

CPU OK - 4%|'CPU average'=4%;100;100;0;100 'CPU core 1'=4%;;;0;100 'CPU core 2'=3%;;;0;100 'CPU core 3'=3%;;;0;100 'CPU core 4'=12%;;;0;100 'CPU core 5'=2%;;;0;100 'CPU core 6'=4%;;;0;100 'CPU core 7'=2%;;;0;100 'CPU core 0'=7%;;;0;100
```

Retrieve disk usage and return a warning if the disk usage exceeds 80%, and return a critical if the usage exceeds 95%

```
check_snmp_status.exe -host 10.0.0.3 -community public -mode disk -path E -os windows -C 95 -W 80

DISK E CRITICAL - 99% used|'Disk'=994761728000KB;;;0;1000067821568 'Disk %'=99%;80;95;0;100
```
</details>

All available parameters;

```
Usage of check_snmp_status.exe:
  -C int
        Percentage that should trigger a critical level (default 100)
  -W int
        Percentage that should trigger a warning level (default 100)
  -community string
        SNMP community (default "public")
  -host string
        Host IP address, required parameter
  -mode string
        [disk|cpu|ram|interface] Specify the mode to be used
  -os string
        [linux|windows] Operating system of the target (default "linux")
  -path string
        Partition mount-point with unix or drive letter with windows when retrieving disk usage, is also used for interface name
  -port int
        SNMP port (default 161)
  -version
        Using this parameter will display the version number
```

## Example of configuring Icinga2 to use this plugin
<details>
<summary>commands.conf</summary>

```
object CheckCommand "check-snmp-ram" {
  command = [ PluginDir + "/check_snmp_status" ]

  arguments = {
        "-host"         = "$address$"
        "-port"         = "$snmp_port$"
        "-community"    = "$snmp_community$"
        "-mode"         = "ram"
        "-os"           = "$os$"
        "-C"            = "$ram_critical$"
        "-W"            = "$ram_warning$"
  }
}

object CheckCommand "check-snmp-disk" {
  command = [ PluginDir + "/check_snmp_status" ]

  arguments = {
        "-host"         = "$address$"
        "-port"         = "$snmp_port$"
        "-community"    = "$snmp_community$"
        "-mode"         = "disk"
        "-os"           = "$os$"
        "-C"            = "$disk_critical$"
        "-W"            = "$disk_warning$"
        "-path"         = "$disk_path$"
  }
}

object CheckCommand "check-snmp-cpu" {
  command = [ PluginDir + "/check_snmp_status" ]

  arguments = {
        "-host"         = "$address$"
        "-port"         = "$snmp_port$"
        "-community"    = "$snmp_community$"
        "-mode"         = "cpu"
        "-os"           = "$os$"
        "-C"            = "$cpu_critical$"
        "-W"            = "$cpu_warning$"
  }
}

object CheckCommand "check-snmp-interface" {
  command = [ PluginDir + "/check_snmp_status" ]

  arguments = {
        "-host"         = "$address$"
        "-port"         = "$snmp_port$"
        "-community"    = "$snmp_community$"
        "-mode"         = "interface"
        "-path"         = "$interface_path$"
  }
}

```
</details>
<details>
<summary>services.conf</summary>

```
apply Service "snmp-cpu" {
        import "generic-service"

        check_command = "check-snmp-cpu"

        assign where host.vars.snmp_community != ""
}

apply Service "snmp-ram" {
        import "generic-service"

        check_command = "check-snmp-ram"
        check_interval = 5m

        assign where host.vars.snmp_community != ""
}

apply Service for (disk in host.vars.snmp_disks) {
        import "generic-service"

        check_command = "check-snmp-disk"
        check_interval = 5m
        display_name = "snmp-disk"
        vars.disk_path = disk

        assign where host.vars.snmp_community != "" && host.vars.snmp_disks
}

apply Service for (interface in host.vars.snmp_interfaces) {
        import "generic-service"

        check_command = "check-snmp-interface"
        check_interval = 5m
        display_name = "snmp-interface"
        vars.interface_path = interface

        assign where host.vars.snmp_community != "" && host.vars.snmp_interfaces
}

```
</details>
<details>
<summary>hosts.conf</summary>

```
object Host "linux-host1" {
        import "generic-host"
        address = "10.0.0.1"

        vars.os = "linux"
        vars.snmp_community = "public"

        vars.snmp_disks = ["/", "/var"]
        vars.snmp_interfaces = ["eth0"]
}
```
</details>
