# HoppyShare

A cross-platform clipboard and file sharing tool powered by MQTT and Bluetooth Low Energy (BLE). It’s designed to make device-to-device transfers snappy, private, and reliable, without depending on big third-party clouds.

This repository is a monorepo: it contains everything from the Mosquitto Docker container and backend Lambda functions to the frontend web dashboard and desktop/mobile clients.

todo (for me):
- fix client deleting issue for windows
- mobile apps
  - iphone (prolly not gonna happen)
  - android needs BLE impl
- Add a ping notification or something that you can send from the web to identify which device is which



## Features
- **mTLS + End-to-End Encryption** – All communication is authenticated with mutual TLS. Data payloads are encrypted with a shared group key.
- **MQTT Broker Backbone** – Devices publish/subscribe to a user-scoped topic. Transfers (up to 25MB) are lightweight and real-time.
- **Offline Bluetooth Fallback** – Share files over BLE when Wi‑Fi isn’t available.
- **Cross-Platform Clients**
  - **Desktop client** – Written in Go (Windows, macOS, Linux).
  - **Android client** – Written in Kotlin.
  - **iOS client** – Planned.
- **Web Dashboard** – Manage your account, devices, and settings through the web portal.



## Architecture

```
Frontend (Next.js/React) → Lambda Backend (AWS Lambda)
                                    │
                                    ▼
                        Mosquitto + API (Docker/EC2)
                                    │
                          Desktop / Mobile Clients
```

- **Mosquitto-Docker**
  - MQTT broker with mutual TLS authentication.
  - Bundled **Flask API** for generating certificates, updating ACLs, removing/revoking devices.
  - Includes a **Go watchdog** that prevents channel spam and times out offenders.

- **Lambda-Backend**
  - Handles business logic, database operations, onboarding, and device management.
  - Provides a web-facing API for the frontend.
  - Once a device has its certificates, it bypasses Lambda entirely and communicates directly with the Mosquitto container.

- **Desktop Client** (Go)
  - Integrates with OS keychains for secure storage.
  - Supports both MQTT and BLE message transports.

- **Android Client** (Kotlin)
  - Works similarly to the desktop client with MQTT + BLE.



## Using HoppyShare

Hosted (Default)
1. Go to the web dashboard at hoppyshare.com.
2. Download the desktop client and run it.

Certificates and keys are securely generated and delivered during onboarding. The client handles setup, registers itself for startup, and stores credentials in the OS keychain.



## Self-Hosting

If you want to run HoppyShare on your own infrastructure, you can bypass the Lambda backend and run just the **Mosquitto Docker container**.

### 1. Generate Certs and Keys

Using `openssl`, create the following files:

- `cert.pem` – device certificate
- `key.pem` – device private key
- `ca.crt` – root CA certificate
- `group_key.enc` – shared group key (hex string for E2EE, see Lambda code for Python generator)
- `device_id` – unique identifier for the device

### 2. Prepare the Desktop Client

- Download the latest desktop build (static binaries available for Windows, macOS, Linux).
- Place the files listed above in the same directory as the binary.

### 3. Run in Developer Mode

Before launching the client, set the environment variable:

```bash
DEV_MODE=1
```

Then run the binary
- In DEV_MODE, HoppyShare reads the local certs and keys directly.
- It does not relocate files to the OS keychain.
- It does not perform the initial setup phase (startup registration, relocation).
- This mode is designed for flexibility so you can run HoppyShare however you want.

Normally, the desktop client ships with encrypted certs/keys appended to the end. On first run, it calls the Lambda to decrypt them and moves them into the OS keychain as well as registers itself for auto-startup. Dev mode is useful for testing and self hosting because the client skips setup and keychain relocation.   



## Repository Structure

```
/mosquitto-docker   → Mosquitto broker + Flask API + watchdog
/lambda-backend     → AWS Lambda functions (DB + business logic)
/frontend           → Next.js/React web app
/desktop_client     → Go desktop client (Win/macOS/Linux)
/android_client     → Kotlin Android client
```

