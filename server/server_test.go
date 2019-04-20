package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateDir(t *testing.T) {
	fileName := "movie/movie1/1.txt"
	path := filepath.Dir(fileName)
	//递归创建目录
	err := os.MkdirAll(path, os.ModePerm)
	fmt.Println("err:", err)
}
