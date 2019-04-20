package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path"
)

type ResourceInfo struct {
	ResName string `json:"name"`
	ResSize int64  `json:"size"`
}

const (
	DefaultBufferSize = 10 * 1024
)

func Handle(conn net.Conn, downloadPath string) {
	defer conn.Close()
	err := ReceiveResource(conn, downloadPath)
	if err != nil {
		fmt.Println(err)
	}
}

func ReceiveResource(conn net.Conn, downloadPath string) error {
	var info ResourceInfo
	speed := NewSpeed()
	defer speed.Close()

	// read the resourceinfo
	b := make([]byte, 512)
	n, err := conn.Read(b)
	if err != nil {
		if err != io.EOF {
			fmt.Println("recv resource error:", err)
		}
		return err
	}

	err = json.Unmarshal(b[:n], &info)
	if err != nil {
		fmt.Println("json Unmarshal fail!")
	}

	// send the resourcesize to client
	err = binary.Write(conn, binary.BigEndian, info.ResSize)
	if err != nil {
		return err
	}
	var receivedSize int64
	for {
		size, err := receiveFile(conn, downloadPath, speed)
		if err != nil {
			return err
		}
		receivedSize += size
		if receivedSize >= info.ResSize {
			break
		}
	}
	return nil
}

func receiveFile(conn net.Conn, downloadPath string, speed *Speed) (int64, error) {
	var info ResourceInfo
	// read the resourceinfo
	b := make([]byte, 512)
	n, err := conn.Read(b)
	if err != nil {
		if err != io.EOF {
			fmt.Println("recv file name error:", err)
		}
		return 0, err
	}

	err = json.Unmarshal(b[:n], &info)
	if err != nil {
		fmt.Println("json Unmarshal fail!")
		return 0, err
	}
	filePath := path.Join(downloadPath, info.ResName)
	// creat dir
	dir, _ := path.Split(filePath)
	if err = os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Printf("create path %s recursively error:%v\n", dir, err)
		return 0, err
	}

	// create file
	fp, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("create file %q error: %v\n", filePath, err)
		return 0, err
	}
	defer fp.Close()
	fi, err := fp.Stat()
	if err != nil {
		fmt.Printf("get file %q stat error: %v", filePath, err)
		return 0, err
	}
	defer fp.Sync()

	err = binary.Write(conn, binary.BigEndian, fi.Size())
	if err != nil {
		return 0, err
	}

	buffer := make([]byte, DefaultBufferSize)
	restRecv := info.ResSize - fi.Size()

	for restRecv > 0 {
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("copy file %q error: %v\n", filePath, err)
			}
			return 0, err
		}
		restRecv -= int64(n)
		// record progress
		speed.Write(buffer[:n])
		if _, err = fp.Write(buffer[:n]); err != nil {
			return 0, err
		}
	}

	// send the resourcesize to client
	err = binary.Write(conn, binary.BigEndian, info.ResSize)
	if err != nil {
		return 0, err
	}

	return info.ResSize, nil
}
