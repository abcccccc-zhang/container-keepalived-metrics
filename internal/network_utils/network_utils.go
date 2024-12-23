package network_utils

import (
	"log"
	"net"
)

// FindSpecificIP 查找特定的 VIP IP 地址
func FindSpecificIP(targetIP string) int {
	// 获取宿主机的所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatalf("Error getting interfaces: %v", err)
	}

	// 遍历网络接口，查找目标 IP 地址
	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Printf("Error getting addresses for interface %v: %v", i.Name, err)
			continue
		}

		for _, addr := range addrs {
			// 检查是否为 IPv4 地址
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}

			// 排除回环地址
			if ip.IsLoopback() {
				continue
			}

			// 如果找到了目标 IP 地址，返回该地址
			if ip.String() == targetIP {
				return 0
			}
		}
	}

	// 如果没有找到目标 IP 地址，返回空字符串
	return 1
}
