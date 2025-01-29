package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"archpasswd/archpasswd"
)

// TestArchivePassword 测试单个密码
// 参数：
//
//		filePath: 文件路径（必需）
//		password: 要测试的密码（必需）
//		fileType: 文件类型（可选，默认自动检测）
//	            可用值：TYPE_ZIP, TYPE_RAR, TYPE_7Z, TYPE_ZIP_PART, TYPE_RAR_PART, TYPE_7Z_PART
func TestArchivePassword(filePath string, password string, fileType ...archpasswd.Format) (bool, string, error) {
	// 创建对应类型的密码检查器
	checker, err := archpasswd.NewChecker(filePath, fileType...)
	if err != nil {
		return false, password, err
	}

	// 直接返回检查器的结果
	return checker.CheckPassword(filePath, password)
}

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("获取工作目录失败: %v", err)
	}

	filePath := filepath.Join(pwd, "testdata", "test7z999-64.zip")
	fmt.Println(filePath)

	// 定义要使用的文件类型
	var fileType archpasswd.Format = archpasswd.TYPE_ZIP

	// 测试密码列表，包含正确密码 "666"
	passwords := []string{"885", "666", "65418", "23423asdfgasdf", "998", "999"}

	startTime := time.Now()
	totalChecked := 0

	// 循环测试密码
	for _, password := range passwords {
		totalChecked++
		isCorrect, foundPassword, err := TestArchivePassword(filePath, password, fileType)
		if err != nil {
			fmt.Printf("检查失败: %v\n", err)
			continue
		}
		if isCorrect {
			duration := time.Since(startTime)
			fmt.Printf("密码正确: %s (用时: %v, 尝试: %d次)\n",
				foundPassword, duration, totalChecked)
			return
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("未找到正确密码 (用时: %v, 尝试: %d次)\n",
		duration, totalChecked)
}
