package main

import (
    "fmt"
    "log"
    "time"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/iot"
    "github.com/eclipse/paho.mqtt.golang"
    "./tlsconfig"
)

const (
    ThingName    = "home-raspberry"
    SubTopic     = "topic/to/subscribe"
    PubTopic     = "topic/to/publish"
    PubMsg       = `{"message": "こんにちは"}`
    QoS          = 1
)

func handleMsg(_ mqtt.Client, msg mqtt.Message) {
    fmt.Println(msg)
}

func main() {
    // Get AWS Iot Core Endpoint
    s := session.Must(session.NewSession(&aws.Config{
        Region: aws.String("ap-northeast-1"),
    }))
    endpoint, err := iot.New(s).DescribeEndpoint(&iot.DescribeEndpointInput{})
    if err != nil {
        panic(fmt.Sprintf("failed to discribe AWS IoT endpoint: %v", err))
    }
    log.Println("iot endpoint:", *endpoint.EndpointAddress)

    // Connect Broker
    tlsConfig, err := tlsconfig.NewTLSConfig()
    if err != nil {
        panic(fmt.Sprintf("failed to construct tls config: %v", err))
    }
    opts := mqtt.NewClientOptions()
    opts.AddBroker(fmt.Sprintf("ssl://%s:%d", *endpoint.EndpointAddress, 443))
    opts.SetTLSConfig(tlsConfig)
    opts.SetClientID(ThingName)
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        panic(fmt.Sprintf("failed to connect broker: %v", token.Error()))
    }
    defer client.Disconnect(250)

    // Subscribe
    log.Printf("subscribing %s...\n", SubTopic)
    if token := client.Subscribe(SubTopic, QoS, handleMsg); token.Wait() && token.Error() != nil {
        panic(fmt.Sprintf("failed to subscribe %s: %v", SubTopic, token.Error()))
    }

    for {
        // Publish
        log.Printf("publishing %s...\n", PubTopic)
        if token := client.Publish(PubTopic, QoS, false, PubMsg); token.Wait() && token.Error() != nil {
            panic(fmt.Sprintf("failed to publish %s: %v", PubTopic, token.Error()))
        }
        time.Sleep(10 * time.Second)
    }
}
