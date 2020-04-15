package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/soniah/gosnmp"
)

// Version of release
const Version = "1.2"

// UCD-SNMP-MIB, for agents using NET-SNMP
const (
	// RAM
	snmpMemTotal     = ".1.3.6.1.4.1.2021.4.5.0"
	snmpMemAvailable = ".1.3.6.1.4.1.2021.4.6.0"
	snmpMemCached    = ".1.3.6.1.4.1.2021.4.15.0"
	snmpMemBuffered  = ".1.3.6.1.4.1.2021.4.14.0"
)

// HOST-RESOURCES-MIB
const (
	// Disk, RAM
	snmpHrStorageDescr      = ".1.3.6.1.2.1.25.2.3.1.3"
	snmpHrStorageAllocUnits = ".1.3.6.1.2.1.25.2.3.1.4"
	snmpHrStorageSize       = ".1.3.6.1.2.1.25.2.3.1.5"
	snmpHrStorageUsed       = ".1.3.6.1.2.1.25.2.3.1.6"

	// CPU
	snmpHrProcessorLoad = ".1.3.6.1.2.1.25.3.3.1.2"
)

// ifMib, 64-bit interface counters
const (
	snmpIfName				= ".1.3.6.1.2.1.31.1.1.1.1"
	snmpIfHCInOctets		= ".1.3.6.1.2.1.31.1.1.1.6"
	snmpIfHCOutOctets		= ".1.3.6.1.2.1.31.1.1.1.10"
	snmpIfConnectorPresent	= ".1.3.6.1.2.1.31.1.1.1.17"
)

const (
	osLinux   = "linux"
	osWindows = "windows"
)

// Nagios plugin format exit codes
const (
	nagiosOk       = 0
	nagiosWarning  = 1
	nagiosCritical = 2
	nagiosUnknown  = 3
)

// Possible values for ifMib::ifConnectorPresent
const (
	ConnectorPresentTrue	= 1
	ConnectorPresentFalse	= 2
)

// snmpDisk contains the disk usage information of a partition
type snmpDisk struct {
	Index string // Index in the dskTable
	Path string // Path of disk mount point
	Total int64 // Total disk capacity
	Used int64 // Used disk capacity
	Percent int // Used disk capacity in percentage
}

// snmpRAM contains the values for RAM
type snmpRAM struct {
	Total int64 // Total amount of RAM in system
	Available int64 // Available RAM capacity
	Buffered int64 // Amount of buffered RAM
	Cached int64 // Amount of cached RAM
	Used int64 // Used RAM capacity
	Percent int // Used RAM capacity in percentage
}

type snmpCPU map[int]int

// Calculate the average CPU percentage
func (cpu snmpCPU) Average() int {
	var avg int

	for _, load := range cpu {
		avg += load
	}

	avg = avg / len(cpu)

	return avg
}

// snmpInterface contains interface statistics
type snmpInterface struct {
	Name string // Name of the interface
	Index string // Index in the snmp ifXEntry table
	InOctets int64 // Received octets
	OutOctets int64 // Sent octets
	ConnectorPresent int // Is the interface connected
}

func main() {
	var host = flag.String("host", "", "Host IP address, required parameter")
	var port = flag.Int("port", 161, "SNMP port")
	var community = flag.String("community", "public", "SNMP community")
	var path = flag.String("path", "", "Partition mount-point with unix or drive letter with windows when retrieving disk usage, is also used for interface name")
	var mode = flag.String("mode", "", "[disk|cpu|ram|interface] Specify the mode to be used")
	var operatingSystem = flag.String("os", osLinux, "[linux|windows] Operating system of the target")
	var warning = flag.Int("W", 100, "Percentage that should trigger a warning level")
	var critical = flag.Int("C", 100, "Percentage that should trigger a critical level")
	var version = flag.Bool("version", false, "Using this parameter will display the version number")

	flag.Parse()

	// Convert OS to lowercase so the argument can use whatever casing
	*operatingSystem = strings.ToLower(*operatingSystem)

	if *version == true {
		fmt.Println("check_snmp_status version " + Version)
		os.Exit(nagiosUnknown)
	} else if *host == "" || (*mode == "disk" && *path == "") || (*operatingSystem != osLinux && *operatingSystem != osWindows) { // Check for invalid arguments
		flag.Usage()
		os.Exit(nagiosUnknown)
	}

	// Initialise the SNMP object
	snmp := &gosnmp.GoSNMP{
		Target:    *host,
		Community: *community,
		Port:      uint16(*port),
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(5) * time.Second,
	}

	// Set the default status to unknown, if something goes wrong it's better than ok
	var returnCode = nagiosUnknown

	switch *mode {
	case "cpu":
		cpu, err := getCPU(snmp)

		if err != nil {
			fmt.Println("getCPU error:", err)
			os.Exit(nagiosUnknown)
		}

		var avg = cpu.Average()

		returnCode = getStatus(*warning, *critical, avg)

		// Print display and performance data for the average of all cores
		fmt.Printf("CPU %s - %d%%", convertStatus(returnCode), avg)
		fmt.Printf("|'CPU average'=%d%%;%d;%d;0;100", avg, *warning, *critical)

		// Print performance data for each core
		for i, load := range cpu {
			fmt.Printf(" 'CPU core %d'=%d%%;;;0;100", i, load)
		}
	case "disk":
		disk, err := getDisk(snmp, *path, *operatingSystem)

		if err != nil {
			fmt.Println("getDisk error:", err)
			os.Exit(nagiosUnknown)
		}
		returnCode = getStatus(*warning, *critical, disk.Percent)

		// Print display data, performance data in bytes, and performance data in percentage, including warning levels
		fmt.Printf("DISK %s %s - %d%% used", *path, convertStatus(returnCode), disk.Percent)
		fmt.Printf("|'Disk'=%dB;;;0;%d", disk.Used, disk.Total)
		fmt.Printf(" 'Disk %%'=%d%%;%d;%d;0;100", disk.Percent, *warning, *critical)
	case "ram":
		switch *operatingSystem {
		case osLinux:
			ram, err := getRAM(snmp)

			if err != nil {
				fmt.Println("getRAM error:", err)
				os.Exit(nagiosUnknown)
			}

			returnCode = getStatus(*warning, *critical, ram.Percent)

			fmt.Printf("RAM %s - %d%% used", convertStatus(returnCode), ram.Percent)
			fmt.Printf("|'RAM'=%dKB;;;0;%d", ram.Used, ram.Total)
			fmt.Printf(" 'RAM %%'=%d%%;%d;%d;0;100", ram.Percent, *warning, *critical)
		case osWindows:
			ram, err := getDisk(snmp, "Physical Memory", *operatingSystem)

			if err != nil {
				fmt.Println("getRAM error:", err)
				os.Exit(nagiosUnknown)
			}
			returnCode = getStatus(*warning, *critical, ram.Percent)

			fmt.Printf("RAM %s - %d%% used", convertStatus(returnCode), ram.Percent)
			fmt.Printf("|'RAM'=%dB;;;0;%d", ram.Used, ram.Total)
			fmt.Printf(" 'RAM %%'=%d%%;%d;%d;0;100", ram.Percent, *warning, *critical)
		}
	case "interface":
		iface, err := getInterface(snmp, *path)

		if err != nil {
			fmt.Println("getInterface error:", err)
			os.Exit(nagiosUnknown)
		}

		// Check if the interface is connected
		if iface.ConnectorPresent == ConnectorPresentTrue {
			returnCode = nagiosOk
			fmt.Printf("INTERFACE %s - Connected", *path)
		} else {
			returnCode = nagiosWarning
			fmt.Printf("INTERFACE %s - Disconnected", *path)
		}

		// Print performance data
		fmt.Printf("|'Interface In'=%dc", iface.InOctets)
		fmt.Printf(" 'Interface Out'=%dc", iface.OutOctets)
	case "":
		fmt.Println("No mode selected")
	}

	os.Exit(returnCode)
}

func getCPU(snmp *gosnmp.GoSNMP) (snmpCPU, error) {
	var cpu = snmpCPU{}
	err := snmp.Connect()

	if err != nil {
		return cpu, err
	}
	defer snmp.Conn.Close()

	regexCPU := regexp.MustCompile(regexp.QuoteMeta(snmpHrProcessorLoad) + `\.\d+`)

	var i int

	err = snmp.BulkWalk(snmpHrProcessorLoad, func(pdu gosnmp.SnmpPDU) error {

		if regexCPU.MatchString(pdu.Name) {
			// hrProcessorLoad index numbers are completely arbitrary so we just generate a new one, this means that the generated index might not represent the actual core number,
			// but we're not getting the actual number either way
			cpu[i] = pdu.Value.(int)
			i++
		}

		return nil
	})

	if err != nil {
		return cpu, err
	} else if i == 0 {
		return cpu, fmt.Errorf("No CPU cores found")
	}

	return cpu, err
}

// getDisk retrieves disk information using the HOST-RESOURCES-MIB::hrStorage MIB
func getDisk(snmp *gosnmp.GoSNMP, path string, operatingSystem string) (snmpDisk, error) {
	var disk = snmpDisk{}

	err := snmp.Connect()

	if err != nil {
		return disk, err
	}
	defer snmp.Conn.Close()

	regexPaths := regexp.MustCompile(regexp.QuoteMeta(snmpHrStorageDescr) + `\.(\d+)`)

	if operatingSystem == osWindows {
		regexLetter := regexp.MustCompile(path + `:\\.*Label:.*Serial Number.*`)

		// Retrieve the index number of the drive
		err = snmp.BulkWalk(snmpHrStorageDescr, func(pdu gosnmp.SnmpPDU) error {
			if regexPaths.MatchString(pdu.Name) {
				var group = regexPaths.FindStringSubmatch(pdu.Name)
				var diskPath = string(pdu.Value.([]byte))

				// If we're trying to get RAM, don't use the regex as it won't match
				if (path == "Physical Memory" && diskPath == "Physical Memory") || regexLetter.MatchString(diskPath) {
					disk = snmpDisk{
						Index: group[1],
						Path:  diskPath,
					}
				}

			}

			return nil
		})
	} else if operatingSystem == osLinux {
		index, err2 := findIndexByName(snmp, snmpHrStorageDescr, path)

		if err2 != nil {
			return disk, err2
		} else if index == "" {
			return disk, fmt.Errorf("Disk %s not found", path)
		}

		disk = snmpDisk{
			Index: index,
			Path: path,
		}
	} else {
		return disk, fmt.Errorf("Unsupported operating system")
	}

	if err != nil {
		return disk, err
	} else if disk.Index == "" { // Check if the index field has been populated, if it's not, the disk doesn't exist
		return disk, fmt.Errorf("Disk %s not found", path)
	}

	// Append the index number to the OIDs, this is how we get information only regarding that partition
	var oidTotal = snmpHrStorageSize + "." + disk.Index
	var oidUsed = snmpHrStorageUsed + "." + disk.Index
	var oidUnits = snmpHrStorageAllocUnits + "." + disk.Index

	var oids = []string{oidTotal, oidUsed, oidUnits}

	result, err2 := snmp.Get(oids)

	if err2 != nil {
		return disk, err2
	}

	var units int64

	// Loop the results and populate the snmpDisk object
	for _, pdu := range result.Variables {
		switch pdu.Name {
		case oidTotal:
			disk.Total = gosnmp.ToBigInt(pdu.Value).Int64()
		case oidUsed:
			disk.Used = gosnmp.ToBigInt(pdu.Value).Int64()
		case oidUnits:
			units = gosnmp.ToBigInt(pdu.Value).Int64()
		}
	}

	// hrStorage used/total values need to be multiplied by units to get the actual value
	disk.Total *= units
	disk.Used *= units

	// Do a calculation of the percentage
	disk.Percent = calculatePercentage(disk.Used, disk.Total)

	return disk, nil
}

// getRAM retrieves RAM usage information from UCD-SNMP-MIB MIB
func getRAM(snmp *gosnmp.GoSNMP) (snmpRAM, error) {
	var ram = snmpRAM{}
	var oids = []string{snmpMemAvailable, snmpMemTotal, snmpMemBuffered, snmpMemCached}

	err := snmp.Connect()

	if err != nil {
		return ram, err
	}
	defer snmp.Conn.Close()

	result, err2 := snmp.Get(oids)

	if err2 != nil {
		return ram, err2
	}

	for _, pdu := range result.Variables {
		switch pdu.Name {
		case snmpMemTotal:
			ram.Total = gosnmp.ToBigInt(pdu.Value).Int64()
		case snmpMemAvailable:
			ram.Available = gosnmp.ToBigInt(pdu.Value).Int64()
		case snmpMemBuffered:
			ram.Buffered = gosnmp.ToBigInt(pdu.Value).Int64()
		case snmpMemCached:
			ram.Cached = gosnmp.ToBigInt(pdu.Value).Int64()
		}
	}

	// Reduce buffered and cached RAM from the result, they're actually not unavailable since they can be allocated to any process that needs them
	ram.Used = (ram.Total - ram.Available) - ram.Buffered - ram.Cached
	ram.Percent = calculatePercentage(ram.Used, ram.Total)

	return ram, nil
}

// getInterface retrieves interface statistics from ifMib
func getInterface(snmp *gosnmp.GoSNMP, interfaceName string) (snmpInterface, error) {
	var iface = snmpInterface{}
	err := snmp.Connect()

	if err != nil {
		return iface, err
	}
	defer snmp.Conn.Close()

	index, err2 := findIndexByName(snmp, snmpIfName, interfaceName)

	if err2 != nil {
		return iface, err
	} else if index == "" { // Check if the index field has been populated, if it's not, the interface doesn't exist
		return iface, fmt.Errorf("Interface %s not found", interfaceName)
	}

	iface = snmpInterface{
		Index: index,
		Name: interfaceName,
	}

	// Append the index number to the OIDs
	var oidInOctets = snmpIfHCInOctets + "." + iface.Index
	var oidOutOctets = snmpIfHCOutOctets + "." + iface.Index
	var oidConnPresent = snmpIfConnectorPresent + "." + iface.Index

	var oids = []string{oidInOctets, oidOutOctets, oidConnPresent}

	result, err2 := snmp.Get(oids)

	if err2 != nil {
		return iface, err2
	}

	for _,pdu := range result.Variables {
		switch pdu.Name {
		case oidInOctets:
			iface.InOctets = gosnmp.ToBigInt(pdu.Value).Int64()
		case oidOutOctets:
			iface.OutOctets = gosnmp.ToBigInt(pdu.Value).Int64()
		case oidConnPresent:
			iface.ConnectorPresent = pdu.Value.(int)
		}
	}

	return iface, nil
}

// find the oid index number for a value
func findIndexByName(snmp *gosnmp.GoSNMP, tableName string, indexName string) (string, error) {
	var index = ""
	regexNames := regexp.MustCompile(regexp.QuoteMeta(tableName) + `\.(\d+)`)
	
	err := snmp.BulkWalk(tableName, func(pdu gosnmp.SnmpPDU) error {
		if regexNames.MatchString(pdu.Name) {
			var group = regexNames.FindStringSubmatch(pdu.Name)
			var name = string(pdu.Value.([]byte))

			if name == indexName {
				index = group[1]
				return nil
			}
		}

		return nil
	})

	return index, err
}

// getStatus checks if the value parameter is greater than critical level or warning level, and returns a nagios format exit code
func getStatus(warning int, critical int, value int) int {
	if critical < 100 && critical < value {
		return nagiosCritical
	} else if warning < 100 && warning < value {
		return nagiosWarning
	} else {
		return nagiosOk
	}
}

// Convert exit codes into displayable names
func convertStatus(status int) string {
	if status == nagiosCritical {
		return "CRITICAL"
	} else if status == nagiosWarning {
		return "WARNING"
	} else if status == nagiosOk {
		return "OK"
	} else {
		return "UNKNOWN"
	}
}

// Calculate the percentage between two values
func calculatePercentage(used int64, total int64) int {
	var percent = float64(used) / float64(total) * 100

	return int(math.RoundToEven(percent))
}

