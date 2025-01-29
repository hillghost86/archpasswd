package archpasswd

import "errors"

// 定义包级别的错误
var (
	// ErrFileNotEncrypted 文件未加密
	ErrFileNotEncrypted = errors.New("文件未加密")
	// ErrUnsupportedFormat 不支持的文件格式
	ErrUnsupportedFormat = errors.New("不支持的文件格式")
	// ErrOpenFileFailed 打开文件失败
	ErrOpenFileFailed = errors.New("打开文件失败")
)
