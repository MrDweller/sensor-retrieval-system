package main

import (
	"log"
	"os"
	"strconv"

	"github.com/MrDweller/sensor-retrieval-system/cli"
	sensorretrievalsystem "github.com/MrDweller/sensor-retrieval-system/sensor-retrieval-system"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	address := os.Getenv("ADDRESS")
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Panic(err)
	}
	systemName := os.Getenv("SYSTEM_NAME")

	serviceRegistryAddress := os.Getenv("SERVICE_REGISTRY_ADDRESS")
	serviceRegistryPort, err := strconv.Atoi(os.Getenv("SERVICE_REGISTRY_PORT"))
	if err != nil {
		log.Panic(err)
	}

	sensorRetrievalSystem, err := sensorretrievalsystem.NewSensorRetrievalSystem(address, port, systemName, serviceRegistryAddress, serviceRegistryPort)
	if err != nil {
		log.Panic(err)
	}
	sensorRetrievalSystem.StartSensorRetrievalSystem()

	cli.StartCli(*sensorRetrievalSystem)
}
