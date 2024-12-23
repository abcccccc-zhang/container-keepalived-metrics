// package main

// import (
// 	"context"
// 	"flag"
// 	"fmt"
// 	"log"
// 	"net"
// 	"os"
// 	"strings"
// 	"time"

// 	"net/http"

// 	"github.com/docker/docker/api/types/container"
// 	"github.com/docker/docker/api/types/filters"
// 	"github.com/docker/docker/client"
// 	"github.com/prometheus/client_golang/prometheus"
// 	"github.com/prometheus/client_golang/prometheus/promauto"
// 	"github.com/prometheus/client_golang/prometheus/promhttp"
// )

// // Prometheus 指标定义
// var (
// 	restartCount = promauto.NewGaugeVec(
// 		prometheus.GaugeOpts{
// 			Name: "container_restart_count",
// 			Help: "The number of restarts of a container",
// 		},
// 		[]string{"container_id"},
// 	)

// 	containerStatus = promauto.NewGaugeVec(
// 		prometheus.GaugeOpts{
// 			Name: "container_status",
// 			Help: "The status of the container: 1=running, 0=stopped",
// 		},
// 		[]string{"container_id"},
// 	)

// 	vipFound = promauto.NewGaugeVec(
// 		prometheus.GaugeOpts{
// 			Name: "vip_found",
// 			Help: "Whether the target VIP is found: 1=found, 0=not found",
// 		},
// 		[]string{"vip_res"},
// 	)

// 	vipChangeCount = promauto.NewCounter(
// 		prometheus.CounterOpts{
// 			Name: "vip_change_count",
// 			Help: "Counter that increments every time the VIP address changes",
// 		},
// 	)
// )

// func getContainerMetrics(ctx context.Context, cli *client.Client, containerName string, VIP string, lastVipStatus *string) {
// 	// 查找容器
// 	args := filters.NewArgs()
// 	args.Add("name", "^/"+containerName)

// 	containers, err := cli.ContainerList(ctx, container.ListOptions{Filters: args})
// 	if err != nil {
// 		log.Println("Error listing containers:", err)
// 		return
// 	}

// 	if len(containers) == 0 {
// 		log.Println("No containers found")
// 		return
// 	}

// 	containerID := containers[0].ID
// 	// log.Printf("Found container ID: %s\n", containerID)

// 	// 获取容器信息
// 	containerInspect, err := cli.ContainerInspect(ctx, containerID)
// 	if err != nil {
// 		log.Println("Error inspecting container:", err)
// 		return
// 	}

// 	// 更新重启次数指标
// 	restartCount.WithLabelValues(containerID).Set(float64(containerInspect.RestartCount))

// 	// 更新容器状态指标
// 	if containerInspect.State.Running {
// 		containerStatus.WithLabelValues(containerID).Set(1)
// 	} else {
// 		containerStatus.WithLabelValues(containerID).Set(0)
// 	}

// 	// 查找 VIP
// 	vipFoundValue := findSpecificIP(VIP) // 检查VIP是否存在
// 	var currentVipStatus string
// 	if vipFoundValue == 0 {
// 		vipFound.WithLabelValues(VIP).Set(1)
// 		currentVipStatus = "found"
// 	} else {
// 		vipFound.WithLabelValues(VIP).Set(0)
// 		currentVipStatus = "not_found"
// 	}

// 	// 第一次运行时记录状态
// 	if *lastVipStatus == "" {
// 		*lastVipStatus = currentVipStatus
// 		writeVipStatusToFile(currentVipStatus, VIP)
// 		return
// 	}

// 	// 检查VIP状态是否发生变化
// 	if currentVipStatus != *lastVipStatus {
// 		log.Printf("VIP %s status changed from %s to %s", VIP, *lastVipStatus, currentVipStatus)
// 		vipChangeCount.Inc()                        // 增加VIP变化计数
// 		writeVipStatusToFile(currentVipStatus, VIP) // 更新文件中的状态
// 		*lastVipStatus = currentVipStatus           // 更新lastVipStatus
// 	}
// }

// // 查找特定的 VIP IP 地址
// func findSpecificIP(targetIP string) int {
// 	// 获取宿主机的所有网络接口
// 	interfaces, err := net.Interfaces()
// 	if err != nil {
// 		log.Fatalf("Error getting interfaces: %v", err)
// 	}

// 	// 遍历网络接口，查找目标 IP 地址
// 	for _, i := range interfaces {
// 		addrs, err := i.Addrs()
// 		if err != nil {
// 			log.Printf("Error getting addresses for interface %v: %v", i.Name, err)
// 			continue
// 		}

// 		for _, addr := range addrs {
// 			// 检查是否为 IPv4 地址
// 			ip, _, err := net.ParseCIDR(addr.String())
// 			if err != nil {
// 				continue
// 			}

// 			// 排除回环地址
// 			if ip.IsLoopback() {
// 				continue
// 			}

// 			// 如果找到了目标 IP 地址，返回该地址
// 			if ip.String() == targetIP {
// 				return 0
// 			}
// 		}
// 	}

// 	// 如果没有找到目标 IP 地址，返回空字符串
// 	return 1
// }

// // 记录 VIP 状态到文件
// func writeVipStatusToFile(status, VIP string) {
// 	fmt.Println("----------", VIP)
// 	file, err := os.OpenFile(fmt.Sprintf("/tmp/vip_status_%s.txt", VIP), os.O_RDWR|os.O_CREATE, 0644)
// 	if err != nil {
// 		log.Fatalf("Error opening VIP status file: %v", err)
// 	}
// 	defer file.Close()

// 	// 先清空文件内容，再写入新的状态
// 	_, err = file.Seek(0, 0)
// 	if err != nil {
// 		log.Fatalf("Error seeking in VIP status file: %v", err)
// 	}
// 	_, err = file.WriteString(status + "\n")
// 	if err != nil {
// 		log.Fatalf("Error writing to VIP status file: %v", err)
// 	}

// 	log.Printf("VIP %s status updated to: %s", VIP, status)
// }

// func readLastVipStatus(VIP string) (string, error) {
// 	file, err := os.OpenFile(fmt.Sprintf("/tmp/vip_status_%s.txt", VIP), os.O_RDONLY, 0644)
// 	if err != nil {
// 		return "", fmt.Errorf("no previous VIP status file for %s, starting fresh", VIP)
// 	}
// 	defer file.Close()

// 	var status string
// 	_, err = fmt.Fscanf(file, "%s\n", &status)
// 	if err != nil {
// 		return "", fmt.Errorf("error reading VIP status: %v", err)
// 	}

// 	return status, nil
// }

// func main() {
// 	// 初始化 Docker 客户端
// 	containerNames := flag.String("container_name", "keepalive-master", "The names of the containers to monitor, separated by commas")
// 	VIPs := flag.String("vip", "192.168.1.100", "The IP addresses (VIPs) to monitor, separated by commas")
// 	flag.Parse()

// 	ctx := context.Background()
// 	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
// 	if err != nil {
// 		log.Fatalf("Error creating Docker client: %v", err)
// 	}

// 	// 解析命令行参数
// 	containerNameList := strings.Split(*containerNames, ",")
// 	vipList := strings.Split(*VIPs, ",")

// 	// 检查两个列表长度是否一致
// 	if len(containerNameList) != len(vipList) {
// 		log.Fatalf("Number of container names and VIPs must match")
// 	}

// 	// 为每个容器/VIP对启动一个goroutine
// 	for i, containerName := range containerNameList {
// 		vip := vipList[i]
// 		lastVipStatus, _ := readLastVipStatus(vip) // 读取上次的VIP状态

// 		go func(cn, v, lvs string) {
// 			lvsPtr := &lvs
// 			for {
// 				getContainerMetrics(ctx, cli, cn, v, lvsPtr)
// 				time.Sleep(5 * time.Second) // 每5秒更新一次指标
// 			}
// 		}(containerName, vip, lastVipStatus)
// 	}

//		// 暴露 Prometheus 指标
//		http.Handle("/metrics", promhttp.Handler())
//		port := ":2112"
//		// 启动 HTTP 服务并记录日志
//		log.Printf("Starting HTTP server on port %s", port)
//		err = http.ListenAndServe(port, nil)
//		if err != nil {
//			log.Fatalf("HTTP server failed to start: %v", err)
//		}
//	}
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"keepalived/internal/container_metrics"
	"keepalived/internal/file_utils"
	"keepalived/internal/prometheus_metrics"

	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 初始化 Docker 客户端
	containerNames := flag.String("container_name", "keepalive-master", "The names of the containers to monitor, separated by commas")
	VIPs := flag.String("vip", "192.168.1.100", "The IP addresses (VIPs) to monitor, separated by commas")
	flag.Parse()

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}

	// 解析命令行参数
	containerNameList := strings.Split(*containerNames, ",")
	vipList := strings.Split(*VIPs, ",")

	// 检查两个列表长度是否一致
	if len(containerNameList) != len(vipList) {
		log.Fatalf("Number of container names and VIPs must match")
	}

	// 为每个容器/VIP对启动一个goroutine
	for i, containerName := range containerNameList {
		vip := vipList[i]

		lastVipStatus, _ := file_utils.ReadLastVipStatus(vip) // 读取上次的VIP状态
		prometheus_metrics.VipChangeCount.WithLabelValues(vip)
		go func(cn, v, lvs string) {
			lvsPtr := &lvs
			for {
				container_metrics.GetContainerMetrics(ctx, cli, cn, v, lvsPtr)
				time.Sleep(5 * time.Second) // 每5秒更新一次指标
			}
		}(containerName, vip, lastVipStatus)
	}

	// 暴露 Prometheus 指标
	http.Handle("/metrics", promhttp.Handler())
	port := ":2112"
	// 启动 HTTP 服务并记录日志
	log.Printf("Starting HTTP server on port %s", port)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("HTTP server failed to start: %v", err)
	}
}
