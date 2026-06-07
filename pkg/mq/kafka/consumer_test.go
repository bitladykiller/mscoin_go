package kafka

import (
	"context"
	"errors"
	"testing"
	"time"

	rawkafka "github.com/segmentio/kafka-go"
)

type fakeReader struct {
	commitErr error
	committed []rawkafka.Message
}

func (f *fakeReader) FetchMessage(context.Context) (rawkafka.Message, error) {
	return rawkafka.Message{}, nil
}

func (f *fakeReader) CommitMessages(_ context.Context, msgs ...rawkafka.Message) error {
	f.committed = append(f.committed, msgs...)
	return f.commitErr
}

func (f *fakeReader) Close() error {
	return nil
}

type fakeProducer struct {
	pushes []struct {
		key   string
		value string
	}
	pushErr error
}

func (f *fakeProducer) PushWithKey(_ context.Context, key string, value string) error {
	f.pushes = append(f.pushes, struct {
		key   string
		value string
	}{key: key, value: value})
	return f.pushErr
}

func (f *fakeProducer) Close() error {
	return nil
}

func TestNewConsumerServiceValidatesRequiredFields(t *testing.T) {
	t.Parallel()

	validHandler := func(context.Context, Message) error { return nil }

	if _, err := NewConsumerService(ConsumerConfig{
		Topic:   "withdraw",
		GroupID: "jobcenter",
	}, validHandler, nil); err == nil {
		t.Fatal("NewConsumerService() should fail when brokers are missing")
	}
	if _, err := NewConsumerService(ConsumerConfig{
		Brokers: []string{"127.0.0.1:9092"},
		GroupID: "jobcenter",
	}, validHandler, nil); err == nil {
		t.Fatal("NewConsumerService() should fail when topic is missing")
	}
	if _, err := NewConsumerService(ConsumerConfig{
		Brokers: []string{"127.0.0.1:9092"},
		Topic:   "withdraw",
	}, validHandler, nil); err == nil {
		t.Fatal("NewConsumerService() should fail when group id is missing")
	}
	if _, err := NewConsumerService(ConsumerConfig{
		Brokers: []string{"127.0.0.1:9092"},
		Topic:   "withdraw",
		GroupID: "jobcenter",
	}, nil, nil); err == nil {
		t.Fatal("NewConsumerService() should fail when handler is missing")
	}
}

func TestConsumerHandleFetchedMessageCommitsOnAck(t *testing.T) {
	t.Parallel()

	reader := &fakeReader{}
	service := &consumerService{
		cfg: ConsumerConfig{
			Topic:   "withdraw",
			GroupID: "jobcenter",
		},
		reader:        reader,
		handler:       func(context.Context, Message) error { return nil },
		classifier:    defaultClassifier,
		ctx:           context.Background(),
		commitTimeout: time.Second,
	}

	msg := rawkafka.Message{Topic: "withdraw", Offset: 9, Key: []byte("1"), Value: []byte(`{}`)}
	if err := service.handleFetchedMessage(msg); err != nil {
		t.Fatalf("handleFetchedMessage() error = %v", err)
	}
	if len(reader.committed) != 1 || reader.committed[0].Offset != 9 {
		t.Fatalf("handleFetchedMessage() committed = %+v, want offset 9", reader.committed)
	}
}

func TestConsumerHandleFetchedMessageDeadLettersPoisonMessage(t *testing.T) {
	t.Parallel()

	reader := &fakeReader{}
	producer := &fakeProducer{}
	service := &consumerService{
		cfg: ConsumerConfig{
			Topic:           "withdraw",
			GroupID:         "jobcenter",
			DeadLetterTopic: "withdraw.dlq",
		},
		reader:     reader,
		deadLetter: producer,
		handler: func(context.Context, Message) error {
			return errors.New("bad payload")
		},
		classifier: func(err error) ConsumeAction {
			if err == nil {
				return ConsumeAck
			}
			return ConsumeDeadLetter
		},
		ctx:           context.Background(),
		commitTimeout: time.Second,
	}

	msg := rawkafka.Message{Topic: "withdraw", Offset: 7, Key: []byte("2"), Value: []byte(`bad`)}
	if err := service.handleFetchedMessage(msg); err != nil {
		t.Fatalf("handleFetchedMessage() error = %v", err)
	}
	if len(producer.pushes) != 1 {
		t.Fatalf("dead letter pushes = %d, want 1", len(producer.pushes))
	}
	if len(reader.committed) != 1 || reader.committed[0].Offset != 7 {
		t.Fatalf("committed offset = %+v, want 7", reader.committed)
	}
}

func TestConsumerHandleFetchedMessageRetriesThenCommits(t *testing.T) {
	t.Parallel()

	reader := &fakeReader{}
	attempts := 0
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service := &consumerService{
		cfg: ConsumerConfig{
			Topic:          "withdraw",
			GroupID:        "jobcenter",
			RetryBackoffMs: 1,
		},
		reader: reader,
		handler: func(context.Context, Message) error {
			attempts++
			if attempts == 1 {
				return errors.New("temporary")
			}
			return nil
		},
		classifier:    defaultClassifier,
		ctx:           ctx,
		cancel:        cancel,
		commitTimeout: time.Second,
	}

	msg := rawkafka.Message{Topic: "withdraw", Offset: 11}
	if err := service.handleFetchedMessage(msg); err != nil {
		t.Fatalf("handleFetchedMessage() error = %v", err)
	}
	if attempts != 2 {
		t.Fatalf("retry attempts = %d, want 2", attempts)
	}
	if len(reader.committed) != 1 || reader.committed[0].Offset != 11 {
		t.Fatalf("committed offset = %+v, want 11", reader.committed)
	}
}
