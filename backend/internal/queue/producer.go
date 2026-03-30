package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/media-parser/backend/internal/config"
	"github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	conn     *amqp091.Connection
	channel  *amqp091.Channel
	queue    amqp091.Queue
	exchange string
	config   *config.RabbitMQConfig
	mu       sync.Mutex
}

func NewProducer(cfg *config.RabbitMQConfig) *Producer {
	return &Producer{
		config:   cfg,
		exchange: "media_parser_exchange",
	}
}

func (p *Producer) Connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn, err := amqp091.Dial(p.config.URL())
	if err != nil {
		return fmt.Errorf("dial rabbitmq: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("open channel: %w", err)
	}

	if err := channel.ExchangeDeclare(
		p.exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("declare exchange: %w", err)
	}

	queue, err := channel.QueueDeclare(
		p.config.Queue,
		true,
		false,
		false,
		false,
		amqp091.Table{
			"x-max-priority": 10,
		},
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("declare queue: %w", err)
	}

	if err := channel.QueueBind(
		queue.Name,
		p.config.Queue,
		p.exchange,
		false,
		nil,
	); err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("bind queue: %w", err)
	}

	p.conn = conn
	p.channel = channel
	p.queue = queue

	return nil
}

func (p *Producer) Publish(ctx context.Context, task *ParseTask) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.channel == nil {
		if err := p.Connect(); err != nil {
			return err
		}
	}

	body, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}

	priority := uint8(0)
	if task.Priority > 0 && task.Priority <= 10 {
		priority = uint8(task.Priority)
	}

	msg := amqp091.Publishing{
		DeliveryMode: amqp091.Persistent,
		ContentType:  "application/json",
		Timestamp:    time.Now(),
		Body:         body,
		MessageId:    uuid.New().String(),
		Priority:     priority,
	}

	return p.channel.PublishWithContext(
		ctx,
		p.exchange,
		p.config.Queue,
		false,
		false,
		msg,
	)
}

func (p *Producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			return err
		}
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

func (p *Producer) IsConnected() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.conn != nil && !p.conn.IsClosed()
}
