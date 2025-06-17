package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "strings"
    "time"
    "bytes"
    "net/http"

    mqtt "github.com/eclipse/paho.mqtt.golang"
    "github.com/redis/go-redis/v9"
)

var (
    redisClient *redis.Client
    ctx         = context.Background()
    messageCap  = 30
    windowSec   = 10
)

func main() {
    // Redis setup
    redisClient = redis.NewClient(&redis.Options{
        Addr: "redis:6379",
    })

    restoreAllUsers()

    // MQTT client setup
    opts := mqtt.NewClientOptions().
        AddBroker("tls://mosquitto:8883").
        SetClientID("watchdog").
        SetTLSConfig(loadTLS("watchdog.crt", "watchdog.key", "/certs/ca.crt")).
        SetDefaultPublishHandler(handleMessage)

    client := mqtt.NewClient(opts)
    for {
        token := client.Connect()
        token.Wait()
        if token.Error() != nil {
            log.Printf("MQTT connect failed: %v, retrying in 2s...", token.Error())
            time.Sleep(10 * time.Second)
            continue
        }
        break
    }

    // Subscribe to all note traffic
    if token := client.Subscribe("users/+/notes", 0, handleMessage); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }

    log.Println("Watchdog running...")
    select {}
}

func reloadMosquitto() {
    url := "http://mosquitto:5000/reload"

    req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte{}))
    if err != nil {
        log.Printf("Error creating HTTP request to Flask reload: %v", err)
        return
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Printf("Error calling Flask /reload: %v", err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        log.Printf("Flask /reload failed with status: %d", resp.StatusCode)
        return
    }

    log.Println("🔄 Mosquitto reload triggered via Flask")
}

func restoreAllUsers() {
    files, err := os.ReadDir("/dynamic_acl")
    if err != nil {
        log.Printf("Error reading dynamic_acl directory: %v", err)
        return
    }

    for _, f := range files {
        name := f.Name()
        if !strings.HasPrefix(name, "user_") || !strings.HasSuffix(name, ".acl") {
            continue
        }

        path := "/dynamic_acl/" + name
        content, err := os.ReadFile(path)
        if err != nil {
            log.Printf("Error reading %s: %v", name, err)
            continue
        }

        if strings.Contains(string(content), "deny write") {
            uid := strings.TrimSuffix(strings.TrimPrefix(name, "user_"), ".acl")
            log.Printf("🔁 Restoring write access for %s", uid)

            rule := fmt.Sprintf(`user %s
            topic write users/%s/notes
            topic write users/%s/settings
            `, uid, uid, uid)

            if err := os.WriteFile(path, []byte(rule), 0644); err != nil {
                log.Printf("Failed to restore %s: %v", name, err)
                continue
            }
        }
    }
    
    reloadMosquitto()
}

func handleMessage(client mqtt.Client, msg mqtt.Message) {
    topic := msg.Topic()
    parts := strings.Split(topic, "/")

    // Expect exactly: users/<uid>/notes
    if len(parts) != 3 || parts[0] != "users" || parts[2] != "notes" {
        log.Printf("Ignoring topic: %s", topic)
        return
    }

    uid := parts[1]
    key := fmt.Sprintf("msgcount:%s", uid)

    // INCR and EXPIRE
    count, err := redisClient.Incr(ctx, key).Result()
    if err != nil {
        log.Printf("Redis error: %v", err)
        return
    }
    if count == 1 {
        redisClient.Expire(ctx, key, time.Duration(windowSec)*time.Second)
    }

    log.Printf("User %s sent message #%d", uid, count)

    if count > int64(messageCap) {
        log.Printf("⚠️  User %s exceeded message limit", uid)
        blockUser(uid)
    }
}

func blockUser(uid string) {
    aclFile := fmt.Sprintf("/dynamic_acl/user_%s.acl", uid)
    rule := fmt.Sprintf(`user %s
    topic deny write users/%s/notes
    topic deny write users/%s/settings
    `, uid, uid, uid)

    // Write block rule to dedicated file
    err := os.WriteFile(aclFile, []byte(rule), 0644)
    if err != nil {
        log.Printf("Failed to write ACL for %s: %v", uid, err)
        return
    }

    // Reload Mosquitto
    reloadMosquitto()

    log.Printf("⛔ Blocked user %s for 10 minutes", uid)

    // Schedule unblock after 10 minutes
    go func() {
        time.Sleep(10 * time.Minute)

        rule := fmt.Sprintf(`user %s
        topic write users/%s/notes
        topic write users/%s/settings
        `, uid, uid, uid)

        if err := os.WriteFile(aclFile, []byte(rule), 0644); err != nil {
            panic(err.Error())
        }

        // Reload Mosquitto
        reloadMosquitto()

        log.Printf("✅ Unblocked user %s", uid)
    }()
}

