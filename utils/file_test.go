package utils

import (
	"os"
	"path/filepath"
	"testing"
)

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

	testFilename := "test_copy_dir_file.txt"
	testContent := []byte("test copy dir content")
	if err := os.WriteFile(filepath.Join(srcDir, testFilename),
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
