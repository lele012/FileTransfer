package main

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestGetDirSize(t *testing.T) {
	filePath := "E:\\BaiduDownload"

	size, err := getDirSize(filePath)
	fmt.Printf("size:%d err:%v\n", size, err)
}

func TestSplitPath(t *testing.T) {
	filePath := "c:\\test\\1"
	paths, fileName := filepath.Split(filePath)
	fmt.Println(paths)
	fmt.Println(fileName)
}
