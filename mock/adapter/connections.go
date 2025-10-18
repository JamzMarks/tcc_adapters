package adapter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/streadway/amqp"
)

// Device representa um dispositivo retornado pelo sistema principal.
type Device struct {
	ID string `json:"id"`
}

// FetchDevices busca devices pela URL
func FetchDevices(url string) ([]Device, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad status %d: %s", resp.StatusCode, string(b))
	}

	var devices []Device
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return nil, err
	}

	for i, d := range devices {
		if d.ID == "" {
			return nil, fmt.Errorf("device at index %d missing id", i)
		}
	}
	return devices, nil
}

// ConnectRabbit estabelece a conexão com o RabbitMQ e retorna o canal ativo.
func ConnectRabbit(rabbitURL string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	return conn, ch, nil
}

// PublishWithRetry publica uma mensagem no RabbitMQ com tentativas de retry.
func PublishWithRetry(ch *amqp.Channel, queue string, body []byte, retries int) error {
	for i := 0; i <= retries; i++ {
		err := ch.Publish(
			"", // exchange padrão
			queue,
			false,
			false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "application/json",
				Body:         body,
			},
		)
		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(200*(i+1)) * time.Millisecond)
	}
	return errors.New("publish retries exhausted")
}

// UpdateMessage representa o payload publicado no RabbitMQ.
type UpdateMessage struct {
	DeviceID string  `json:"device_id"`
	Value    float64 `json:"value"`
	TS       string  `json:"ts"`
}
