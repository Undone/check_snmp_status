# check_snmp_status
Nagios/Icinga plugin for retrieving temperature or CPU/RAM/disk/network usage written in Go. It uses a [gosnmp library created by soniah](https://github.com/soniah/gosnmp)
for handling the network side of SNMP.

## CPU mode
The CPU mode will output performance data in percentage for all cores, and a seperate average percent of all cores.

## RAM mode
RAM mode will output performance data of physical memory in percentage and kilobytes.

## Disk mode
Disk will also output performance data of the specified partition in percentage and kilobytes.

## Interface mode
Interface mode will check if the interface is connected, and output the counter values for sent/received octets

## Temperature mode
Temperature mode will get the temperature for a specified sensor, requires lm-sensors on the host. Sensor name has to be manually looked up with snmpwalk, sensors don't have a proper naming system so they will be different on different systems.
<details>
<summary>Example</summary>
If you compare the snmpwalk output to your sensors output you can easily find the sensor name you want to monitor.

```
[root@srv ~]# snmpwalk -v 2c -c public srv01 .1.3.6.1.4.1.2021.13.16.2
LM-SENSORS-MIB::lmTempSensorsIndex.9 = INTEGER: 9
LM-SENSORS-MIB::lmTempSensorsIndex.10 = INTEGER: 10
LM-SENSORS-MIB::lmTempSensorsIndex.11 = INTEGER: 11
LM-SENSORS-MIB::lmTempSensorsIndex.12 = INTEGER: 12
LM-SENSORS-MIB::lmTempSensorsIndex.13 = INTEGER: 13
LM-SENSORS-MIB::lmTempSensorsIndex.14 = INTEGER: 14
LM-SENSORS-MIB::lmTempSensorsIndex.15 = INTEGER: 15
LM-SENSORS-MIB::lmTempSensorsIndex.16 = INTEGER: 16
LM-SENSORS-MIB::lmTempSensorsIndex.17 = INTEGER: 17
LM-SENSORS-MIB::lmTempSensorsIndex.18 = INTEGER: 18
LM-SENSORS-MIB::lmTempSensorsIndex.19 = INTEGER: 19
LM-SENSORS-MIB::lmTempSensorsIndex.20 = INTEGER: 20
LM-SENSORS-MIB::lmTempSensorsIndex.21 = INTEGER: 21
LM-SENSORS-MIB::lmTempSensorsIndex.22 = INTEGER: 22
LM-SENSORS-MIB::lmTempSensorsDevice.9 = STRING: temp1
LM-SENSORS-MIB::lmTempSensorsDevice.10 = STRING: temp2
LM-SENSORS-MIB::lmTempSensorsDevice.11 = STRING: temp3
LM-SENSORS-MIB::lmTempSensorsDevice.12 = STRING: temp4
LM-SENSORS-MIB::lmTempSensorsDevice.13 = STRING: temp5
LM-SENSORS-MIB::lmTempSensorsDevice.14 = STRING: temp6
LM-SENSORS-MIB::lmTempSensorsDevice.15 = STRING: temp8
LM-SENSORS-MIB::lmTempSensorsDevice.16 = STRING: acpitz-virtual-0:temp1
LM-SENSORS-MIB::lmTempSensorsDevice.17 = STRING: acpitz-virtual-0:temp2
LM-SENSORS-MIB::lmTempSensorsDevice.18 = STRING: Package id 0
LM-SENSORS-MIB::lmTempSensorsDevice.19 = STRING: Core 0
LM-SENSORS-MIB::lmTempSensorsDevice.20 = STRING: Core 1
LM-SENSORS-MIB::lmTempSensorsDevice.21 = STRING: Core 2
LM-SENSORS-MIB::lmTempSensorsDevice.22 = STRING: Core 3
LM-SENSORS-MIB::lmTempSensorsValue.9 = Gauge32: 34000
LM-SENSORS-MIB::lmTempSensorsValue.10 = Gauge32: 33000
LM-SENSORS-MIB::lmTempSensorsValue.11 = Gauge32: 31000
LM-SENSORS-MIB::lmTempSensorsValue.12 = Gauge32: 55000
LM-SENSORS-MIB::lmTempSensorsValue.13 = Gauge32: 32000
LM-SENSORS-MIB::lmTempSensorsValue.14 = Gauge32: 30000
LM-SENSORS-MIB::lmTempSensorsValue.15 = Gauge32: 34000
LM-SENSORS-MIB::lmTempSensorsValue.16 = Gauge32: 27800
LM-SENSORS-MIB::lmTempSensorsValue.17 = Gauge32: 29800
LM-SENSORS-MIB::lmTempSensorsValue.18 = Gauge32: 36000
LM-SENSORS-MIB::lmTempSensorsValue.19 = Gauge32: 35000
LM-SENSORS-MIB::lmTempSensorsValue.20 = Gauge32: 35000
LM-SENSORS-MIB::lmTempSensorsValue.21 = Gauge32: 36000
LM-SENSORS-MIB::lmTempSensorsValue.22 = Gauge32: 36000
```
Notice how the temp1 is listed for two modules so it will have the module name prefixed in the snmp table when it occurs the 2nd time, this makes it difficult to know the names in advance when dealing with different hardware.
```
[root@srv ~]# sensors
theseus-isa-0640
Adapter: ISA adapter
3.3V:         +3.31 V  
VREF:         +1.06 V  
VBAT:         +3.11 V  
3.3AUX:       +3.31 V  
12V:         +12.05 V  
fan1:           0 RPM
fan2:           0 RPM
fan4:           FAULT
temp1:        +34.0°C  
temp2:        +33.0°C  
temp3:        +31.0°C  
temp4:        +55.0°C  
temp5:        +32.0°C  
temp6:        +30.0°C  
temp8:        +34.0°C  

acpitz-virtual-0
Adapter: Virtual device
temp1:        +27.8°C  (crit = +106.0°C)
temp2:        +29.8°C  (crit = +106.0°C)

coretemp-isa-0000
Adapter: ISA adapter
Package id 0:  +38.0°C  (high = +85.0°C, crit = +105.0°C)
Core 0:        +35.0°C  (high = +85.0°C, crit = +105.0°C)
Core 1:        +33.0°C  (high = +85.0°C, crit = +105.0°C)
Core 2:        +38.0°C  (high = +85.0°C, crit = +105.0°C)
Core 3:        +36.0°C  (high = +85.0°C, crit = +105.0°C)

```
</details>

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

Retrieve temperature and return a warning if the temperature exceeds 50, and return a critical if the temperature exceeds 70

```
./check_snmp_status -host 10.0.0.3 -community public -mode temp -path temp1 -W 50 -C 70

TEMP OK - 34|'Temperature'=34;50;70
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
        Partition mount-point with unix or drive letter with windows when retrieving disk usage, is also used for interface name and sensor name
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

object CheckCommand "check-snmp-temperature" {
  command = [ PluginDir + "/check_snmp_status" ]

  arguments = {
        "-host"         = "$address$"
        "-port"         = "$snmp_port$"
        "-community"    = "$snmp_community$"
        "-mode"         = "temp"
        "-path"         = "$sensor_name$"
        "-os"           = "$os$"
        "-C"            = "$temp_critical$"
        "-W"            = "$temp_warning$"
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

apply Service for (sensor in host.vars.snmp_sensors) {
        import "generic-service"

        check_command = "check-snmp-temperature"
        check_interval = 1m
        display_name = "snmp-temperature"
        vars.sensor_name = sensor

        assign where host.vars.snmp_community != "" && host.vars.snmp_sensors
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
        vars.snmp_sensors = ["Package id 0"]
}
```
</details>
