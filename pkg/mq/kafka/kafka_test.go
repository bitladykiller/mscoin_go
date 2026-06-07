package kafka

import "testing"

func TestNewProducerValidatesRequiredFields(t *testing.T) {
	t.Parallel()

	if _, err := NewProducer(Config{Topic: "withdraw"}); err == nil {
		t.Fatal("NewProducer() should fail when brokers are missing")
	}
	if _, err := NewProducer(Config{Brokers: []string{"127.0.0.1:9092"}}); err == nil {
		t.Fatal("NewProducer() should fail when topic is missing")
	}
}

func TestProducerCloseHandlesNilWriter(t *testing.T) {
	t.Parallel()

	var raw *producer
	if err := raw.Close(); err != nil {
		t.Fatalf("Close() error = %v, want nil", err)
	}
}
