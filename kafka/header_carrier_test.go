package kafka

import (
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestKafkaHeaderCarrier_Get(t *testing.T) {
	headers := []kafka.Header{
		{Key: "traceparent", Value: []byte("00-abc123-def456-01")},
		{Key: "content-type", Value: []byte("application/json")},
	}
	carrier := &kafkaHeaderCarrier{headers: headers}

	assert.Equal(t, "00-abc123-def456-01", carrier.Get("traceparent"))
	assert.Equal(t, "application/json", carrier.Get("content-type"))
	assert.Equal(t, "", carrier.Get("non-existent"))
}

func TestKafkaHeaderCarrier_Set_NewKey(t *testing.T) {
	headers := []kafka.Header{
		{Key: "existing", Value: []byte("value")},
	}
	carrier := &kafkaHeaderCarrier{headers: headers}

	carrier.Set("new-key", "new-value")

	assert.Len(t, carrier.headers, 2)
	assert.Equal(t, "new-value", carrier.Get("new-key"))
}

func TestKafkaHeaderCarrier_Set_UpdateExisting(t *testing.T) {
	headers := []kafka.Header{
		{Key: "key", Value: []byte("old-value")},
	}
	carrier := &kafkaHeaderCarrier{headers: headers}

	carrier.Set("key", "new-value")

	assert.Len(t, carrier.headers, 1)
	assert.Equal(t, "new-value", carrier.Get("key"))
}

func TestKafkaHeaderCarrier_Keys(t *testing.T) {
	headers := []kafka.Header{
		{Key: "key1", Value: []byte("value1")},
		{Key: "key2", Value: []byte("value2")},
		{Key: "key3", Value: []byte("value3")},
	}
	carrier := &kafkaHeaderCarrier{headers: headers}

	keys := carrier.Keys()

	assert.Equal(t, []string{"key1", "key2", "key3"}, keys)
}

func TestKafkaHeaderCarrier_Keys_Empty(t *testing.T) {
	carrier := &kafkaHeaderCarrier{headers: []kafka.Header{}}

	keys := carrier.Keys()

	assert.Empty(t, keys)
}
