package utils

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	homeDir = ""
	hdonce  sync.Once
	// HOME_DIR_ENV_NAME 家目录环境变量名称
	HOME_DIR_ENV_NAME = "ASJARD_HOME_DIR"
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
	return filepath.Join(GetHomeDir(), CONF_DIR)
}

// GetCertDir 获取证书存放路径
func GetCertDir() string {
	return filepath.Join(GetHomeDir(), CONF_DIR, CERT_DIR)
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
func IsDir(dir string) bool {
	s, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// IsFile 是否为文件
func IsFile(file string) bool {
	return !IsDir(file)
}
