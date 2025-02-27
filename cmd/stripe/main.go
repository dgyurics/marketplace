package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"os/exec"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// Define CLI commands
	command := flag.String("cmd", "", "Command to execute: listen, update, confirm, refund")
	paymentIntent := flag.String("pi", "", "Payment Intent ID (required for update, confirm, refund)")
	flag.Parse()

	switch *command {
	case "listen":
		startStripeListener()
	case "update":
		execStripeCommand("payment_intents", "update", *paymentIntent, "--payment-method", "pm_card_visa")
	case "confirm":
		execStripeCommand("payment_intents", "confirm", *paymentIntent)
	case "refund":
		execStripeCommand("refunds", "create", "--payment_intent", *paymentIntent)
	default:
		fmt.Println("Invalid command. Use -cmd=listen, update, confirm, or refund")
	}
}

func startStripeListener() {
	secretKey := os.Getenv("STRIPE_SECRET_KEY")
	if secretKey == "" {
		log.Fatal("Error: STRIPE_SECRET_KEY is not set in the .env file")
	}

	cmd := exec.Command("stripe", "listen", "--api-key", secretKey, "--forward-to", "http://localhost:8000/orders/events")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to start Stripe listener: %v", err)
	}
}

func execStripeCommand(args ...string) {
	if args[1] != "listen" && args[2] == "" {
		log.Fatalf("Error: Payment Intent ID is required for %s", args[1])
	}

	cmd := exec.Command("stripe", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to execute Stripe command: %v", err)
	}
}
