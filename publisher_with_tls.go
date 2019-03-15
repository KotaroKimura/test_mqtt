package main

import (
    "fmt"
    "io/ioutil"
    "crypto/tls"
    "crypto/x509"
    "log"

    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/iot"
)

const (
    RootCAPath   = "AmazonRootCA1.pem"
    CertFilePath = "c394189c23-certificate.pem.crt"
    KeyFilePath  = "c394189c23-private.pem.key"
)

func newTLSConfig() (*tls.Config, error) {
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
	Certificates:       []tls.Certificate{cert},
    }, nil
}

func generateCertPath(fileName string) string {
    return "../../cert/aws-iot/" + fileName
}

func main() {
    // Get AWS Iot Core Endpoint
    s := session.Must(session.NewSession())
    endpoint, err := iot.New(s).DescribeEndpoint(&iot.DescribeEndpointInput{})
    if err != nil {
        panic(fmt.Sprintf("failed to discribe AWS iot endpoint: %v", err))
    }
    log.Println("iot endpoint:", *endpoint.EndpointAddress)

    // Connect Broker
    tlsConfig, err := newTLSConfig()
    if err != nil {
        panic(fmt.Sprintf())
    }

    fmt.Print("hello!")
}