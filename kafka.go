package main

import (
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	kafkaclients "github.com/superbet-group/kafka.clients/v3"
)

func newKafkaConsumer(
	topic string,
	index int32,
	count int32,
	debugConfig string,
) *kafkaclients.Consumer {
	brokers := "localhost:29092"
	topics := []string{topic}
	configMap := make(map[string]interface{})
	configMap["auto.offset.reset"] = "latest"
	consumerConfig := kafkaclients.ConsumerConfig{
		CloseDelay: time.Duration(1000) * time.Millisecond,
	}

	consumerGroupName := "testGroup"
	var kafkaConsumer *kafkaclients.Consumer
	err := backoff.Retry(
		func() error {
			c, err := kafkaclients.NewConsumerWithDynamicPartitions(brokers, consumerGroupName, topics, index, count, configMap, consumerConfig)
			if err != nil {
				return backoff.Permanent(err)
			}

			// if an error occurs, exponential backoff is used for retrying
			err = prepareConsumer(c, brokers, topic)
			if err != nil {
				fmt.Println(err)
				return err
			}

			kafkaConsumer = c
			return nil
		},
		backoff.NewExponentialBackOff(),
	)
	if err != nil {
		panic(err)
	}

	return kafkaConsumer
}

func prepareConsumer(
	consumer *kafkaclients.Consumer,
	brokers string,
	topic string,
) error {
	// test connection to Kafka
	err := consumer.TestConnection(nil)
	if err != nil {
		return err
	}

	err = consumer.TestTopicsExist(nil)
	if err != nil {
		return err
	}

	return nil
}

func newKafkaProducer(
	debugConfig string,
) *kafkaclients.Producer {
	configMap := make(map[string]interface{})
	producerConfig := kafkaclients.ProducerConfig{
		UseCustomPartitioning: false,
	}

	var kafkaProducer *kafkaclients.Producer
	err := backoff.Retry(
		func() error {
			p, err := kafkaclients.NewProducer("localhost:29092", configMap, producerConfig)
			if err != nil {
				return backoff.Permanent(err)
			}

			// if an error occurs, exponential backoff is used for retrying
			err = prepareProducer(p, "localhost:29092")
			if err != nil {
				return err
			}

			kafkaProducer = p
			return nil
		},
		backoff.NewExponentialBackOff(),
	)
	if err != nil {
		panic(err)
	}

	return kafkaProducer
}

func prepareProducer(
	producer *kafkaclients.Producer,
	brokers string,
) error {
	// test connection to Kafka
	err := producer.TestConnection()
	if err != nil {
		return err
	}

	err = producer.Start()
	if err != nil {
		return err
	}

	return nil
}
