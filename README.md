# snap-notes
cross platform file/clipboard sharing app so i can send stuff from one PC to the other. Sits in systray. Uses mqtt broker and bluetooth LE.

todo:
- icons
- webpage
- mobile apps, store keys in the key store
  - iphone
  - android
- Device settings integration
- Store keys using go-keychain
  - lambda encrypt certs and keys in binary but leave device_id.txt not encrypted
  - On first run, client reaches out to api /device_id, for auth pass its encrypted public cert
  - Lambda decrypts the cert using its random key, check that its == to the device's pub cert in DB, returns with the one-time-key to decrypt.
  - client decrypts its certs and keys, then stores them in keychain.
- Register the binary to run on startup
  - Linux
  - MacOS
  - Windows  
- Bluetooth BLE for offline sharing
