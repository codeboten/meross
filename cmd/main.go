package main

import (
	"fmt"
	"os"
	"time"

	"github.com/codeboten/meross/api"
)

func main() {
	client := api.NewClient(os.Getenv("MEROSS_EMAIL"), os.Getenv("MEROSS_PASSWORD"))

	err := client.Login()
	if err != nil {
		fmt.Printf("Login error: %v\n", err)
	}
	devices, err := client.GetSupportedDevices()
	if err != nil {
		fmt.Printf("GetSupportedDevices error: %v\n", err)
	}

	if len(devices) == 0 {
		fmt.Printf("No devices found")
		return
	}

	for _, device := range devices {
		fmt.Printf("Setting up MQTT channel for %s", device.Name)
		device.Connect(client.Key)
		fmt.Printf("Turning off: %s\n", device.Name)
		device.TurnOff(client.Key)
		time.Sleep(5 * time.Second)
		fmt.Printf("Turning on: %s\n", device.Name)
		device.TurnOn(client.Key)
		time.Sleep(5 * time.Second)
		device.Disconnect()
	}
}
