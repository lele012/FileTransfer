package main

func main() {
	host := "127.0.0.1:7010"
	// fileName := "1G.rar"s
	// UploadFileTo(host, fileName)

	resPath := "E:\\BaiduDownload"

	SendResourceTo(host, resPath, true)
}
