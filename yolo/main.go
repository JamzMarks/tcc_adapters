package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"yolo/services"
)

func main() {
	cfg := services.LoadConfig()
	conn, ch, err := services.ConnectRabbit(cfg.RabbitURL)
	if err != nil {
		log.Fatalf("failed to connect RabbitMQ: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	_, err = ch.QueueDeclare(cfg.QueueName, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("queue declare: %v", err)
	}

	devices, err := services.FetchDevices(cfg.DeviceAPI)
	if err != nil {
		log.Fatalf("fetch devices: %v", err)
	}
	if len(devices) == 0 {
		log.Println("warning: no devices returned")
	}

	last := make(map[string]*float64)
	for _, d := range devices {
		last[d.DeviceID] = nil
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ticker := time.NewTicker(time.Duration(cfg.PollMs) * time.Millisecond)
	log.Printf("Publicando para %d dispositivos", len(last))
	defer ticker.Stop()

	var wg sync.WaitGroup

	frames, err := services.FetchFrames(d.IP, 5, 500*time.Millisecond)
	if err != nil {
		log.Printf("erro capturando frames de %s: %v", d.DeviceID, err)
		continue
	}

	for _, frame := range frames {
		// converter frame para bytes para enviar ao YOLO
		buf, _ := gocv.IMEncode(".jpg", frame)
		go func(f []byte, deviceID string) {
			results, err := services.SendToYOLO(f)
			if err != nil {
				log.Printf("erro enviando para YOLO: %v", err)
			} else {
				log.Printf("detecção %s: %s", deviceID, results)
			}
		}(buf, d.DeviceID)
	}

	wg.Wait()
	log.Println("adapter stopped")
}
