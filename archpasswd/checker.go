package archpasswd

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/h2non/filetype"
)

var ErrInvalidArgument = errors.New("invalid argument: file path is required")

// Format 的特殊值定义
const (
	FORMAT_UNKNOWN Format = -1 // 未知格式
	FORMAT_AUTO    Format = 0  // 自动检测
)

// NewChecker 创建密码检查器
// 参数：
//
//	filePath: 文件路径（必需）
//	format: 压缩文件格式（可选，默认为 FORMAT_AUTO）
//
// 返回值：
//
//	Checker: 密码检查器
//	error: 可能的错误
func NewChecker(filePath string, format ...Format) (Checker, error) {
	// 检查必需参数
	if filePath == "" {
		return nil, ErrInvalidArgument
	}

	// 处理可选的格式参数
	actualFormat := FORMAT_AUTO
	if len(format) > 0 {
		actualFormat = format[0]
	}

	// 获取第一个分卷的路径
	firstVolume, err := findFirstVolume(filePath)
	if err != nil {
		return nil, err
	}
	filePath = firstVolume

	// 如果用户没有指定格式，进行自动检测
	if actualFormat == FORMAT_AUTO {
		actualFormat = Format(getFileType(filePath))
	}

	// 根据格式创建对应的检查器
	switch actualFormat {
	case TYPE_ZIP, TYPE_ZIP_PART:
		return &zipChecker{FilePath: filePath}, nil
	case TYPE_RAR, TYPE_RAR_PART:
		return &rarChecker{FilePath: filePath}, nil
	case TYPE_7Z, TYPE_7Z_PART:
		return &sevenZipChecker{FilePath: filePath}, nil
	default:
		return nil, ErrUnsupportedFormat
	}
}

// internal 内部使用的通用函数和类型
// type internal struct{}

// 支持的文件类型
const (
	TYPE_ZIP      = iota // .zip
	TYPE_RAR             // .rar
	TYPE_7Z              // .7z
	TYPE_ZIP_PART        // .zip.001, .z01 等分卷
	TYPE_RAR_PART        // .part1.rar, .r01 等分卷
	TYPE_7Z_PART         // .7z.001 等分卷
	TYPE_GZ              // .gz, .tar.gz, .tgz
	TYPE_BZ2             // .bz2, .tar.bz2, .tbz2
	TYPE_TAR             // .tar
	TYPE_TAR_PART        // .tar.001, .tar.002 等分卷
	TYPE_XZ              // .xz, .tar.xz, .txz
	TYPE_CAB             // .cab
	TYPE_ISO             // .iso
	TYPE_ARJ             // .arj
	TYPE_LZH             // .lzh, .lha
	TYPE_WIM             // .wim, .swm (分段 WIM)
)

// 其他内部辅助函数...

// 函数说明：获取文件类型
// 参数：
// path: 文件路径
// 返回：文件类型
func getFileType(path string) int {
	// 1. 首先检查分卷格式（通过文件名）
	baseName := strings.ToLower(filepath.Base(path))

	// 检查7z分卷（支持任意序号）
	if matched, _ := regexp.MatchString(`\.7z\.\d{3}$`, baseName); matched {
		return TYPE_7Z_PART
	}

	if strings.Contains(baseName, ".zip.") || strings.HasSuffix(baseName, ".z01") {
		return TYPE_ZIP_PART
	}
	if strings.Contains(baseName, ".part") && strings.HasSuffix(baseName, ".rar") ||
		strings.HasSuffix(baseName, ".r01") {
		return TYPE_RAR_PART
	}
	if matched, _ := regexp.MatchString(`\.tar\.\d{3}$`, baseName); matched {
		return TYPE_TAR_PART
	}

	// 2. 读取文件头（只读取前 8KB）
	file, err := os.Open(path)
	if err != nil {
		return -1
	}
	defer file.Close()

	// 只读取文件头部分
	header := make([]byte, 8192)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return -1
	}
	header = header[:n]

	// 使用 filetype 库检测文件类型
	kind, err := filetype.Match(header)
	if err == nil && kind != filetype.Unknown {
		switch kind.MIME.Value {
		case "application/zip":
			return TYPE_ZIP
		case "application/x-rar-compressed":
			return TYPE_RAR
		case "application/x-7z-compressed":
			return TYPE_7Z
		case "application/gzip":
			return TYPE_GZ
		case "application/x-bzip2":
			return TYPE_BZ2
		case "application/x-tar":
			return TYPE_TAR
		case "application/x-xz":
			return TYPE_XZ
		case "application/vnd.ms-cab-compressed":
			return TYPE_CAB
		case "application/x-iso9660-image":
			return TYPE_ISO
		}
	}

	// 3. 如果文件类型检测失败，回退到扩展名检测
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".zip":
		return TYPE_ZIP
	case ".rar":
		return TYPE_RAR
	case ".7z":
		return TYPE_7Z
	case ".gz", ".tgz":
		return TYPE_GZ
	case ".bz2", ".tbz2":
		return TYPE_BZ2
	case ".tar":
		return TYPE_TAR
	case ".xz", ".txz":
		return TYPE_XZ
	case ".cab":
		return TYPE_CAB
	case ".iso":
		return TYPE_ISO
	case ".arj":
		return TYPE_ARJ
	case ".lzh", ".lha":
		return TYPE_LZH
	case ".wim", ".swm":
		return TYPE_WIM
	}

	// 4. 检查特殊格式
	if strings.HasSuffix(baseName, ".tar.gz") {
		return TYPE_GZ
	}
	if strings.HasSuffix(baseName, ".tar.bz2") {
		return TYPE_BZ2
	}
	if strings.HasSuffix(baseName, ".tar.xz") {
		return TYPE_XZ
	}

	return -1
}

// findFirstVolume 获取第一个分卷的路径
// 内部函数，供具体实现使用
func findFirstVolume(archivePath string) (string, error) {
	baseName := filepath.Base(archivePath)
	baseDir := filepath.Dir(archivePath)

	// 7Z 分卷 (.7z.001, .7z.002, ...)
	if matched, _ := regexp.MatchString(`\.7z\.\d{3}$`, baseName); matched {
		baseFile := baseName[:len(baseName)-7] // 移除 .7z.NNN
		return filepath.Join(baseDir, baseFile+".7z.001"), nil
	}

	// ZIP 分卷格式1 (.zip.001, .zip.002, ...)
	if matched, _ := regexp.MatchString(`\.zip\.\d{3}$`, baseName); matched {
		baseFile := baseName[:len(baseName)-8] // 移除 .zip.NNN
		firstPart := filepath.Join(baseDir, baseFile+".zip.001")
		if _, err := os.Stat(firstPart); err == nil {
			return firstPart, nil
		}
	}

	// ZIP 分卷格式2 (.zip, .z01, .z02, ...)
	if matched, _ := regexp.MatchString(`\.z\d{2}$`, baseName); matched {
		baseFile := baseName[:len(baseName)-4] // 移除 .zNN
		firstPart := filepath.Join(baseDir, baseFile+".zip")
		if _, err := os.Stat(firstPart); err == nil {
			return firstPart, nil
		}
	}

	// RAR 分卷 (.part1.rar, .part2.rar, ...)
	if matched, _ := regexp.MatchString(`\.part\d+\.rar$`, baseName); matched {
		baseFile := strings.Split(baseName, ".part")[0]
		return filepath.Join(baseDir, baseFile+".part1.rar"), nil
	}

	// RAR 旧格式分卷 (.r01, .r02, ...)
	if matched, _ := regexp.MatchString(`\.r\d{2}$`, baseName); matched {
		baseFile := baseName[:len(baseName)-4] // 移除 .rNN
		return filepath.Join(baseDir, baseFile+".rar"), nil
	}

	// 如果是 .zip 文件，检查是否是分卷的主文件
	if strings.HasSuffix(baseName, ".zip") {
		// 检查是否存在 .z01 文件
		baseFile := baseName[:len(baseName)-4] // 移除 .zip
		z01File := filepath.Join(baseDir, baseFile+".z01")
		if _, err := os.Stat(z01File); err == nil {
			return archivePath, nil // 这是分卷的主文件
		}
	}

	// 如果不是分卷，返回原始路径
	return archivePath, nil
}
