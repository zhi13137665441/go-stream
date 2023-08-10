package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gocv.io/x/gocv"
)

func main() {
	camera, err := gocv.VideoCaptureDevice(0)
	if err != nil {
		panic(err)
	}
	defer camera.Close()
	// 创建 mjpeg 流
	//stream := mjpeg.NewStream()
	// 启动一个协程以处理视频流
	a := app.New()
	w := a.NewWindow("Hello")

	hello := widget.NewLabel("my video")
	card := widget.Card{}
	//card.Resize(fyne.NewSize(1000, 1000))
	w.SetContent(container.NewVBox(
		hello,
		widget.NewCard("", "", &card),
	))
	go func() {
		for {

			// 读取一帧图像
			img := gocv.NewMat()
			if ok := camera.Read(&img); !ok {
				fmt.Println("err occurred:", err)
			}
			// 在图像上添加一些注解（如文本、线条、图形等）
			// TODO: 在图像上添加注解的代码
			// 将图像转换为 JPEG 格式
			buf, _ := gocv.IMEncode(".jpg", img)
			// 将 JPEG 数据写入 mjpeg 流
			//stream.UpdateJPEG(buf.GetBytes())
			image := fyne.NewStaticResource("ddddd", buf.GetBytes())
			card.SetImage(canvas.NewImageFromResource(image))
			card.Image.Resize(fyne.NewSize(1000, 1000))
			//card.Image.Refresh()
		}
	}()
	w.ShowAndRun()
}
