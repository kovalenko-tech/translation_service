package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

// Service represents service for working with RabbitMQ
type Service struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

// TranslationTask represents translation task
type TranslationTask struct {
	RequestID  uuid.UUID         `json:"request_id"`
	SourceData map[string]string `json:"source_data"`
	Languages  []string          `json:"languages"`
}

// NewService creates a new RabbitMQ service instance
func NewService(url, queueName string) (*Service, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare queue
	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &Service{
		conn:    conn,
		channel: ch,
		queue:   queueName,
	}, nil
}

// PublishTask publishes task to queue
func (s *Service) PublishTask(ctx context.Context, task *TranslationTask) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	err = s.channel.Publish(
		"",      // exchange
		s.queue, // routing key
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish task: %w", err)
	}

	log.Printf("Published translation task for request ID: %s", task.RequestID)
	return nil
}

// ConsumeTasks starts consuming tasks from queue
func (s *Service) ConsumeTasks(ctx context.Context, handler func(*TranslationTask) error) error {
	msgs, err := s.channel.Consume(
		s.queue, // queue
		"",      // consumer
		false,   // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgs:
				var task TranslationTask
				if err := json.Unmarshal(msg.Body, &task); err != nil {
					log.Printf("Failed to unmarshal task: %v", err)
					msg.Ack(false)
					continue
				}

				log.Printf("Processing translation task for request ID: %s", task.RequestID)

				if err := handler(&task); err != nil {
					log.Printf("Failed to process task: %v", err)
					// Send message to error queue or log
					msg.Nack(false, true) // requeue
				} else {
					msg.Ack(false)
					log.Printf("Successfully processed translation task for request ID: %s", task.RequestID)
				}
			}
		}
	}()

	return nil
}

// Close closes connection to RabbitMQ
func (s *Service) Close() error {
	if s.channel != nil {
		if err := s.channel.Close(); err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}
	}

	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}

	return nil
}

// GetQueueInfo returns queue information
func (s *Service) GetQueueInfo() (int, error) {
	q, err := s.channel.QueueInspect(s.queue)
	if err != nil {
		return 0, fmt.Errorf("failed to inspect queue: %w", err)
	}

	return q.Messages, nil
}
