package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	srcFile := filepath.Join(os.TempDir(), "test_copy_src.txt")
	if err := os.WriteFile(srcFile, []byte("test copy file"), os.ModePerm); err != nil {
		t.Log(err)
		t.FailNow()
	}
	defer os.Remove(srcFile)
	destFile := filepath.Join(os.TempDir(), "test_copy_dst.txt")
	if err := CopyFile(srcFile, destFile); err != nil {
		t.Log(err)
		t.FailNow()
	}
	defer os.Remove(destFile)
	data, err := os.ReadFile(destFile)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log(string(data))
}
