# ArchPasswd

ArchPasswd 是一个用于检查压缩文件密码的 Go 语言库。

## 功能特性

- 支持 7z 格式文件的密码验证
- 支持 rar 格式文件的密码验证
- 支持 zip 格式文件的密码验证
- 支持空密码检测
- 优化的性能，快速验证密码
- 简单易用的 API

## 安装

```bash 
go get github.com/hillghost86/archpasswd
```

## 使用示例

```go
package main

import (
	"fmt"
	"github.com/hillghost86/archpasswd"
)

func main() {       
    // 创建一个7z格式检查器
    checker, err := archpasswd.NewChecker(archpasswd.TYPE_7Z, "test.7z")
        if err != nil {
        panic(err)
    }
    // 检查密码
    ok, password, err := checker.CheckPassword("test.7z", "123456")
        if err != nil {
        panic(err)
    }

    if ok {
        fmt.Printf("密码正确: %s\n", password)
    } else {
        fmt.Println("密码错误")
    }
}
```

## API 文档

### NewChecker 创建一个压缩文件密码检查器

```go
func NewChecker(filePath string, format ...Format) (Checker, error)
```

创建一个新的压缩文件密码检查器。

参数：
- format: 压缩文件格式（目前支持 TYPE_7Z）
- filePath: 压缩文件路径

### CheckPassword 检查指定的密码是否正确

```go
func (c *sevenZipChecker) CheckPassword(data []byte, password string) bool
```

检查指定的密码是否正确。

参数：
- filePath: 压缩文件路径
- password: 要检查的密码

返回值：
- bool: 密码是否正确
- string: 正确的密码（如果找到）
- error: 错误信息

## 性能说明

- 密码验证速度主要受文件大小和加密算法影响
- 程序会优先选择较小的文件进行验证以提高性能
- 对于大文件，每次密码验证可能需要约100ms

## 版本历史
 见[CHANGELOG.md](CHANGELOG.md)

## 许可证

[您的许可证类型]

## 贡献

欢迎提交 Issue 和 Pull Request！

## 作者

[hillghost86](https://github.com/hillghost86)
## 致谢

- [github.com/bodgit/sevenzip](https://github.com/bodgit/sevenzip) - 7z 文件处理库
- [github.com/nwaples/rardecode](https://github.com/nwaples/rardecode) - rar 文件处理库
- [github.com/yeka/zip](https://github.com/yeka/zip) - zip 文件处理库
- [github.com/h2non/filetype](https://github.com/h2non/filetype) - 文件类型检测库
- [github.com/gabriel-vasile/mimetype](https://github.com/gabriel-vasile/mimetype) - MIME 类型检测库


