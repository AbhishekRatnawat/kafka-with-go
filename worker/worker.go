package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
)

func main() {

	topic := "comments"

	worker, err := connectConsumer([]string{"localhost:29092"})

	if err != nil {
		panic(err)
	}

	consumer, err := worker.ConsumePartition(topic, 0, sarama.OffsetOldest)

	if err != nil {
		panic(err)
	}

	fmt.Println("Consumer is ready")

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	msgCount := 0

	doneChan := make(chan struct{})

	go func() {
		for {
			select {
			case err := <-consumer.Errors():
				fmt.Println(err)
			case msg := <-consumer.Messages():
				msgCount++
				fmt.Printf("Received message count: %d: | Topic (%s)\n", msgCount, string(msg.Topic))
			case <-sigchan:
				fmt.Println("Interruption detected")
				doneChan <- struct{}{}
			}
		}
	}()

	<-doneChan
	fmt.Println("Processed", msgCount, "messages")
	if err := worker.Close(); err != nil {
		fmt.Println("Failed to close consumer", err)
	}

}

func connectConsumer(brokersUrl []string) (sarama.Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	worker, err := sarama.NewConsumer(brokersUrl, config)

	if err != nil {
		return nil, err
	}

	return worker, nil
}
