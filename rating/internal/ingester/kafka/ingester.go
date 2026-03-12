package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"

	model "movieexample.com/rating/pkg"
)

// Ingester defines a Kafka ingester.
type Ingester struct {
	consumer sarama.Consumer
	topic    string
}

// NewIngester creates a new Kafka ingester.
func NewIngester(addr string, topic string) (*Ingester, error) {

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer([]string{addr}, config)
	if err != nil {
		return nil, err
	}

	return &Ingester{consumer: consumer, topic: topic}, nil
}

// Ingest starts ingestion from Kafka and returns a chanel containing
// rating events representing the data consumed from the topic
func (i *Ingester) Ingest(ctx context.Context) (chan model.RatingEvent, error) {

	partitionConsumer, err := i.consumer.ConsumePartition(i.topic, 0, sarama.OffsetNewest)
	if err != nil {
		return nil, err
	}

	ch := make(chan model.RatingEvent, 1)
	go func() {
		defer close(ch)
		defer partitionConsumer.Close()

		for {
			select {
			case msg := <-partitionConsumer.Messages():
				fmt.Println("Processing a message")
				var event model.RatingEvent
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					fmt.Printf("Failed to unmarshal message: %v\n", err)
					continue
				}

				fmt.Printf("Consumed a message: %v\n", event)
				ch <- event

			case err := <-partitionConsumer.Errors():
				fmt.Printf("Consumer error: %v\n", err)
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}
