package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/MrDweller/orchestrator-connection/models"
	sensorretrievalsystem "github.com/MrDweller/sensor-retrieval-system/sensor-retrieval-system"
)

type Cli struct {
	sensorRetrievalSystem sensorretrievalsystem.SensorRetrievalSystem
	running               bool
}

type TemperatureResponse struct {
	Temperature float64 `json:"temperature"`
}

func StartCli(sensorRetrievalSystem sensorretrievalsystem.SensorRetrievalSystem) {

	var output io.Writer = os.Stdout
	var input *os.File = os.Stdin

	fmt.Fprintln(output, "Starting sensor retrieval system cli...")

	cli := Cli{
		sensorRetrievalSystem: sensorRetrievalSystem,
		running:               true,
	}

	for {
		if cli.running == false {
			fmt.Fprintln(output, "Stopping the sensor retrieval system!")

			err := sensorRetrievalSystem.StopSensorRetrievalSystem()
			if err != nil {
				log.Panic(err)
			}
			break
		}

		fmt.Fprint(output, "enter command: ")

		reader := bufio.NewReader(input)
		input, _ := reader.ReadString('\n')

		commands := strings.Fields(input)
		cli.handleCommand(output, commands)
	}
}

func (cli *Cli) Stop() {
	cli.running = false
}

func (cli *Cli) handleCommand(output io.Writer, commands []string) {
	numArgs := len(commands)
	if numArgs <= 0 {
		fmt.Fprintln(output, errors.New("no command found"))
		return
	}

	command := strings.ToLower(commands[0])

	switch command {
	case "temp":
		data, err := cli.sensorRetrievalSystem.GetSensorData(models.ServiceDefinition{
			ServiceDefinition: "temperature",
		})
		if err != nil {
			fmt.Fprintln(output, err)
			return
		}

		var temperatureResponse TemperatureResponse
		err = json.Unmarshal(data, &temperatureResponse)
		if err != nil {
			fmt.Fprintln(output, err)
			return
		}

		fmt.Fprintf(output, "Temperature: %f\n", temperatureResponse.Temperature)

	case "help":
		fmt.Fprintln(output, helpText)

	case "clear":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()

	case "exit":
		cli.Stop()

	default:
		fmt.Fprintln(output, errors.New("no command found"))
	}

}

var helpText = `
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	[ SENSOR RETRIEVAL APPLICATION SYSTEM COMMAND LINE INTERFACE ]
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

COMMANDS:
	command [command options] [args...]

VERSION:
	v1.0
	
COMMANDS:
	temp			Get temperature sensor data
	help			Output this help prompt
	clear			Clear the terminal
	exit			Stop the sensor retrieval system
`
