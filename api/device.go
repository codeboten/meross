package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Channel struct {
	Type string `json:"type"`
	Name string `json:"devName"`
}

// Device contains the information specific to a particular
// Meross device.
type Device struct {
	Name       string    `json:"devName"`
	Channels   []Channel `json:"channels"`
	UUID       string    `json:"uuid"`
	Type       string    `json:"deviceType"`
	Domain     string    `json:"domain"`
	Region     string    `json:"region"`
	userID     string
	mqttClient mqtt.Client
}

func (d *Device) getPassword(key string) string {
	return getMD5Hash(fmt.Sprintf("%s%s", d.userID, key))
}

func getRequestTopic(id string) string {
	return fmt.Sprintf("/appliance/%s/subscribe", id)
}

func (d *Device) getResponseTopic() string {
	return fmt.Sprintf("/app/%s-%s/subscribe", d.userID, getMD5Hash(fmt.Sprintf("API%s", d.UUID)))
}

func (d *Device) getClientID() string {
	return fmt.Sprintf("app:%s", getMD5Hash(fmt.Sprintf("API%s", d.UUID)))
}

func (d *Device) connectHandler(client mqtt.Client) {
	fmt.Printf("GOT ONCONNECT:\n")
	// self._client_response_topic = "/app/%s-%s/subscribe" % (self._user_id, self._app_id)
	// self._user_topic = "/app/%s/subscribe" % self._user_id
	if token := client.Subscribe(fmt.Sprintf("/app/%s/subscribe", d.userID), 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
}

func (d *Device) messageHandler(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func (d *Device) getBrokerURL() string {
	domain := d.Domain
	if len(domain) == 0 {
		domain = iotDomain
	}
	return fmt.Sprintf("%s://%s:%d", iotProtocol, domain, iotPort)
}

// Connect establishes a connection using MQTT http://mqtt.org over TCP
// sockets to allow the user to send signals to the device.
func (d *Device) Connect(key string) error {
	mqtt.DEBUG = log.New(os.Stdout, "", 0)
	mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker(d.getBrokerURL()).SetClientID(d.getClientID()).SetCleanSession(true)

	opts.SetProtocolVersion(4)
	opts.SetUsername(d.userID)
	opts.SetPassword(d.getPassword(key))
	opts.SetKeepAlive(30 * time.Second)
	opts.SetDefaultPublishHandler(d.messageHandler)

	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	opts.SetTLSConfig(tlsConfig)
	opts.SetOnConnectHandler(d.connectHandler)

	d.mqttClient = mqtt.NewClient(opts)
	if token := d.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return nil
}

// Disconnect closes the connection to the upstream server
func (d *Device) Disconnect() {
	d.mqttClient.Disconnect(250)
}

func (d *Device) encodePayload(key string, payload mqttPayload) string {
	randomstring := getNonce(16)
	messageID := getMD5Hash(randomstring)
	timestamp := time.Now().Second()
	signature := getMD5Hash(fmt.Sprintf("%s%s%d", messageID, key, timestamp))
	data := mqttMessage{
		Header: mqttHeader{
			From:           d.getResponseTopic(),
			MessageID:      messageID,
			Method:         "SET",
			Namespace:      "Appliance.Control.ToggleX",
			PayloadVersion: 1,
			Signature:      signature,
			Timestamp:      timestamp,
		},
		Payload: payload,
	}
	encodedJSON, _ := json.Marshal(data)
	return string(encodedJSON)
}

// TurnOn sends a signal to turn on a device
func (d *Device) TurnOn(key string) error {
	token := d.mqttClient.Publish(getRequestTopic(d.UUID), 0, false, d.encodePayload(key, mqttPayload{
		map[string]int{
			"onoff": 1,
		},
	}))
	token.Wait()
	return nil
}

// TurnOff senss a signal to turn off a device
func (d *Device) TurnOff(key string) error {
	token := d.mqttClient.Publish(getRequestTopic(d.UUID), 0, false, d.encodePayload(key, mqttPayload{
		map[string]int{
			"onoff": 0,
		},
	}))
	token.Wait()
	return nil
}
