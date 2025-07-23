//go:build darwin

#import "BLEBridge_darwin.h"

// forward Go callback
extern void GoOnBLEMessage(const char *deviceID, const void *data, int length);

@interface BLEBridgeImpl : NSObject
    <CBCentralManagerDelegate, CBPeripheralManagerDelegate, CBPeripheralDelegate>
@property (nonatomic,strong) CBMutableCharacteristic *localCharacteristic;
@property (nonatomic,strong) CBCharacteristic        *remoteCharacteristic;

@property (nonatomic,strong) CBCentralManager    *centralManager;
@property (nonatomic,strong) CBPeripheralManager *peripheralManager;
@property (nonatomic,strong) CBUUID              *serviceUUID;

@property (nonatomic,strong) NSMutableSet<CBCentral*>  *subscribedCentrals;
@property (nonatomic,strong) NSMutableDictionary<NSUUID*,CBPeripheral*> *discoveredPeripherals;
@property (nonatomic,strong) NSMutableDictionary<NSUUID*,CBPeripheral*> *connectedPeripherals;
@property (nonatomic,copy)   NSString            *deviceID;
@property (nonatomic)        dispatch_queue_t    bleQueue;
@end

@implementation BLEBridgeImpl

- (instancetype)initWithClientID:(NSString*)clientID deviceID:(NSString*)deviceID {
    if (self = [super init]) {
        _deviceID = [deviceID copy];
        _bleQueue = dispatch_queue_create("com.example.ble", DISPATCH_QUEUE_SERIAL);
        // Use simple cross-platform hash
        uint32_t hash = 0;
        const char* str = [clientID UTF8String];
        for (int i = 0; str[i] != '\0'; i++) {
            hash = hash * 31 + (unsigned char)str[i];
        }
        uint16_t shortHash = (uint16_t)(hash & 0xFFFF);
        
        _serviceUUID = [CBUUID UUIDWithString:
            [NSString stringWithFormat:@"0000%04X-0000-1000-8000-00805F9B34FB", shortHash]
        ];
        NSLog(@"Generated service UUID: %@ (hash: %u, shortHash: %u)", _serviceUUID.UUIDString, hash, shortHash);

        _subscribedCentrals   = [NSMutableSet set];
        _discoveredPeripherals = [NSMutableDictionary dictionary];
        _connectedPeripherals  = [NSMutableDictionary dictionary];
    }
    return self;
}

- (void)start {
    NSLog(@"Starting BLE with %@ / %@", self.serviceUUID, self.deviceID);
    self.centralManager    = [[CBCentralManager alloc] initWithDelegate:self queue:self.bleQueue];
    self.peripheralManager = [[CBPeripheralManager alloc] initWithDelegate:self queue:self.bleQueue];
}

- (void)stop {
    NSLog(@"Stopping BLE");
    [self.peripheralManager stopAdvertising];
    [self.peripheralManager removeAllServices];
    [self.centralManager stopScan];
    for (CBPeripheral *p in self.connectedPeripherals.allValues) {
        [self.centralManager cancelPeripheralConnection:p];
    }
    [self.discoveredPeripherals removeAllObjects];
    [self.connectedPeripherals removeAllObjects];
    [self.subscribedCentrals removeAllObjects];
    self.centralManager    = nil;
    self.peripheralManager = nil;
}

- (void)sendData:(NSData*)data {
    if (!self.localCharacteristic) {
        NSLog(@"No localCharacteristic to send on");
        return;
    }
    if (self.subscribedCentrals.count == 0) {
        NSLog(@"No subscribers");
        return;
    }
    NSLog(@"Sending %lu bytes", (unsigned long)data.length);
    BOOL ok = [self.peripheralManager updateValue:data
                                 forCharacteristic:self.localCharacteristic
                              onSubscribedCentrals:nil];
    if (!ok) NSLog(@"UpdateValue failed, will retry when ready");
}

#pragma mark — CBPeripheralManagerDelegate

- (void)peripheralManagerDidUpdateState:(CBPeripheralManager*)pm {
    NSLog(@"Peripheral state %ld", (long)pm.state);
    if (pm.state != CBManagerStatePoweredOn) return;

    // create & store our local characteristic
    _localCharacteristic = [[CBMutableCharacteristic alloc]
        initWithType:[CBUUID UUIDWithString:@"0000FFF1-0000-1000-8000-00805F9B34FB"]
        properties: CBCharacteristicPropertyNotify |
                    CBCharacteristicPropertyRead   |
                    CBCharacteristicPropertyWrite  |
                    CBCharacteristicPropertyWriteWithoutResponse
        value:nil
        permissions: CBAttributePermissionsReadable |
                    CBAttributePermissionsWriteable
    ];

    CBMutableService *svc = [[CBMutableService alloc]
        initWithType:self.serviceUUID primary:YES];
    // register the local characteristic
    svc.characteristics = @[ self.localCharacteristic ];
    [self.peripheralManager addService:svc];
}

- (void)peripheralManager:(CBPeripheralManager*)pm didAddService:(CBService*)svc error:(NSError*)err {
    if (err) { NSLog(@"addService: %@", err); return; }
    NSLog(@"Service added");
    [self.peripheralManager startAdvertising:@{
        CBAdvertisementDataServiceUUIDsKey: @[ self.serviceUUID ],
        CBAdvertisementDataLocalNameKey:   self.deviceID
    }];
}

- (void)peripheralManagerDidStartAdvertising:(CBPeripheralManager*)pm error:(NSError*)err {
    if (err) NSLog(@"startAdvertising: %@", err);
    else    NSLog(@"Advertising as %@", self.deviceID);
}

- (void)peripheralManager:(CBPeripheralManager*)pm
                  central:(CBCentral*)central
didSubscribeToCharacteristic:(CBCharacteristic*)ch {
    NSLog(@"Central subscribed");
    [self.subscribedCentrals addObject:central];
}

- (void)peripheralManager:(CBPeripheralManager*)pm
                  central:(CBCentral*)central
didUnsubscribeFromCharacteristic:(CBCharacteristic*)ch {
    NSLog(@"Central unsubscribed");
    [self.subscribedCentrals removeObject:central];
}

- (void)peripheralManager:(CBPeripheralManager*)pm
      didReceiveWriteRequests:(NSArray<CBATTRequest*>*)reqs {
    for (CBATTRequest *r in reqs) {
        // compare against our localCharacteristic UUID
        if ([r.characteristic.UUID isEqual:self.localCharacteristic.UUID]) {
            NSData *d = r.value;
            if (d) {
                NSLog(@"Received %lu bytes", (unsigned long)d.length);
                [self handleReceivedData:d from:@"local"];
            }
            [pm respondToRequest:r withResult:CBATTErrorSuccess];
        }
    }
}

- (void)peripheralManagerIsReadyToUpdateSubscribers:(CBPeripheralManager*)pm {
    NSLog(@"Ready to retry updateValue");
}

#pragma mark — CBCentralManagerDelegate

- (void)centralManagerDidUpdateState:(CBCentralManager*)cm {
    NSLog(@"Central state %ld", (long)cm.state);
    if (cm.state != CBManagerStatePoweredOn) return;
    [cm scanForPeripheralsWithServices:@[ self.serviceUUID ]
                             options:@{ CBCentralManagerScanOptionAllowDuplicatesKey: @NO }];
    NSLog(@"Scanning for %@", self.serviceUUID);
}

- (void)centralManager:(CBCentralManager*)cm
    didDiscoverPeripheral:(CBPeripheral*)p
        advertisementData:(NSDictionary*)adv
                     RSSI:(NSNumber*)rssi {
    NSString *name = adv[CBAdvertisementDataLocalNameKey] ?: p.name ?: @"?";
    if ([name isEqualToString:self.deviceID]) {
        NSLog(@"Ignoring self-advertisement");
        return;
    }
    self.discoveredPeripherals[p.identifier] = p;
    p.delegate = self;
    [cm connectPeripheral:p options:nil];
}

- (void)centralManager:(CBCentralManager*)cm
 didConnectPeripheral:(CBPeripheral*)p {
    // NSLog(@"Connected to %@", p.name ?: @"?");
    self.connectedPeripherals[p.identifier] = p;
    [p discoverServices:@[ self.serviceUUID ]];
}

- (void)centralManager:(CBCentralManager*)cm
didFailToConnectPeripheral:(CBPeripheral*)p
                 error:(NSError*)err {
    NSLog(@"Failed to connect: %@", err.localizedDescription);
    [self.discoveredPeripherals removeObjectForKey:p.identifier];
}

- (void)centralManager:(CBCentralManager*)cm
didDisconnectPeripheral:(CBPeripheral*)p
                 error:(NSError*)err {
    NSLog(@"Disconnected: %@ (%@)", p.name ?: @"?", err.localizedDescription ?: @"");
    [self.connectedPeripherals removeObjectForKey:p.identifier];
    [self.discoveredPeripherals removeObjectForKey:p.identifier];
    if (cm.state == CBManagerStatePoweredOn && self.connectedPeripherals.count==0) {
        [cm scanForPeripheralsWithServices:@[ self.serviceUUID ]
                                 options:@{CBCentralManagerScanOptionAllowDuplicatesKey:@NO}];
    }
}

#pragma mark — CBPeripheralDelegate

- (void)peripheral:(CBPeripheral*)p didDiscoverServices:(NSError*)err {
    if (err) { NSLog(@"discoverServices: %@", err); return; }
    for (CBService *s in p.services) {
        if ([s.UUID isEqual:self.serviceUUID]) {
            [p discoverCharacteristics:@[
                [CBUUID UUIDWithString:@"0000FFF1-0000-1000-8000-00805F9B34FB"]
            ] forService:s];
        }
    }
}

- (void)peripheral:(CBPeripheral*)p
didDiscoverCharacteristicsForService:(CBService*)s
              error:(NSError*)err {
    if (err) { NSLog(@"discoverChars: %@", err); return; }
    for (CBCharacteristic *c in s.characteristics) {
        if ([c.UUID.UUIDString isEqualToString:@"0000FFF1-0000-1000-8000-00805F9B34FB"]) {
            [p setNotifyValue:YES forCharacteristic:c];
            _remoteCharacteristic = c;
            if (c.properties & CBCharacteristicPropertyRead) {
                [p readValueForCharacteristic:c];
            }
            // no longer assigning to messageCharacteristic here
        }
    }
}

- (void)peripheral:(CBPeripheral*)p
didUpdateValueForCharacteristic:(CBCharacteristic*)c
              error:(NSError*)err {
    if (err) { return; }
    NSData *d = c.value;
    [self handleReceivedData:d from:p.name ?: @"?"];
}

- (void)peripheral:(CBPeripheral*)p
didWriteValueForCharacteristic:(CBCharacteristic*)c
              error:(NSError*)err {
    if (err) NSLog(@"writeValue: %@", err);
    else    NSLog(@"wrote value");
}

- (void)handleReceivedData:(NSData*)data from:(NSString*)name {
    GoOnBLEMessage(name.UTF8String, data.bytes, (int)data.length);
}

@end

static BLEBridgeImpl *gBridge = nil;

void BLEBridgeStart(const char *clientID, const char *deviceID) {
    NSString *c = [NSString stringWithUTF8String:clientID];
    NSString *d = [NSString stringWithUTF8String:deviceID];
    gBridge = [[BLEBridgeImpl alloc] initWithClientID:c deviceID:d];
    [gBridge start];
}

void BLEBridgeStop(void) {
    [gBridge stop];
    gBridge = nil;
}

void BLEBridgeSend(const void *data, int length) {
    NSData *payload = [NSData dataWithBytes:data length:length];
    [gBridge sendData:payload];
}
