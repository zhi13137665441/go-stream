package main

import (
	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
	"net/http"
)

func main() {
	// 打开相机
	camera, err := gocv.VideoCaptureDevice(0)
	if err != nil {
		panic(err)
	}
	defer camera.Close()
	// 创建 mjpeg 流
	stream := mjpeg.NewStream()
	// 启动一个协程以处理视频流
	go func() {
		for {
			// 读取一帧图像
			img := gocv.NewMat()
			if ok := camera.Read(&img); !ok {
				break
			}
			// 在图像上添加一些注解（如文本、线条、图形等）
			// TODO: 在图像上添加注解的代码
			// 将图像转换为 JPEG 格式
			buf, _ := gocv.IMEncode(".jpg", img)
			// 将 JPEG 数据写入 mjpeg 流
			stream.UpdateJPEG(buf.GetBytes())
		}
	}()
	// 启动 web 服务，将 mjpeg 流返回给前端
	http.Handle("/", stream)
	http.ListenAndServe(":8080", nil)
}
