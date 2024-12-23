package container_metrics

import (
	"context"
	"log"

	"keepalived/internal/file_utils"
	"keepalived/internal/network_utils"
	"keepalived/internal/prometheus_metrics"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// GetContainerMetrics 获取并更新容器指标
func GetContainerMetrics(ctx context.Context, cli *client.Client, containerName string, VIP string, lastVipStatus *string) {
	// 查找容器
	args := filters.NewArgs()
	args.Add("name", "^/"+containerName+"$")
	containers, err := cli.ContainerList(ctx, container.ListOptions{Filters: args})
	if err != nil {
		log.Println("Error listing containers:", err)
		return
	}

	if len(containers) == 0 {
		log.Println("No containers found %s", containerName)
		prometheus_metrics.ContainerStatus.WithLabelValues(containerName).Set(0)
		return
	}
	containerID := containers[0].ID
	// 获取容器信息
	containerInspect, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		log.Println("Error inspecting container:", err)
		prometheus_metrics.ContainerStatus.WithLabelValues(containerName).Set(0)
		return
	}

	// 更新重启次数指标
	prometheus_metrics.RestartCount.WithLabelValues(containerName).Set(float64(containerInspect.RestartCount))

	// 更新容器状态指标
	if containerInspect.State.Running {
		prometheus_metrics.ContainerStatus.WithLabelValues(containerName).Set(1)
	} else {
		prometheus_metrics.ContainerStatus.WithLabelValues(containerName).Set(0)
	}

	// 查找 VIP
	vipFoundValue := network_utils.FindSpecificIP(VIP) // 检查VIP是否存在
	var currentVipStatus string
	if vipFoundValue == 0 {
		prometheus_metrics.VipFound.WithLabelValues(VIP).Set(1)
		currentVipStatus = "found"
	} else {
		prometheus_metrics.VipFound.WithLabelValues(VIP).Set(0)
		currentVipStatus = "not_found"
	}

	// 第一次运行时记录状态
	if *lastVipStatus == "" {
		*lastVipStatus = currentVipStatus
		file_utils.WriteVipStatusToFile(currentVipStatus, VIP)
		return
	}

	// 检查VIP状态是否发生变化
	if currentVipStatus != *lastVipStatus {
		log.Printf("VIP %s status changed from %s to %s", VIP, *lastVipStatus, currentVipStatus)
		prometheus_metrics.VipChangeCount.WithLabelValues(VIP).Inc() // 增加VIP变化计数

		file_utils.WriteVipStatusToFile(currentVipStatus, VIP) // 更新文件中的状态
		*lastVipStatus = currentVipStatus                      // 更新lastVipStatus
	}
}
