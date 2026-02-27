package mq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

func NewWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 250 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}
}

func NewReader(brokers []string, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
		MaxWait:        time.Second,
	})
}

func PublishJSON(ctx context.Context, writer *kafka.Writer, key string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: body,
		Time:  time.Now().UTC(),
	})
}

func ParseMessageJSON[T any](msg kafka.Message) (T, error) {
	var payload T
	err := json.Unmarshal(msg.Value, &payload)
	return payload, err
}
