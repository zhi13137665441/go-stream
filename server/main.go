package main

import (
	"fmt"
	"gocv.io/x/gocv"
	"net"
	"os"
)

func main() {
	camera, err := gocv.VideoCaptureDevice(0)
	if err != nil {
		panic(err)
	}
	defer camera.Close()
	udpAddr, err := net.ResolveUDPAddr("udp4", ":22333") // 转换地址，作为服务器使用时需要监听本机的一个端口
	checkError(err)
	conn, err := net.ListenUDP("udp", udpAddr) // 启动UDP监听本机端口
	checkError(err)

	// 创建 mjpeg 流
	// 启动一个协程以处理视频流
	go func() {
		for {
			var buf1 [128]byte
			_, addr, err := conn.ReadFromUDP(buf1[:]) // 读取数据，返回值依次为读取数据长度、远端地址、错误信息 // 读取操作会阻塞直至有数据可读取
			checkError(err)
			// 读取一帧图像
			img := gocv.NewMat()
			if ok := camera.Read(&img); !ok {
				break
			}
			// 在图像上添加一些注解（如文本、线条、图形等）
			// TODO: 在图像上添加注解的代码
			// 将图像转换为 JPEG 格式
			buf, _ := gocv.IMEncode(".jpg", img)
			_, err = conn.WriteToUDP(buf.GetBytes(), addr) // 写数据，返回值依次为写入数据长度、错误信息 // WriteToUDP()并非只能用于应答的，只要有个远程地址可以随时发消息
			checkError(err)

		}
	}()
	//for {
	//	var buf [128]byte
	//	len, addr, err := conn.ReadFromUDP(buf[:]) // 读取数据，返回值依次为读取数据长度、远端地址、错误信息 // 读取操作会阻塞直至有数据可读取
	//	checkError(err)
	//	fmt.Println(string(buf[:len])) // 向终端打印收到的消息
	//
	//	_, err = conn.WriteToUDP([]byte("233~~~"), addr) // 写数据，返回值依次为写入数据长度、错误信息 // WriteToUDP()并非只能用于应答的，只要有个远程地址可以随时发消息
	//	checkError(err)
	//}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error %s", err.Error())
		os.Exit(1)
	}
}
