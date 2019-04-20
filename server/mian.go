package main

import (
	"fmt"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":7010")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("serve start listen %v\n", 7010)
	downloadPath := "D:\\迅雷下载"
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}
		// 并发时不显示进度信息，处理单个上传可以显示终端进度信息
		go Handle(conn, downloadPath)
	}
}
