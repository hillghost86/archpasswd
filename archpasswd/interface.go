package archpasswd

// Format 压缩文件格式
type Format int

const (
	ZIP Format = iota
	RAR
	SEVENZIP
)

// Checker 密码检查器接口
type Checker interface {
	// CheckPassword 检查密码是否正确
	// 参数：
	//   filePath: 文件路径
	//   password: 要检查的密码
	// 返回值：
	//   bool: 密码是否正确
	//   string: 使用的密码
	//   error: 可能的错误
	CheckPassword(filePath, password string) (bool, string, error)
}
