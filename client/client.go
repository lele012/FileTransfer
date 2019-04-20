package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"time"
)

// if the resource is a file like "c:/test/1.txt,ResourceName" should be "1.txt"
// if the resource is a dir like "c:/test",ResourceName should be "test",all the
// file|dir in the resource, the path prefix is "test/",
type ResourceInfo struct {
	ResName string `json:"name"`
	ResSize int64  `json:"size"`
}

// send resource to server by hostAddr
// hostAddr : the server's address,like 119.98.160.164:7010
// resPath : the absolute|relative path of the resource
// isDir : identity of the resource , which is file(false) or folder(true)
func SendResourceTo(hostAddr, resPath string, isDir bool) error {
	// connect server
	conn, err := net.DialTimeout("tcp", hostAddr, time.Second*15)
	if err != nil {
		fmt.Printf("upload res %q to %q failed when dial error: %v\n", resPath, hostAddr, err)
		return err
	}
	defer conn.Close()

	// print the progress
	speed := NewSpeed()
	defer speed.Close()

	// send ResourceInfo to server
	_, name := filepath.Split(resPath)
	size, err := getPathSize(resPath, isDir)
	if err != nil {
		fmt.Println("get path size fail:", err)
		return err
	}
	info := ResourceInfo{name, size}
	b, err := json.Marshal(info)
	if err != nil {
		fmt.Println("json Marshal fail:", err)
		return err
	}

	n, err := conn.Write(b)
	if err != nil || n != len(b) {
		fmt.Println("send resource info to server fail:", err)
		return err
	}

	// wait for receiving response from server
	// read the size from response
	var responseSize int64
	err = binary.Read(conn, binary.BigEndian, &responseSize)
	if err != nil {
		fmt.Println("recv size from server fail:", err)
		return err
	}

	if responseSize != size {
		return fmt.Errorf("send pathsize %d doesn't equal to receive pathsize %d", size, responseSize)
	}

	if isDir {
		return sendDir(conn, resPath, name, speed)
	}

	return sendFile(conn, resPath, name, speed)
}

// get the size of file|dir
func getPathSize(path string, isDir bool) (int64, error) {
	if isDir {
		return getDirSize(path)
	}

	stat, err := os.Stat(path)
	return stat.Size(), err
}

// get the size of dir
func getDirSize(pathname string) (int64, error) {
	var total int64
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		return total, err
	}
	for _, fi := range rd {
		if fi.IsDir() {
			size, err := getDirSize(pathname + string(os.PathSeparator) + fi.Name())
			if err != nil {
				return size, err
			}
			total += size
		} else {
			total += fi.Size()
		}
	}

	return total, nil
}

// send file to server
func sendFile(conn net.Conn, filePath, sendName string, speed *Speed) error {
	size, err := getPathSize(filePath, false)
	if err != nil {
		fmt.Printf("get size of %s fail:%v\n", filePath, err)
		return err
	}
	info := ResourceInfo{sendName, size}
	b, err := json.Marshal(info)
	if err != nil {
		fmt.Println("json Marshal fail:", err)
		return err
	}

	n, err := conn.Write(b)
	if err != nil || n != len(b) {
		fmt.Println("send resource info to server fail:", err)
		return err
	}

	// read the offset (the value which had sended last time)
	var offset int64
	err = binary.Read(conn, binary.BigEndian, &offset)
	if err != nil {
		fmt.Printf("recv server %q ack file %q error: %v\n", conn.RemoteAddr().String(), filePath, err)
		return err
	}

	fp, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("open file %q error: %v\n", filePath, err)
		return err
	}
	defer fp.Close()

	// endpoint continuingly
	_, err = fp.Seek(offset, io.SeekStart)
	if err != nil {
		fmt.Printf("seek file to %v error: %v\n", offset, err)
		return err
	}

	_, err = io.Copy(io.MultiWriter(conn, speed), fp)
	//don't show progress output
	//nc, err := io.Copy(conn, fp)
	if err != nil {
		if err != io.EOF {
			fmt.Printf("send to %q file %q error: %v\n", conn.RemoteAddr().String(), filePath, err)
			return err
		}
	}

	// wait for receiving response from server
	// read the filesize from server
	var responseSize int64
	err = binary.Read(conn, binary.BigEndian, &responseSize)
	if err != nil {
		fmt.Println("recv size from server fail:", err)
		return err
	}

	if responseSize != size {
		return fmt.Errorf("send pathsize %d doesn't equal to receive pathsize %d", size, responseSize)
	}
	return nil
}

// send dir to server
func sendDir(conn net.Conn, dirPath, rootPath string, speed *Speed) error {
	rd, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, fi := range rd {
		childPath := path.Join(dirPath, fi.Name())
		sendPath := path.Join(rootPath, fi.Name())
		if fi.IsDir() {
			if err := sendDir(conn, childPath, sendPath, speed); err != nil {
				return err
			}
		} else {
			if err := sendFile(conn, childPath, sendPath, speed); err != nil {
				return err
			}
		}
	}

	return nil
}
