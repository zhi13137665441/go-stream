package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
)

var addrMap map[string]struct {
	addr    *net.UDPAddr
	channel *chan []byte
}

func init() {
	addrMap = map[string]struct {
		addr    *net.UDPAddr
		channel *chan []byte
	}{}
}

const maxBytes = 30000

func main() {
	udpAddr, _ := net.ResolveUDPAddr("udp4", ":22333") // 转换地址，作为服务器使用时需要监听本机的一个端口

	conn, _ := net.ListenUDP("udp", udpAddr) // 启动UDP监听本机端口\
	conn.SetReadBuffer(maxBytes)
	go func() {
		for {
			//fmt.Println("等待消息中: ")
			var buf [maxBytes]byte
			_, addr, _ := conn.ReadFromUDP(buf[:]) // 读取数据，返回值依次为读取数据长度、远端地址、错误信息 // 读取操作会阻塞直至有数据可读取
			//checkError(err)
			addrString := addr.String()
			//fmt.Println("接到消息了")
			if _, ok := addrMap[addrString]; !ok {
				if addrString == "<nil>" {
					continue
				}
				channel := make(chan []byte, maxBytes)
				addrMap[addrString] = struct {
					addr    *net.UDPAddr
					channel *chan []byte
				}{addr: addr, channel: &channel}
				go sendTask(conn, addr)
			}
			//fmt.Println(fmt.Sprintf("接收了大小为%d字节的数据", length))

			s := buf[:]
			var name [32]byte
			qqq := []byte(addr.String())
			//fmt.Println("名字的长度为", len(qqq))
			copy(name[:], qqq)
			var lis [][]byte
			lis = append(lis, name[:])
			lis = append(lis, s)
			q := bytes.Join(lis, []byte(""))
			_ = sendEveryOne(q, addrString)
			//checkError(err)
		}
	}()
	var done chan string
	<-done
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error %s", err.Error())
		os.Exit(1)
	}
}

func sendEveryOne(bytes []byte, from string) error {
	//fmt.Println(addrMap)
	for i, _ := range addrMap {
		if i == from {
			continue
		}
		//fmt.Println("发送给:", i)
		//fmt.Println(len(addrMap[i].channel))
		// 加锁 防止object被删除 无法.channel
		fmt.Println(addrMap[i].channel)
		*addrMap[i].channel <- bytes[:]
	}
	return nil
}

func sendTask(conn *net.UDPConn, addr *net.UDPAddr) {
	for {
		fmt.Println("监听中")
		channel := *addrMap[addr.String()].channel
		fmt.Println(addrMap[addr.String()].channel)
		select {
		case buf := <-channel:
			_, err := conn.WriteToUDP(buf, addr)
			if err != nil {
				fmt.Println("发送失败:", err.Error())
				delete(addrMap, addr.String())
				return
			}
			fmt.Println("发送给", addr)
		}
	}
}
