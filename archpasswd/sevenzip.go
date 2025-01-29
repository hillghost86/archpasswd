package archpasswd

import (
	"fmt"
	"io"
	"strings"

	"github.com/bodgit/sevenzip"
)

// sevenZipChecker 实现了7z格式文件的密码检查
type sevenZipChecker struct {
	FilePath string
	// 可以在这里添加其他配置项
}

func (c *sevenZipChecker) testFirstFile(r *sevenzip.ReadCloser, password string) (bool, string, error) {
	// 查找第一个非空且最小的文件
	var selectedFile *sevenzip.File
	const maxSize = 512 // 限制文件大小为512字节

	for _, f := range r.File {
		if !f.FileInfo().IsDir() && f.UncompressedSize > 0 && f.UncompressedSize <= maxSize {
			selectedFile = f
			break
		}
	}

	// 如果没找到小文件，就用第一个非空文件
	if selectedFile == nil {
		for _, f := range r.File {
			if !f.FileInfo().IsDir() && f.UncompressedSize > 0 {
				selectedFile = f
				break
			}
		}
	}

	// 如果找到了文件，尝试读取
	if selectedFile != nil {
		rc, err := selectedFile.Open()
		if err != nil {
			// 简化错误处理，只要打开失败就认为是密码错误
			return false, password, nil
		}
		defer rc.Close()

		// 尝试读取内容
		buffer := make([]byte, 1)
		_, err = io.ReadFull(rc, buffer)
		if err != nil {
			// 读取失败也认为是密码错误
			return false, password, nil
		}

		return true, password, nil // 密码正确
	}

	return true, password, nil // 空档案但能打开，说明密码正确
}

func (c *sevenZipChecker) CheckPassword(filePath, password string) (bool, string, error) {
	// 检查文件扩展名
	if !strings.HasSuffix(strings.ToLower(filePath), ".7z") &&
		!strings.HasSuffix(strings.ToLower(filePath), ".001") {
		return false, "", fmt.Errorf("不支持的文件格式，仅支持 .7z 或 .001 文件")
	}

	// 总是先尝试空密码
	r, err := sevenzip.OpenReaderWithPassword(c.FilePath, "")
	if err == nil {
		defer r.Close()
		ok, _, err := c.testFirstFile(r, "")
		if err != nil {
			return false, "", err
		}
		if ok {
			return true, "", nil // 确认是空密码文件
		}
	}

	// 如果指定了密码，尝试使用指定密码
	if password != "" {
		r, err := sevenzip.OpenReaderWithPassword(c.FilePath, password)
		if err != nil {
			// 如果打开失败，认为是密码错误
			return false, password, nil
		}
		defer r.Close()

		return c.testFirstFile(r, password)
	}

	return false, password, nil // 如果空密码测试失败，返回失败
}
