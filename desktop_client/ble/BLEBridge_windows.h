//go:build windows

#ifndef BLE_BRIDGE_WINDOWS_H
#define BLE_BRIDGE_WINDOWS_H

#ifdef __cplusplus
extern "C" {
#endif

// Forward declaration for Go callback
extern void GoOnBLEMessage(char* deviceID, void* data, int length);

// C interface functions
void BLEBridgeStart(const char* clientID, const char* deviceID);
void BLEBridgeStop(void);
void BLEBridgeSend(const void* data, int length);

#ifdef __cplusplus
}
#endif

#endif // BLE_BRIDGE_WINDOWS_H
