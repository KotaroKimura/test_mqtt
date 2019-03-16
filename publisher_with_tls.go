package main

import (
    "fmt"
    "io/ioutil"
    "crypto/tls"
    "crypto/x509"
    "log"
    "time"
    "os"

    "github.com/joho/godotenv"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/iot"
    "github.com/eclipse/paho.mqtt.golang"
)

const (
    ThingName    = "home-raspberry"
    SubTopic     = "topic/to/subscribe"
    PubTopic     = "topic/to/publish"
    PubMsg       = `{"message": "こんにちは"}`
    QoS          = 1
)

func newTLSConfig() (*tls.Config, error) {
    // Load Environment Variables
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
    RootCAPath   := os.Getenv("RootCAPath")
    CertFilePath := os.Getenv("CertFilePath")
    KeyFilePath  := os.Getenv("KeyFilePath")

    // Set RootCA
    rootCA, err := ioutil.ReadFile(generateCertPath(RootCAPath))
    if err != nil {
        return nil, err
    }
    pool := x509.NewCertPool()
    pool.AppendCertsFromPEM(rootCA)

    // Set Cert
    cert, err := tls.LoadX509KeyPair(generateCertPath(CertFilePath), generateCertPath(KeyFilePath))
    if err != nil {
        return nil, err
    }
    cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
    if err != nil {
        return nil, err
    }

    return &tls.Config {
        RootCAs:            pool,
        InsecureSkipVerify: true,
        ClientCAs:          nil,
        Certificates:       []tls.Certificate{cert},
        NextProtos:         []string{"x-amzn-mqtt-ca"},
    }, nil
}

func generateCertPath(fileName string) string {
    return "./.cert/" + fileName
}

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
        panic(fmt.Sprintf("failed to discribe AWS iot endpoint: %v", err))
    }
    log.Println("iot endpoint:", *endpoint.EndpointAddress)

    // Connect Broker
    tlsConfig, err := newTLSConfig()
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
    defer client.Disconnect(10)

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
