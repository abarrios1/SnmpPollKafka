package main`

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/soniah/gosnmp"
)

func main() {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "10.44.64.130:2181"})
	if err != nil {
		panic(err)
	}

	//Where the snmp will be set	

	defer p.Close()

	//Delivery report handler for produced messages

	go func() {

		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to $v\n", ev.TopicPartition)
				}
			}
		}
	}()

	//Produce messages to topic asynchronously
	// Change so a person input topic by command or argument
	topic := "testlog"
	for _, word := range []string{"testing sending"} {
		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:		[]byte(word),
		}, nil)
	}

	//Wait for message deliveries
	p.Flush(15 * 1000)
}
