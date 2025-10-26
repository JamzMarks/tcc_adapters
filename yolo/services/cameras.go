package services

import (
	"fmt"
	"image"
	"log"
	"time"

	"gocv.io/x/gocv"
)

// FetchFrames pega n frames do stream de um device
func FetchFrames(ip string, n int, delay time.Duration) ([]image.Mat, error) {
	url := fmt.Sprintf("http://%s:4747/video", ip) // DroidCam padrão HTTP
	webcam, err := gocv.VideoCaptureFile(url)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir stream: %w", err)
	}
	defer webcam.Close()

	frames := make([]image.Mat, 0, n)
	mat := gocv.NewMat()
	defer mat.Close()

	for i := 0; i < n; i++ {
		if ok := webcam.Read(&mat); !ok {
			log.Printf("falha ao ler frame %d do %s", i, ip)
			continue
		}
		if mat.Empty() {
			log.Printf("frame vazio %d do %s", i, ip)
			continue
		}
		frames = append(frames, mat.Clone())
		time.Sleep(delay)
	}
	return frames, nil
}
