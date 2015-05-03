package main

import (
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/shirou/gopsutil/disk"
	"github.com/stvp/pager"
	"net"
	"net/http"
	"os"
	"strings"
)

type DiskUsageReport struct {
	Mountpoint string              `json:"mountpoint"`
	Usage      *disk.DiskUsageStat `json: "usage"`
}

type ServiceCommand struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Port   string `json:"port"`
	Action string `json:"action"`
}

func contains(haystack []string, needle string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

func parseServiceCmd(cmd string) (ServiceCommand, error) {
	items := strings.Split(cmd, ",")
	if len(items) != 4 {
		return ServiceCommand{}, errors.New("Service Command format was invalid. Not enough arguments")
	}
	return ServiceCommand{
		Name:   items[0],
		Type:   items[1],
		Port:   items[2],
		Action: items[3],
	}, nil
}

func tcpCheckHandler(serviceCmd ServiceCommand, hostName string) (string, error) {
	switch serviceCmd.Action {
	case "dead":
		_, err := net.Dial("tcp", fmt.Sprintf("%s:%s",hostName, serviceCmd.Port))
		if err != nil {
			return "Dead", err
		}
		return "Not Dead!", nil
	default:
		return "", errors.New(fmt.Sprintf("Unknown service action: %s", serviceCmd.Action))
	}
}

func httpCheckHandler(serviceCmd ServiceCommand, hostName string) (string, error) {
	_, err := http.Get(fmt.Sprintf("http://%s:%s%s", hostName, serviceCmd.Port, serviceCmd.Action))
	//defer resp.Body.Close()
	if err != nil {
		return "http Check failed", err;
	}

	return "http Check Succeeded", nil
}

func executeServiceCmd(serviceCmd ServiceCommand, hostName string) (string, error) {
	switch serviceCmd.Type {
	case "tcp":
		return tcpCheckHandler(serviceCmd,hostName)
	case "http":
		return httpCheckHandler(serviceCmd,hostName)
	default:
		return "", errors.New(fmt.Sprintf("Unknown service type supplied: %s", serviceCmd.Type))
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "AlertIf"
	app.Usage = "Sends alerts based on conditions"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "pagerduty-servicekey, p",
			Value: "<something>",
			Usage: "Use your PagerDuty Service Key",
		},
		cli.BoolFlag{
			Name:  "disk",
			Usage: "Alert if Disk usage-percent > percent",
		},
		cli.IntFlag{
			Name:  "disk-threshold, t",
			Value: 90,
			Usage: "Set an integer threshold for alerting.",
		},
		cli.StringFlag{
			Name:  "disk-ignore, i",
			Value: "/dev",
			Usage: "Mount points you want to ignore.",
		},
		cli.StringFlag{
			Name:  "hostname, H",
			Value: "",
			Usage: "Hostname to check, defaults to localhost unless overridden",
		},
		cli.StringFlag{
			Name:  "service, s",
			Usage: "Alert if Some service is misbehaving. name,type,port,action",
		},
	}

	app.Action = func(c *cli.Context) {
		hostName := "localhost"
		overrideHostName := c.String("hostname")
		if overrideHostName == "" {
			hostName2, err2 := os.Hostname()
			if err2 != nil {
				fmt.Println("Hostname couldn't be obtained", err2)
			}else {
				hostName = hostName2
			}
		} else {
			hostName = overrideHostName
		}
		serviceKey := c.String("pagerduty-servicekey")
		if serviceKey == "" {
			fmt.Println("PagerDuty serviceKey not given, writing to stdout instead.")
		}
		pagerDutyService := pager.New(serviceKey)

		checkDisk := c.Bool("disk")
		if checkDisk == false {
			fmt.Println("Disk checking is disabled, there's nothing for me to do.")
		}
		diskThreshold := c.Int("disk-threshold")
		if diskThreshold < 0 || diskThreshold > 100 {
			fmt.Println("Disk Threshold needs to be an integer between 0 and 100")
			os.Exit(65)
		}
		ignoredMountPoints := strings.Split(c.String("disk-ignore"), ",")
		serviceCmdStr := c.String("service")
		serviceCommand, err := parseServiceCmd(serviceCmdStr)
		if err != nil {
			fmt.Printf("Failed to parse service command: %s, err: %v\n", serviceCmdStr, err)
			os.Exit(65)
		}
		fmt.Printf("ServiceCommand: %v\n", serviceCommand)
		fmt.Println("Hostname:", hostName)
		fmt.Println("PagerDuty ServiceKey: ", serviceKey)
		fmt.Println("Check Disk? : ", checkDisk)
		fmt.Println("Disk Threshold: ", diskThreshold)
		fmt.Println("Ignored Disks: ", ignoredMountPoints)

		partitions, _ := disk.DiskPartitions(true)
		reports := make([]DiskUsageReport, 0, len(partitions))
		for _, partition := range partitions {
			mountPoint := partition.Mountpoint
			if !contains(ignoredMountPoints, mountPoint) {
				usage, _ := disk.DiskUsage(mountPoint)
				if usage.UsedPercent > float64(diskThreshold) {
					reports = append(reports, DiskUsageReport{
						Mountpoint: mountPoint,
						Usage:      usage,
					})
				}
			}

		}

		fmt.Printf("length: %d, %v\n", len(reports), reports)

		if len(reports) > 0 {
			title := fmt.Sprintf("[%s] Disk usage threshold reached for %d disk(s)", hostName, len(reports))
			pdMap := make(map[string]interface{})
			for _, report := range reports {
				pdMap[report.Mountpoint] = report.Usage.UsedPercent
			}
			incidentKey, err := pagerDutyService.TriggerWithDetails(title, pdMap)
			if err != nil {
				fmt.Println("Ran into an error pushing to PagerDuty.")
			} else {
				fmt.Println("Reported to PagerDuty with incidentKey: ", incidentKey)

			}
		}

		_, err = executeServiceCmd(serviceCommand,hostName)
		if err != nil {
			title := fmt.Sprintf("[%s][%s] %s service on port %s is down: %v", hostName,
				serviceCommand.Name, serviceCommand.Type, serviceCommand.Port, err)
			incidentKey, err := pagerDutyService.Trigger(title)
			if err != nil {
				fmt.Println("Ran into an error pushing to PagerDuty.")
			} else {
				fmt.Println("Reported to PagerDuty with incidentKey: ", incidentKey)

			}
		}
	}

	app.Run(os.Args)

}
