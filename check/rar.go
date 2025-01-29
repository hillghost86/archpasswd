package check

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/nwaples/rardecode/v2"
)

// rarChecker 结构体需要导出 FilePath 字段
type rarChecker struct {
	FilePath string // 改为大写以导出
}

// rarChecker 实现 Checker 接口

func (c *rarChecker) CheckPassword(filePath, password string) (bool, string, error) {
	// 使用 c.FilePath 而不是传入的 filePath
	// 检查是否是 RAR4
	file, err := os.Open(filePath)
	if err != nil {
		return false, password, fmt.Errorf("%w: %v", ErrOpenFileFailed, err)
	}
	headerBytes := make([]byte, 7)
	_, err = io.ReadFull(file, headerBytes)
	file.Close()
	if err != nil {
		return false, password, fmt.Errorf("%w: %v", ErrOpenFileFailed, err)
	}

	// 检查是否是 RAR4 格式
	isRar4 := string(headerBytes) == "Rar!\x1a\x07\x00"

	if isRar4 {
		return c.checkRar4Password(filePath, password)
	}

	// 处理 RAR5 格式
	// 首先尝试使用空密码打开
	rr, err := rardecode.OpenReader(filePath, rardecode.Password(""))
	if err == nil {
		// 尝试读取文件头来确认是否真的未加密
		_, err = rr.Next()
		if err == nil || err == io.EOF {
			//文件未加密，使用ErrFileNotEncrypted
			return true, "", nil
		}
		rr.Close()
	}

	// 使用提供的密码尝试打开
	rr, err = rardecode.OpenReader(filePath, rardecode.Password(password))
	if err != nil {
		if err.Error() == "rardecode: bad password" ||
			err.Error() == "rardecode: incorrect password" {
			return false, password, nil
		}
		return false, password, fmt.Errorf("%w: %v", ErrOpenFileFailed, err)
	}
	defer rr.Close()

	// 尝试读取文件头
	fileHeader, err := rr.Next()
	if err != nil {
		if err == io.EOF {
			// 空档案但能打开，说明密码正确
			return true, password, nil
		}
		if err.Error() == "rardecode: bad password" ||
			err.Error() == "rardecode: incorrect password" {
			return false, password, nil
		}
		return false, password, fmt.Errorf("%w: %v", ErrOpenFileFailed, err)
	}

	// 尝试读取一些内容来验证密码
	if fileHeader != nil {
		content := make([]byte, 1024)
		_, err = rr.Read(content)
		if err != nil && err != io.EOF {
			//读取内容错误，使用ErrOpenFileFailed
			return false, password, fmt.Errorf("%w: %v", ErrOpenFileFailed, err)
		}
	}

	return true, password, nil
}

// checkRar4Password 专门处理 RAR4 格式的密码验证
func (c *rarChecker) checkRar4Password(filePath, password string) (bool, string, error) {
	// 尝试使用提供的密码打开文件
	rr, err := rardecode.OpenReader(filePath, rardecode.Password(password))
	if err != nil {
		if err.Error() == "rardecode: encrypted archive" ||
			err.Error() == "rardecode: archive encrypted, password needed" {
			return false, password, nil
		}

		if err.Error() == "rardecode: bad password" ||
			err.Error() == "rardecode: incorrect password" {
			return false, password, nil
		}

		return false, password, fmt.Errorf("%w: %v", ErrOpenFileFailed, err)
	}
	defer rr.Close()

	// 尝试读取文件头
	_, err = rr.Next()
	if err != nil {
		if err == io.EOF {
			return true, password, nil // 空档案但能打开，说明密码正确
		}
		return false, password, nil
	}

	// 创建缓冲区来存储解压的内容
	var buf bytes.Buffer
	_, err = io.CopyN(&buf, rr, 1024) // 只复制前1024字节用于验证
	buf.Reset()                       // 清空缓冲区

	if err != nil && err != io.EOF {
		return false, password, nil
	}

	// 如果能成功复制内容，说明密码正确
	return true, password, nil
}
