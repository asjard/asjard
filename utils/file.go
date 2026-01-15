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
	// Cached paths with sync.Once to ensure thread-safe, single initialization.
	homeDir = ""
	hdonce  sync.Once
	confDir = ""
	cdonce  sync.Once

	// Environmental variable keys for overriding default paths.
	HOME_DIR_ENV_NAME = "ASJARD_HOME_DIR"
	CONF_DIR_ENV_NAME = "ASJARD_CONF_DIR"

	// Default sub-directory names within the Home directory.
	CONF_DIR = "conf"
	CERT_DIR = "certs"
)

// GetWorkDir returns the absolute path of the directory containing the running executable.
func GetWorkDir() (string, error) {
	wd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	return wd, nil
}

// GetHomeDir determines the root directory of the application.
// Priority: Environment Variable > Executable Work Directory.
func GetHomeDir() string {
	hdonce.Do(func() {
		homeDir = os.Getenv(HOME_DIR_ENV_NAME)
		if homeDir == "" {
			wd, err := GetWorkDir()
			if err != nil {
				panic(err) // Critical failure if work dir cannot be resolved.
			}
			homeDir = wd
		}
	})
	return homeDir
}

// GetConfDir returns the configuration directory path.
// Priority: Environment Variable > {HomeDir}/conf.
func GetConfDir() string {
	cdonce.Do(func() {
		confDir = strings.TrimSpace(os.Getenv(CONF_DIR_ENV_NAME))
		if confDir == "" {
			confDir = filepath.Join(GetHomeDir(), CONF_DIR)
		}
	})
	return confDir
}

// GetCertDir returns the directory where security certificates are stored.
func GetCertDir() string {
	return filepath.Join(GetConfDir(), CERT_DIR)
}

// IsPathExists checks if a file or directory exists at the given path.
func IsPathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// IsDir returns true if the path exists and is a directory.
func IsDir(dir string) bool {
	s, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// IsFile returns true if the path exists and is not a directory.
func IsFile(file string) bool {
	return !IsDir(file)
}

// CopyFile copies a single file from src to dest with standard permissions.
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

	// Efficiently stream the file content.
	if _, err := io.Copy(d, s); err != nil {
		return err
	}
	return nil
}

// CopyDir recursively copies a directory and all its contents.
func CopyDir(srcDir, destDir string) error {
	if srcDir == destDir {
		return nil
	}
	items, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	// Create destination with restrictive permissions (0750).
	if err := os.MkdirAll(destDir, 0750); err != nil {
		return err
	}
	for _, item := range items {
		src := filepath.Join(srcDir, item.Name())
		dst := filepath.Join(destDir, item.Name())
		if !item.IsDir() {
			if err := CopyFile(src, dst); err != nil {
				return err
			}
			continue
		}
		// Recursive call for sub-directories.
		if err := CopyDir(src, dst); err != nil {
			return err
		}
	}
	return nil
}

// FileMD5 calculates the MD5 checksum of a file to verify integrity.
func FileMD5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// MergeFile is a wrapper that merges multiple input files into one output file.
func MergeFile(ctx context.Context, inputFiles []string, outputFile string, concurrency int) error {
	return MergeFiles(ctx, append([]string{outputFile}, inputFiles...), concurrency)
}

// MergeFiles implements a recursive binary merge.
// It merges pairs of files in parallel to optimize I/O and CPU usage.
func MergeFiles(ctx context.Context, files []string, concurrency int) error {
	eg, ctx := errgroup.WithContext(ctx)
	if concurrency == 0 {
		concurrency = runtime.NumCPU()
	}
	eg.SetLimit(concurrency)

	n := len(files)
	if n <= 1 {
		return nil
	}

	newFiles := make([]string, 0, (n+1)/2)
	for j := 0; j < n; j += 2 {
		i := j
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if i+1 >= n {
					// Single file left, nothing to merge it with.
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
	// Recursively merge the newly merged files until only one remains.
	return MergeFiles(ctx, newFiles, concurrency)
}

// mergefile appends the content of src to dst.
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

// SplitFile divides a large file into multiple smaller chunks (parts).
// Useful for multi-part uploads or distributed processing.
func SplitFile(srcFile, dstDir string, chunkSize int64) ([]string, error) {
	var parts []string
	f, err := os.Open(srcFile)
	if err != nil {
		return parts, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return parts, err
	}

	totalPartsNum := int64(math.Ceil(float64(fi.Size()) / float64(chunkSize)))
	for i := int64(0); i < totalPartsNum; i++ {
		// Calculate precise size for the current chunk.
		partSize := int(math.Min(float64(chunkSize), float64(fi.Size()-i*chunkSize)))
		partBuffer := make([]byte, partSize)

		_, err := f.Read(partBuffer)
		if err != nil && err != io.EOF {
			return parts, err
		}

		fileName := filepath.Join(dstDir, "part_"+strconv.Itoa(int(i)))
		parts = append(parts, fileName)

		// Write the chunk to a new file.
		if err := os.WriteFile(fileName, partBuffer, 0640); err != nil {
			return parts, err
		}
	}
	return parts, nil
}
