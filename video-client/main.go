package main

import (
	"bytes"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
	"gocv.io/x/gocv"
	"image"
	"image/jpeg"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
)

var DataMap map[string]struct {
	addr string
	card *widget.Card
}
var box *fyne.Container
var window fyne.Window

const maxBytes = 30000

func init() {

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	DataMap = map[string]struct {
		addr string
		card *widget.Card
	}{}
	DataMap["myself"] = struct {
		addr string
		card *widget.Card
	}{addr: "myself", card: &widget.Card{Title: "myself"}}

	a := app.New()
	window = a.NewWindow("对话框")

	hello := widget.NewLabel("my video")
	DataMap["myself"].card.Resize(fyne.NewSize(1000, 1000))
	box = container.NewVBox(
		hello,
		DataMap["myself"].card,
	)
	window.SetContent(box)
}

func main() {
	camera, err := gocv.VideoCaptureDevice(0)
	if err != nil {
		panic(err)
	}
	defer camera.Close()
	udpAddr, err := net.ResolveUDPAddr("udp4", "192.168.2.161:22333") // 转换地址，作为客户端使用要向远程发送消息，这里用远程地址与端口号
	checkError(err)
	conn, err := net.DialUDP("udp", nil, udpAddr) // 建立连接，第二个参数为nil时通过默认本地地址（猜测可能是第一个可用的地址，未进行测试）发送且端口号自动分配，第三个参数为远程端地址与端口号
	conn.SetReadBuffer(maxBytes + 32)
	checkError(err)
	go receive(conn) // 使用DialUDP建立连接后也可以监听来自远程端的数据
	go Send(conn, camera)
	window.ShowAndRun()
}

func receive(conn *net.UDPConn) {
	for {
		var buf [maxBytes + 32]byte
		read, err := conn.Read(buf[:]) // 读取数据 // 读取操作会阻塞直至有数据可读取
		if read < 32 {
			fmt.Println("数据格式不匹配")
			continue
		}
		checkError(err)
		addr := buf[:32]
		//fmt.Println("地址为: ", addr)
		if _, ok := DataMap[string(addr)]; !ok {
			card := &widget.Card{Title: string(addr)}
			DataMap[string(addr)] = struct {
				addr string
				card *widget.Card
			}{addr: string(addr), card: card}
			card.Resize(fyne.NewSize(1000, 1000))
			box.Add(card)
			fmt.Println("增加了一个card")
			window.Canvas().Refresh(box)
		}
		data := buf[32:]
		setCardImage(DataMap[string(addr)].card, data)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error %s", err.Error())
		os.Exit(1)
	}
}

func getCurrentCamara(camera *gocv.VideoCapture) []byte {
	time.Sleep(time.Millisecond * 20)
	// 读取一帧图像
	img := gocv.NewMat()
	if ok := camera.Read(&img); !ok {
		return []byte("")
	}
	// 在图像上添加一些注解（如文本、线条、图形等）
	// TODO: 在图像上添加注解的代码
	// 将图像转换为 JPEG 格式
	buf, _ := gocv.IMEncode(".jpeg", img)
	img.Close()
	bufbytes := buf.GetBytes()
	buf.Close()
	bytesRes := compressImageResource(bufbytes)
	setCardImage(DataMap["myself"].card, bytesRes)
	fmt.Println(len(bytesRes))
	if len(bytesRes) > maxBytes {
		return []byte{}
	}
	return bytesRes
}

func Send(conn *net.UDPConn, camera *gocv.VideoCapture) {
	for {
		_, err := conn.Write(getCurrentCamara(camera))
		if err != nil {
			fmt.Println("error occurred:", err)
		}
	}
}

func setCardImage(card *widget.Card, bytes []byte) {

	// 内存泄漏项
	//card.SetImage(canvas.NewImageFromResource(fyne.NewStaticResource("", bytes)))
	/*
		//card.Image = &canvas.Image{
		//	Resource: &fyne.StaticResource{StaticContent: bytes}}
		//card.Refresh()
	*/
	var resource *fyne.StaticResource
	resource = &fyne.StaticResource{
		StaticName:    "current",
		StaticContent: bytes,
	}
	//img := canvas.Image{Resource: resource}
	if card.Image != nil {
		card.Image.Resource = resource
	} else {
		img := canvas.Image{Resource: resource}
		card.SetImage(&img)
	}
	card.Refresh()
}

func compressImageResource(data []byte) []byte {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data
	}
	var buf bytes.Buffer
	defer buf.Reset()
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 50})
	if err != nil {
		return data
	}
	if buf.Len() > len(data) {
		return data
	}
	return buf.Bytes()
}
