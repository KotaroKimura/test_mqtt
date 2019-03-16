package tlsconfig

import (
    "io/ioutil"
    "crypto/tls"
    "crypto/x509"
    "log"
    "os"

    "github.com/joho/godotenv"
)

func NewTLSConfig() (*tls.Config, error) {
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
