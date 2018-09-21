package main

import (
	"fmt"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/k-sone/snmpgo"
)

// Struct for oids 
type modem struct {
	SysDescr      string 
	SysObjectID   string 
	SysUpTimeTick string 
}

// TODO: Create separate functions instead all in main
// Main function
func main() {

	// Initialize snmp instance with correct host arguments
	snmp, err := snmpgo.NewSNMP(snmpgo.SNMPArguments{
		Version:   snmpgo.V2c,
		Address:   "10.44.3.2:161",
		Retries:   1,
		Community: "public",
	})

	if err != nil {
		// Failed to create snmpgo.SNMP object
		fmt.Println(err)
		return
	}
	
	// Initialize Producer instance
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "10.44.64.135"})
	if err != nil {
		panic(err)
	}
	// close producer instance after everything has executed and returned
	defer p.Close()

	// Concurrently run func and determine if message was delivered succesfully or failed
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
	
	// Array of OIDs 
	oids, err := snmpgo.NewOids([]string{
		"1.3.6.1.2.1.1.1.0",
		"1.3.6.1.2.1.1.2.0",
		"1.3.6.1.2.1.1.3.0",
	})
	
	// Check err status on Oids
	if err != nil {
		// Failed to parse Oids
		fmt.Println(err)
		return
	}

	// Catches error on establishing connection
	if err = snmp.Open(); err != nil {
		// Failed to open connection
		fmt.Println(err)
		return
	}
	// Defer used to close snmp after everything has been executed and returned
	defer snmp.Close()

	// PDU requests to get the oids specified
	pdu, err := snmp.GetRequest(oids)
	if err != nil {
		// Failed to request
		fmt.Println(err)
		return
	}

	// If pdu has an error print the error
	if pdu.ErrorStatus() != snmpgo.NoError {
		// Received an error from the agent
		fmt.Println(pdu.ErrorStatus(), pdu.ErrorIndex())
	}

	//Produce messages to topic asynchronously
	// TODO: Change so a person input topic by command or argument
	topic := "elk_snmp"

	// Set varbinds to strings
	sysDes := (pdu.VarBinds().MatchOid(oids[0]).String())
	sysObjID := (pdu.VarBinds().MatchOid(oids[1]).String())
	sysUpTim := (pdu.VarBinds().MatchOid(oids[2]).String())

	// Set b to json form of struct
	b, err := json.Marshal(modem {
		SysDescr: sysDes,
		SysObjectID: sysObjID,
		SysUpTimeTick: sysUpTim,
	});
	
	// If there is an error, close program and send error
	if err != nil {
		panic(err)
	}
	// Print the string of the struct
	// fmt.Println(string(b))
	p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(string(b)),
	}, nil)
	
	// Flush the producer for sending data next
	p.Flush(15 * 1000)
}
