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
