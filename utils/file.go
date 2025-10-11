package utils

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

var (
	homeDir = ""
	hdonce  sync.Once
	confDir = ""
	cdonce  sync.Once
	// HOME_DIR_ENV_NAME 家目录环境变量名称
	HOME_DIR_ENV_NAME = "ASJARD_HOME_DIR"
	// CONF_DIR_ENV_NAME 配置目录环境变量名称
	CONF_DIR_ENV_NAME = "ASJARD_CONF_DIR"
	// CONF_DIR 配置文件目录名称
	CONF_DIR = "conf"
	// CERT_DIR 证书存放路径
	CERT_DIR = "certs"
)

// GetWorkDir 获取当前工作目录
func GetWorkDir() (string, error) {
	wd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	return wd, nil
}

// GetHomeDir 获取家目录
func GetHomeDir() string {
	hdonce.Do(func() {
		homeDir = os.Getenv(HOME_DIR_ENV_NAME)
		if homeDir == "" {
			wd, err := GetWorkDir()
			if err != nil {
				panic(err)
			}
			homeDir = wd
		}
	})
	return homeDir
}

// GetConfDir 获取配置目录
func GetConfDir() string {
	cdonce.Do(func() {
		confDir = strings.TrimSpace(os.Getenv(CONF_DIR_ENV_NAME))
		if confDir == "" {
			confDir = filepath.Join(GetHomeDir(), CONF_DIR)
		}
	})
	return confDir
}

// GetCertDir 获取证书存放路径
func GetCertDir() string {
	return filepath.Join(GetConfDir(), CERT_DIR)
}

// IsPathExists 目录或文件是否存在
func IsPathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// IsDir 是否为目录
// 不存在的目录会返回false
func IsDir(dir string) bool {
	s, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// IsFile 是否为文件
// 不存在的文件会返回true
func IsFile(file string) bool {
	return !IsDir(file)
}

// CopyFile 拷贝文件
func CopyFile(srcPath, destPath string) error {
	s, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer d.Close()
	if _, err := io.Copy(d, s); err != nil {
		return err
	}
	return nil
}

// CopyDir 拷贝目录
func CopyDir(srcDir, destDir string) error {
	if srcDir == destDir {
		return nil
	}
	items, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(destDir, 0750); err != nil {
		return err
	}
	for _, item := range items {
		if !item.IsDir() {
			if err := CopyFile(filepath.Join(srcDir, item.Name()), filepath.Join(destDir, item.Name())); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Join(destDir, item.Name()), os.ModePerm); err != nil {
			return err
		}
		if err := CopyDir(filepath.Join(srcDir, item.Name()), filepath.Join(destDir, item.Name())); err != nil {
			return err
		}
	}
	return nil
}

// FileMD5 计算文件MD5
func FileMD5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// MergeFile 多个小文件合并为一个大文件
func MergeFile(ctx context.Context, inputFiles []string, outputFile string, concurrency int) error {
	return MergeFiles(ctx, append([]string{outputFile}, inputFiles...), concurrency)
}

// MergeFiles 所有文件都会合并到列表的第一个文件中
func MergeFiles(ctx context.Context, files []string, concurrency int) error {
	eg, ctx := errgroup.WithContext(ctx)
	if concurrency == 0 {
		concurrency = runtime.NumCPU()
	}
	eg.SetLimit(concurrency)
	n := len(files)
	if n == 1 {
		return nil
	}
	newFiles := make([]string, 0, n)
	for j := 0; j < n; j = j + 2 {
		i := j
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return nil
			default:
				if i+1 >= n {
					return mergefile(files[i], "")
				}
				return mergefile(files[i], files[i+1])
			}
		})
		newFiles = append(newFiles, files[i])
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return MergeFiles(ctx, newFiles, concurrency)
}

func mergefile(dst, src string) error {
	if src != "" {
		df, err := os.OpenFile(dst, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0640)
		if err != nil {
			return err
		}
		defer df.Close()
		sf, err := os.Open(src)
		if err != nil {
			return err
		}
		defer sf.Close()
		if _, err := io.Copy(df, sf); err != nil {
			return fmt.Errorf("copy file %s to %s fail[%s]", src, dst, err.Error())
		}
	}
	return nil
}

// SplitFile 分隔文件，一个大文件切分为多个小文件
func SplitFile(srcFile, dstDir string, chunkSize int64) ([]string, error) {
	var parts []string
	// 打开原始文件
	f, err := os.Open(srcFile)
	if err != nil {
		return parts, err
	}
	defer f.Close()

	// 获取原始文件信息
	fi, err := f.Stat()
	if err != nil {
		return parts, err
	}
	totalPartsNum := int64(math.Ceil(float64(fi.Size()) / float64(chunkSize)))
	for i := int64(0); i < totalPartsNum; i++ {
		partSize := int(math.Min(float64(chunkSize), float64(fi.Size()-i*chunkSize)))
		partBuffer := make([]byte, partSize)
		f.Read(partBuffer)
		fileName := filepath.Join(dstDir, "part_"+strconv.Itoa(int(i)))
		parts = append(parts, fileName)
		_, err := os.Create(fileName)
		if err != nil {
			return parts, err
		}
		if err := os.WriteFile(fileName, partBuffer, os.ModeAppend); err != nil {
			return parts, err
		}
	}
	return parts, nil
}
