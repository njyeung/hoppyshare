package mqttclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"time"
	_ "embed"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

//go:embed lambda_output/cert.pem
var certPem []byte

//go:embed lambda_output/key.pem
var keyPem []byte

//go:embed lambda_output/ca.crt
var caPem []byte

func RunMQTT() {

	fmt.Printf("caPem length: %d\n", len(caPem))

	// Load CA cert
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(caPem); !ok {
		log.Fatalf("Failed to append CA certs")
	}

	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		log.Fatalf("Failed to load client certificate/key: %v", err)
	}

	leafCert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		log.Fatalf("Failed to parse client certificate: %v", err)
	}
	cert.Leaf = leafCert
	clientID := leafCert.Subject.CommonName

	tlsConfig := &tls.Config{
		RootCAs:      pool,
		Certificates: []tls.Certificate{cert},
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker("tls://18.188.110.246:8883")
	opts.SetClientID(fmt.Sprintf("%s-%d", clientID, time.Now().Unix()))
	opts.SetTLSConfig(tlsConfig)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetMaxReconnectInterval(30 * time.Second)

	opts.OnConnect = func(c mqtt.Client) {
		fmt.Printf("Connected as %s\n", clientID)

		notesTopic := fmt.Sprintf("users/%s/notes", clientID)
		fmt.Printf("Subscribing to %s\n", notesTopic)
		if token := c.Subscribe(notesTopic, 1, func(c mqtt.Client, m mqtt.Message) {
			fmt.Printf("[NOTES] %s: %d bytes\n", m.Topic(), len(m.Payload()))
		}); token.Wait() && token.Error() != nil {
			log.Fatalf("Subscribe error (notes): %v", token.Error())
		}

		settingsTopic := fmt.Sprintf("users/%s/settings", clientID)
		fmt.Printf("Subscribing to %s\n", settingsTopic)
		if token := c.Subscribe(settingsTopic, 1, func(c mqtt.Client, m mqtt.Message) {
			fmt.Printf("[SETTINGS] %s: %s\n", m.Topic(), string(m.Payload()))
		}); token.Wait() && token.Error() != nil {
			log.Fatalf("Subscribe error (settings): %v", token.Error())
		}
	}

	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		fmt.Printf("Connection lost: %v\n", err)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Connection error: %v", token.Error())
	}

	log.Printf("MQTT mTLS client (%s) running; press Ctrl+C to exit.", clientID)
	select {}
}
