#import <Foundation/Foundation.h>
#import <CoreBluetooth/CoreBluetooth.h>

#ifndef BLEBRIDGE_H
#define BLEBRIDGE_H

#ifdef __cplusplus
extern "C" {
#endif

void BLEBridgeStart(const char *clientID, const char *deviceID);

void BLEBridgeStop(void);

void BLEBridgeSend(const void *data, int length);

#ifdef __cplusplus
}
#endif

#endif /* BLEBRIDGE_H */
