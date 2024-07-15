package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetWorkDir(t *testing.T) {
	_, err := GetWorkDir()
	assert.NoError(t, err)
}

func TestGetDir(t *testing.T) {
	var homeDir string
	var confDir string
	t.Run("GetHomeDir", func(t *testing.T) {
		homeDir = GetHomeDir()
		assert.NotEmpty(t, homeDir)
		// 后续如何设置并获取都是第一次的值
		tmpDir := t.TempDir()
		os.Setenv(HOME_DIR_ENV_NAME, tmpDir)
		assert.Equal(t, homeDir, GetHomeDir(), tmpDir)
	})
	t.Run("GetConfDir", func(t *testing.T) {
		confDir = t.TempDir()
		os.Setenv(CONF_DIR_ENV_NAME, confDir)
		assert.NotEmpty(t, GetConfDir())
		assert.Equal(t, confDir, GetConfDir(), confDir)
	})
	t.Run("GetCertDir", func(t *testing.T) {
		assert.NotEmpty(t, GetCertDir())
		assert.Equal(t, filepath.Join(confDir, CERT_DIR), GetCertDir())
	})
}

func TestIsPathExists(t *testing.T) {
	tmpDir := "/tmp/never_exist_dir"
	assert.Equal(t, false, IsPathExists(tmpDir), tmpDir)
	tmpDir = t.TempDir()
	assert.Equal(t, true, IsPathExists(tmpDir), tmpDir)

	tmpFile := filepath.Join(tmpDir, "test_is_path_exists")
	assert.Equal(t, false, IsPathExists(tmpFile), tmpFile)
	_, err := os.Create(tmpFile)
	assert.NoError(t, err)
	assert.Equal(t, true, IsPathExists(tmpFile), tmpFile)
}

func TestIsDir(t *testing.T) {
	tmpDir := t.TempDir()
	assert.Equal(t, true, IsDir(tmpDir), tmpDir)
	tmpFile := filepath.Join(tmpDir, "never_exist_file")
	assert.Equal(t, false, IsDir(tmpFile))
	_, err := os.Create(tmpFile)
	assert.NoError(t, err, tmpFile)
	assert.Equal(t, false, IsDir(tmpFile))
}

func TestIsFile(t *testing.T) {
	tmpDir := t.TempDir()
	assert.Equal(t, false, IsFile(tmpDir), tmpDir)
	tmpFile := filepath.Join(tmpDir, "never_exist_file")
	// 不存在的文件会返回true
	assert.Equal(t, true, IsFile(tmpFile), tmpFile)
	_, err := os.Create(tmpFile)
	assert.NoError(t, err, tmpFile)
	assert.Equal(t, true, IsFile(tmpFile))
}

func TestCopyFile(t *testing.T) {
	testContent := []byte("test copy file")
	srcFile := filepath.Join(os.TempDir(), "test_copy_src.txt")
	if err := os.WriteFile(srcFile, testContent, os.ModePerm); err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer os.Remove(srcFile)
	destFile := filepath.Join(os.TempDir(), "test_copy_dst.txt")
	if err := CopyFile(srcFile, destFile); err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer os.Remove(destFile)
	data, err := os.ReadFile(destFile)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Log(string(data))
	if string(data) != string(testContent) {
		t.Error("conent not match")
		t.FailNow()
	}
}

func TestCopyDir(t *testing.T) {
	srcDir, err := os.MkdirTemp(os.TempDir(), "test_copy_dir_src_*")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer os.RemoveAll(srcDir)
	subSrcDir, err := os.MkdirTemp(srcDir, "test_copy_dir_src_sub_*")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	testFilename := "test_copy_dir_file.txt"
	testContent := []byte("test copy dir content")
	if err := os.WriteFile(filepath.Join(srcDir, testFilename),
		testContent, os.ModePerm); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := os.WriteFile(filepath.Join(subSrcDir, testFilename),
		testContent, os.ModePerm); err != nil {
		t.Error(err)
		t.FailNow()
	}

	destDir, err := os.MkdirTemp(os.TempDir(), "test_copy_dir_dst_*")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer os.RemoveAll(destDir)
	if err := CopyDir(srcDir, destDir); err != nil {
		t.Error(err)
		t.FailNow()
	}

	data, err := os.ReadFile(filepath.Join(destDir, testFilename))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Log(string(data))

	if string(data) != string(testContent) {
		t.Error("content not match")
		t.FailNow()
	}
}
