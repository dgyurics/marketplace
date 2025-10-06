package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/dgyurics/marketplace/utilities"
)

func main() {
	command := flag.String("cmd", "", "Command to execute: generate-id or decode-id")
	n := flag.Int("n", 1, "Number of IDs to generate")
	flag.Parse()

	switch *command {
	case "generate-id":
		generateIDs(*n)
	case "decode-id":
		if flag.NArg() < 1 {
			fmt.Println("You must provide an ID to decode")
			return
		}
		decodeID(flag.Arg(0))
	default:
		fmt.Println("Invalid command. Use -cmd=generate-id or -cmd=decode-id")
	}
}

func generateIDs(count int) {
	if err := initializeIDGeneratorForCLI(); err != nil {
		log.Fatalf("Error initializing ID generator: %v", err)
	}
	for i := 0; i < count; i++ {
		id, _ := utilities.GenerateID()
		fmt.Println(id)
	}
}

func decodeID(idStr string) {
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid ID format: %v", err)
	}
	timestamp, machineID, seqID := utilities.DecodeID(id)
	fmt.Printf("Timestamp: %v\nMachine ID: %d\nSequence ID: %d\n", timestamp, machineID, seqID)
}

func initializeIDGeneratorForCLI() error {
	utilities.InitIDGenerator(255)
	return nil
}
