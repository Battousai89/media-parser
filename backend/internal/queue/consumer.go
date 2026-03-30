package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/media-parser/backend/internal/config"
	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn         *amqp091.Connection
	channel      *amqp091.Channel
	config       *config.RabbitMQConfig
	mu           sync.Mutex
	stopChan     chan struct{}
	isRunning    bool
	exchange     string
	parseChan    chan *ParseTask
	httpChan     chan *HTTPFetchTask
	downloadChan chan *DownloadTask
}

func NewConsumer(cfg *config.RabbitMQConfig) *Consumer {
	return &Consumer{
		config:       cfg,
		exchange:     "media_parser_exchange",
		parseChan:    make(chan *ParseTask, 100),
		httpChan:     make(chan *HTTPFetchTask, 100),
		downloadChan: make(chan *DownloadTask, 100),
	}
}

func (c *Consumer) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := amqp091.Dial(c.config.URL())
	if err != nil {
		return fmt.Errorf("dial rabbitmq: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("open channel: %w", err)
	}

	if err := channel.ExchangeDeclare(
		c.exchange,
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

	if err := channel.Qos(
		c.config.Prefetch,
		0,
		false,
	); err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("set qos: %w", err)
	}

	queue, err := channel.QueueDeclare(
		c.config.Queue,
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
		c.config.Queue,
		c.exchange,
		false,
		nil,
	); err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("bind queue: %w", err)
	}

	c.conn = conn
	c.channel = channel
	return nil
}

func (c *Consumer) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channel == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	msgs, err := c.channel.Consume(
		c.config.Queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	c.stopChan = make(chan struct{})
	c.isRunning = true

	go c.dispatcher(msgs)

	return nil
}

func (c *Consumer) dispatcher(msgs <-chan amqp091.Delivery) {
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				time.Sleep(5 * time.Second)
				_ = c.Connect()
				continue
			}

			var baseTask struct {
				Type      string    `json:"type,omitempty"`
				RequestID uuid.UUID `json:"request_id"`
				SourceID  int       `json:"source_id"`
				URL       string    `json:"url"`
			}
			if err := json.Unmarshal(msg.Body, &baseTask); err == nil && baseTask.RequestID != uuid.Nil {
				task := &ParseTask{
					RequestID:  baseTask.RequestID,
					SourceID:   baseTask.SourceID,
					URL:        baseTask.URL,
					CreatedAt:  time.Now(),
					Limit:      10,
					MaxRetries: 3,
				}
				if err := json.Unmarshal(msg.Body, task); err == nil {
					select {
					case c.parseChan <- task:
						_ = msg.Ack(false)
						continue
					default:
					}
				}
			}

			log.Printf("Consumer: failed to process message, rejecting")
			_ = msg.Nack(false, false)

		case <-c.stopChan:
			return
		}
	}
}

func (c *Consumer) ParseTasks() <-chan *ParseTask {
	return c.parseChan
}

func (c *Consumer) HTTPFetchTasks() <-chan *HTTPFetchTask {
	return c.httpChan
}

func (c *Consumer) DownloadTasks() <-chan *DownloadTask {
	return c.downloadChan
}

func (c *Consumer) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isRunning {
		return
	}

	close(c.stopChan)
	c.isRunning = false

	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *Consumer) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isRunning
}

func (c *Consumer) PublishRetry(task *ParseTask) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channel == nil {
		return fmt.Errorf("not connected")
	}

	body, err := json.Marshal(task)
	if err != nil {
		return err
	}

	msg := amqp091.Publishing{
		DeliveryMode: amqp091.Persistent,
		ContentType:  "application/json",
		Body:         body,
	}

	return c.channel.Publish(
		c.exchange,
		c.config.Queue,
		false,
		false,
		msg,
	)
}
