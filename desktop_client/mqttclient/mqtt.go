package mqttclient

import (
	"crypto/tls"
	"crypto/x509"
	"desktop_client/config"
	"desktop_client/settings"
	"desktop_client/systrayhelpers"
	_ "embed"
	"fmt"
	"log"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	client   mqtt.Client
	clientID string
)

// Connect sets up the MQTT client with mTLS and connects to the broker
func Connect() (string, error) {

	cert, err := tls.X509KeyPair(config.CertPem, config.KeyPem)
	if err != nil {
		log.Fatalf("Failed to load client certificate/key: %v", err)
	}

	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(config.CAPem); !ok {
		log.Fatalf("Failed to append CA certs")
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

	go func() {
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
			systrayhelpers.SetTooltip("Connected")
		}

		opts.OnConnectionLost = func(c mqtt.Client, err error) {
			fmt.Printf("Connection lost: %v\n", err)

			systrayhelpers.SetTooltip("Disconnected")
		}

		opts.OnReconnecting = func(c mqtt.Client, opts *mqtt.ClientOptions) {
			log.Println("Attempting MQTT reconnect...")
		}

		client = mqtt.NewClient(opts)

		token := client.Connect()
		token.Wait()
		err = token.Error()
		if err != nil {
			log.Printf("Count not connect to MQTT as %s", clientID)
		}

		log.Printf("Connected to MQTT as %s", clientID)
	}()

	return clientID, nil
}

func Disconnect() {
	if client != nil && client.IsConnected() {
		client.Disconnect(250)
	}
}

type lastMessage struct {
	Filename    string
	ContentType string
	Payload     []byte
}

var (
	lastMsg   lastMessage
	lastMsgMu sync.RWMutex
)

func Subscribe(client mqtt.Client) {

	notesTopic := fmt.Sprintf("users/%s/notes", clientID)
	settingsTopic := fmt.Sprintf("users/%s/settings", clientID)

	log.Printf("Subscribing to %s and %s", notesTopic, settingsTopic)

	if token := client.Subscribe(notesTopic, 1, func(client mqtt.Client, m mqtt.Message) {

		if !settings.GetSettings().Enabled {
			return
		}

		decoded, err := DecodeMessage(m.Payload())

		if err != nil {
			log.Printf("Failed to decode message: %v", err)
			return
		}

		if !settings.GetSettings().SendToSelf && decoded.DeviceID == hashDeviceID(config.DeviceID) {
			return
		}

		cacheMsg(decoded.Filename, decoded.Type, decoded.Payload)

		// log.Printf("[NOTES] Received %s (%s), %d bytes", decoded.Filename, decoded.Type, len(decoded.Payload))
	}); token.Wait() && token.Error() != nil {
		// log.Printf("Subscribe error (notes): %v", token.Error())
	}

	if token := client.Subscribe(settingsTopic, 1, func(client mqtt.Client, m mqtt.Message) {
		// log.Printf("[SETTINGS] %s: %s", m.Topic(), string(m.Payload()))

		settings.ParseSettings(m.Payload())
	}); token.Wait() && token.Error() != nil {
		// log.Printf("Subscribe error (settings): %v", token.Error())
	}
}

var onMessage func()

func SetOnMessageCallback(cb func()) {
	onMessage = cb
}
func cacheMsg(fname, ctype string, payload []byte) {
	lastMsgMu.Lock()
	lastMsg = lastMessage{
		Filename:    fname,
		ContentType: ctype,
		Payload:     payload,
	}
	lastMsgMu.Unlock()

	if onMessage != nil {
		onMessage()
	}

}

func ClearMsg() {
	lastMsgMu.Lock()
	defer lastMsgMu.Unlock()
	lastMsg = lastMessage{}
}

func GetLastMessage() (filename, contentType string, data []byte, ok bool) {
	lastMsgMu.RLock()

	defer lastMsgMu.RUnlock()

	if len(lastMsg.Payload) == 0 {
		return "", "", nil, false
	}

	return lastMsg.Filename, lastMsg.ContentType, lastMsg.Payload, true
}

func Publish(topic string, data []byte, contentType, filename string) error {
	if client == nil || !client.IsConnected() {
		return fmt.Errorf("cannot publish: client not connected")
	}

	encoded, err := EncodeMessage(contentType, filename, config.DeviceID, data)

	if err != nil {
		log.Printf("Failed to encode message")
	}

	token := client.Publish(topic, 1, false, encoded)

	ok := token.WaitTimeout(10 * time.Second)
	if !ok {
		return fmt.Errorf("cannot publish: client not connected")
	}

	if err := token.Error(); err != nil {
		log.Printf("Publish error %v", err)
	} else {
		log.Printf("Published %s (%s)", filename, contentType)
	}

	return nil
}
