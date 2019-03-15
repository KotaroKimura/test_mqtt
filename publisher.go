package main

import (
    "fmt"
    MQTT "github.com/eclipse/paho.mqtt.golang"
)

func main () {
    const MQTT_BROKER = "test.mosquitto.org:1883"
    opts := MQTT.NewClientOptions().AddBroker(MQTT_BROKER)
    opts.SetClientID("test")
    client := MQTT.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        fmt.Println("Error %s\n", token.Error())
    }

    token := client.Publish("kotaro.kimura", 0, false, "{\"message\":\"hello\"}")
    token.Wait()

    client.Disconnect(250)

}