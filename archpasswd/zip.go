package archpasswd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yeka/zip"
)

type zipChecker struct {
	FilePath string
}

func (c *zipChecker) CheckPassword(filePath, password string) (bool, string, error) {
	//fmt.Fprintf(os.Stderr, "\n正在尝试密码: %s\n", password)

	// 打开ZIP文件
	r, err := zip.OpenReader(c.FilePath)
	if err != nil {
		return false, password, fmt.Errorf("%w: %v", ErrOpenFileFailed, err)
	}
	defer r.Close()

	// 检查是否有加密文件
	hasEncrypted := false
	encryptedFiles := make([]*zip.File, 0)
	for _, f := range r.File {
		if f.IsEncrypted() {
			hasEncrypted = true
			encryptedFiles = append(encryptedFiles, f)
		}
	}

	if !hasEncrypted {
		// 文件未加密，返回密码为空字符串
		return true, "", nil
	}

	// 只尝试加密文件
	for _, f := range encryptedFiles {
		//fmt.Fprintf(os.Stderr, "尝试打开文件: %s\n", f.Name)
		f.SetPassword(password)
		rc, err := f.Open()
		if err != nil {
			fmt.Fprintf(os.Stderr, "打开文件失败: %v\n", err)
		}
		if err != nil {
			if strings.Contains(err.Error(), "password") {
				//fmt.Fprintf(os.Stderr, "密码错误\n")
				// 密码错误
				return false, password, nil
			}
			if strings.Contains(err.Error(), "not a valid zip file") {
				// 文件格式错误
				return false, password, ErrUnsupportedFormat
			}
			return false, password, fmt.Errorf("%w: %v", ErrOpenFileFailed, err)
		}
		defer rc.Close() // 关闭文件

		// 尝试读取内容
		_, err = io.ReadAll(rc) // 读取文件内容
		if err != nil {
			// 读取文件失败
			if strings.Contains(err.Error(), "password") ||
				strings.Contains(err.Error(), "checksum error") {
				// 密码错误或校验和错误
				//fmt.Fprintf(os.Stderr, "密码错误或校验和错误\n")
				return false, password, nil
			}
			return false, password, fmt.Errorf("%w: %v", ErrOpenFileFailed, err)
		} else {
			return true, password, nil
		}

	}

	fmt.Fprintln(os.Stderr, "所有文件尝试完毕，未找到正确密码")
	return false, password, nil
}
