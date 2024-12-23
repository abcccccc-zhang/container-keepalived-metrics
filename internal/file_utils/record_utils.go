package file_utils

import (
	"fmt"
	"log"
	"os"
)

// WriteVipStatusToFile 记录 VIP 状态到文件
func WriteVipStatusToFile(status, VIP string) {
	file, err := os.OpenFile(fmt.Sprintf("/tmp/vip_status_%s.txt", VIP), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Error opening VIP status file: %v", err)
	}
	defer file.Close()

	// 先清空文件内容，再写入新的状态
	_, err = file.Seek(0, 0)
	if err != nil {
		log.Fatalf("Error seeking in VIP status file: %v", err)
	}
	_, err = file.WriteString(status + "\n")
	if err != nil {
		log.Fatalf("Error writing to VIP status file: %v", err)
	}

	log.Printf("VIP %s status updated to: %s", VIP, status)
}

// ReadLastVipStatus 读取上次的 VIP 状态
func ReadLastVipStatus(VIP string) (string, error) {
	file, err := os.OpenFile(fmt.Sprintf("/tmp/vip_status_%s.txt", VIP), os.O_RDONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("no previous VIP status file for %s, starting fresh", VIP)
	}
	defer file.Close()

	var status string
	_, err = fmt.Fscanf(file, "%s\n", &status)
	if err != nil {
		return "", fmt.Errorf("error reading VIP status: %v", err)
	}

	return status, nil
}
