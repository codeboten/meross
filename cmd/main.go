package main

import (
	"fmt"
	"os"

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

	fmt.Printf("Found devices: %v\n", devices)
	devices[0].TurnOn()
	devices[0].TurnOff()
}
