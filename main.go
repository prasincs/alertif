package main

import (
     "fmt"
     "github.com/codegangsta/cli"
     "github.com/shirou/gopsutil/disk"
     "github.com/stvp/pager"
     "os"
     "strings"
)

type DiskUsageReport struct {
     Mountpoint string `json:"mountpoint"`
     Usage      *disk.DiskUsageStat
}

func contains(haystack []string, needle string) bool {
     for _, item := range haystack {
          if item == needle {
               return true
          }
     }
     return false
}

func main() {
     app := cli.NewApp()
     app.Name = "AlertIf"
     app.Usage = "Sends alerts based on conditions"

     app.Flags = []cli.Flag{
          cli.StringFlag{
               Name:  "pagerduty-servicekey, s",
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
     }

     app.Action = func(c *cli.Context) {
          hostName, err := os.Hostname()
          if err != nil {
               fmt.Println("Hostname couldn't be obtained", err)
          }
          serviceKey := c.String("pagerduty-servicekey")
          if serviceKey == "" {
               fmt.Println("PagerDuty serviceKey not given, writing to stdout instead.")
          }
          pagerDutyService := pager.New(serviceKey)

          checkDisk := c.Bool("disk")
          if checkDisk == false {
               fmt.Println("Disk checking is disabled, there's nothing for me to do.")
               os.Exit(65)
          }
          diskThreshold := c.Int("disk-threshold")
          if diskThreshold < 0 || diskThreshold > 100 {
               fmt.Println("Disk Threshold needs to be an integer between 0 and 100")
               os.Exit(65)
          }
          ignoredMountPoints := strings.Split(c.String("disk-ignore"), ",")

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

          fmt.Printf("length: %d, %v", len(reports), reports)

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
     }

     app.Run(os.Args)

}