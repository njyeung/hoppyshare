package mqttclient

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

//go:embed lambda_output/cert.pem
var certPem []byte

//go:embed lambda_output/key.pem
var keyPem []byte

//go:embed lambda_output/ca.crt
var caPem []byte

//go:embed lambda_output/device_id.txt
var deviceIdRaw []byte

var deviceID [32]byte

var client mqtt.Client
var clientID string

// Connect sets up the MQTT client with mTLS and connects to the broker
func Connect() (string, error) {
	rawID := string(deviceIdRaw)
	if len(rawID) != 32 && len(rawID) != 36 {
		return "", errors.New("deviceID is wrong length")
	}

	copy(deviceID[:], []byte(rawID))

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
	clientID = leafCert.Subject.CommonName

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
		Subscribe(c)
	}

	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		fmt.Printf("Connection lost: %v\n", err)
	}

	opts.OnReconnecting = func(c mqtt.Client, opts *mqtt.ClientOptions) {
		log.Println("Attempting MQTT reconnect...")
	}

	client = mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return "", fmt.Errorf("MQTT connect failed: %w", token.Error())
	}

	log.Printf("Connected to MQTT as %s", clientID)

	return clientID, nil
}

func Subscribe(client mqtt.Client) {

	notesTopic := fmt.Sprintf("users/%s/notes", clientID)
	settingsTopic := fmt.Sprintf("users/%s/settings", clientID)

	log.Printf("Subscribing to %s and %s", notesTopic, settingsTopic)

	if token := client.Subscribe(notesTopic, 1, func(client mqtt.Client, m mqtt.Message) {

		decoded, err := decodeMessage(m.Payload())
		if err != nil {
			log.Printf("Failed to decode message: %v", err)
			return
		}

		if decoded.DeviceID == deviceID {
			// Ignore messages from self
			log.Printf("[FROM SELF] Received %s (%s), %d bytes", decoded.Filename, decoded.Type, len(decoded.Payload))
			return
		}

		log.Printf("[NOTES] Received %s (%s), %d bytes", decoded.Filename, decoded.Type, len(decoded.Payload))
	}); token.Wait() && token.Error() != nil {
		log.Printf("Subscribe error (notes): %v", token.Error())
	}

	if token := client.Subscribe(settingsTopic, 1, func(client mqtt.Client, m mqtt.Message) {
		log.Printf("[SETTINGS] %s: %s", m.Topic(), string(m.Payload()))
	}); token.Wait() && token.Error() != nil {
		log.Printf("Subscribe error (settings): %v", token.Error())
	}
}

func Publish(topic string, data []byte, contentType, filename string) {
	if client == nil || !client.IsConnected() {
		log.Println("Cannot publish: client not connected")
		return
	}

	encoded, err := encodeMessage(contentType, filename, deviceID, data)

	if err != nil {
		log.Printf("Failed to encode message")
	}

	token := client.Publish(topic, 1, false, encoded)
	token.Wait()

	if err := token.Error(); err != nil {
		log.Printf("Publish error %v", err)
	} else {
		log.Printf("Published %s (%s)", filename, contentType)
	}
}
