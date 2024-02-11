package models

import (
	"github.com/hybridgroup/mjpeg"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
	"gocv.io/x/gocv"
	"io"
	"log"
	"os/exec"
	"strconv"
	"time"
)

const (
	DefaultSpeed = 100
	FrameX       = 960
	FrameY       = 720
	FrameCenterX = FrameX / 2
	FrameCenterY = FrameY / 2
	FrameArea    = FrameX * FrameY
	FrameSize    = FrameArea * 3
)

type DroneManager struct {
	*tello.Driver

	Speed     int
	ffmpegIn  io.WriteCloser
	ffmpegOut io.ReadCloser
	Stream    *mjpeg.Stream
}

func NewDroneManager() *DroneManager {
	droneDriver := tello.NewDriver("8889")

	// ffmpeg command
	ffmpeg := exec.Command("ffmpeg", "-hwaccel", "auto", "-hwaccel_device", "opencl", "-i", "pipe:0",
		"-pix_fmt", "bgr24", "-s", strconv.Itoa(FrameX)+"x"+strconv.Itoa(FrameY), "-f", "rawvideo", "pipe:1")

	ffmpegIn, err := ffmpeg.StdinPipe()
	ffmpegOut, err := ffmpeg.StdoutPipe()
	if err != nil {
		log.Println(err)
	}

	droneManager := &DroneManager{
		Driver:    droneDriver,
		Speed:     DefaultSpeed,
		ffmpegIn:  ffmpegIn,
		ffmpegOut: ffmpegOut,
	}

	work := func() {
		if err = ffmpeg.Start(); err != nil {
			log.Println(err)
			return
		}

		// when connected
		_ = droneDriver.On(tello.ConnectedEvent, func(data interface{}) {
			log.Println("connected")
			_ = droneDriver.StartVideo()
			_ = droneDriver.SetVideoEncoderRate(tello.VideoBitRateAuto)
			_ = droneDriver.SetExposure(0)

			gobot.Every(100*time.Microsecond, func() {
				_ = droneDriver.StartVideo()
			})

			droneManager.StreamVideo()
		})

		// when video stream
		_ = droneDriver.On(tello.VideoFrameEvent, func(data interface{}) {
			packet := data.([]byte)
			if _, err = ffmpegIn.Write(packet); err != nil {
				log.Println(err)
			}
		})
	}

	robot := gobot.NewRobot("tello", []gobot.Connection{}, []gobot.Device{droneDriver}, work)
	_ = robot.Start()

	return droneManager
}

func (dm *DroneManager) StreamVideo() {
	go func(dm *DroneManager) {
		for {
			buf := make([]byte, 256)
			_, err := io.ReadFull(dm.ffmpegOut, buf)
			if err != nil {
				log.Println(err)
			}

			// to img
			img, _ := gocv.NewMatFromBytes(FrameX, FrameY, gocv.MatTypeCV8UC3, buf)
			if img.Empty() {
				continue
			}

			jpegBuf, _ := gocv.IMEncode(".jpg", img)
			dm.Stream.UpdateJPEG(jpegBuf)
		}
	}(dm)
}
