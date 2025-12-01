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

	"mock/adapter"
)

func main() {
	cfg := adapter.LoadConfig()

	// ---- RabbitMQ ----
	conn, ch, err := adapter.ConnectRabbit(cfg.RabbitURL)
	if err != nil {
		log.Fatalf("failed to connect RabbitMQ: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	_, err = ch.QueueDeclare(cfg.QueueName, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("queue declare: %v", err)
	}

	// ---- Carrega dispositivos ----
	devices, err := adapter.FetchDevices(cfg.DeviceAPI)
	if err != nil {
		log.Fatalf("fetch devices: %v", err)
	}
	if len(devices) == 0 {
		log.Println("warning: no devices returned")
	}

	// Mapa dos últimos valores
	last := make(map[string]*float64)
	for _, d := range devices {
		last[d.DeviceID] = nil
	}

	// ---- Gerador randômico único (CORRETO) ----
	rg := adapter.NewRandomGenerator(cfg.Seed)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ticker := time.NewTicker(time.Duration(cfg.PollMs) * time.Millisecond)
	defer ticker.Stop()

	log.Printf("Publicando para %d dispositivos", len(last))

	var wg sync.WaitGroup

loop:
	for {
		select {
		case <-ctx.Done():
			log.Println("stopping adapter…")
			break loop

		case <-ticker.C:
			for deviceID := range last {
				wg.Add(1)
				go func(id string) {
					defer wg.Done()

					// Gera valor novo usando o gerador REAL
					newVal := rg.ComputeNewValue(last[id], cfg.DeltaRange)
					last[id] = &newVal

					// Mensagem
					msg := adapter.EdgeMessage{
						DeviceId:   id,
						DeviceType: "mock",
						Data: struct {
							Confiability float64 `json:"confiability"`
							Flow         float64 `json:"flow"`
						}{
							Confiability: 0.9,
							Flow:         newVal,
						},

						TS: time.Now().UTC().Format(time.RFC3339),
					}

					log.Printf("Publicando mensagem: %+v", msg)

					b, _ := json.Marshal(msg)

					if err := adapter.PublishWithRetry(ch, cfg.QueueName, b, 3); err != nil {
						log.Printf("publish failed for %s: %v", id, err)
					}

				}(deviceID)
			}

			wg.Wait()
		}
	}

	wg.Wait()
	log.Println("adapter stopped")
}
