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
	ID         int     `json:"id"`
	MacAddress *string `json:"macAddress,omitempty"`
	IP         *string `json:"ip,omitempty"`
	DeviceKey  *string `json:"deviceKey,omitempty"`
	DeviceID   string  `json:"deviceId"`
	IsActive   *bool   `json:"isActive,omitempty"`
	CreatedAt  *string `json:"createdAt,omitempty"`
	UpdatedAt  *string `json:"updatedAt,omitempty"`
}
type DevicesResponse struct {
	Success bool     `json:"success"`
	Data    []Device `json:"data"`
	Message string   `json:"message"`
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

	var res DevicesResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return res.Data, nil
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

type EdgeMessage struct {
	DeviceId   string `json:"deviceId"`
	DeviceType string `json:"deviceType"`
	Data       struct {
		Confiability float64 `json:"confiability"`
		Flow         float64 `json:"flow"`
	} `json:"data"`
	TS string `json:"ts"`
}
