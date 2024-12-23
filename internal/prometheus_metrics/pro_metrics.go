package prometheus_metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus 指标定义
var (
	RestartCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "container_restart_count",
			Help: "The number of restarts of a container",
		},
		[]string{"container_id"},
	)

	ContainerStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "container_status",
			Help: "The status of the container: 1=running, 0=stopped",
		},
		[]string{"container_id"},
	)

	VipFound = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "vip_found",
			Help: "Whether the target VIP is found: 1=found, 0=not found",
		},
		[]string{"vip_res"},
	)

	VipChangeCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vip_change_count",
			Help: "Counter that increments every time the VIP address changes",
		},
		[]string{"vip_res"},
	)
)
