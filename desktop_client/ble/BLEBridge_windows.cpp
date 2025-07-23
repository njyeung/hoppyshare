//go:build windows

#include "BLEBridge_windows.h"
#include <windows.h>
#define WINRT_NO_MODULE_LOCK
#define _DISABLE_CONSTEXPR_MUTEX_CONSTRUCTOR
#define WINRT_LEAN_AND_MEAN
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wtemplate-body"
#include <winrt/base.h>
#include <winrt/Windows.Foundation.h>
#include <winrt/Windows.Foundation.Collections.h>
#include <winrt/Windows.Devices.Bluetooth.h>
#include <winrt/Windows.Devices.Bluetooth.Advertisement.h>
#include <winrt/Windows.Devices.Bluetooth.GenericAttributeProfile.h>
#include <winrt/Windows.Devices.Radios.h>
#include <winrt/Windows.Storage.Streams.h>
#pragma GCC diagnostic pop
#include <string>
#include <map>
#include <memory>
#include <iostream>

using namespace winrt;
using namespace Windows::Foundation;
using namespace Windows::Devices::Bluetooth;
using namespace Windows::Devices::Bluetooth::Advertisement;
using namespace Windows::Devices::Bluetooth::GenericAttributeProfile;
using namespace Windows::Devices::Radios;
using namespace Windows::Storage::Streams;

class BLEBridge {
private:
    std::string m_clientID;
    std::string m_deviceID;
    guid m_serviceUuid;
    guid m_characteristicUuid;
    
    // Central (Scanner) components
    BluetoothLEAdvertisementWatcher m_watcher{nullptr};
    std::map<uint64_t, BluetoothLEDevice> m_discoveredDevices;
    std::map<uint64_t, GattDeviceService> m_connectedServices;
    std::map<uint64_t, GattCharacteristic> m_remoteCharacteristics;
    
    // Peripheral (Advertiser) components  
    BluetoothLEAdvertisementPublisher m_publisher{nullptr};
    GattServiceProvider m_serviceProvider{nullptr};
    GattLocalCharacteristic m_localCharacteristic{nullptr};
    std::map<std::string, GattSession> m_subscribedSessions;
    
    bool m_isStarted = false;

    uint32_t simpleHash(const std::string& str) {
        uint32_t hash = 0;
        for (char c : str) {
            hash = hash * 31 + static_cast<unsigned char>(c);
        }
        return hash;
    }

    // Generate service UUID from clientID hash
    void generateServiceUuid() {
        uint32_t hash = simpleHash(m_clientID);
        uint16_t shortHash = static_cast<uint16_t>(hash & 0xFFFF);
        
        std::wcout << L"Generated service UUID hash: " << hash << L", shortHash: " << shortHash << std::endl;
        
        // Format: 0000XXXX-0000-1000-8000-00805F9B34FB
        m_serviceUuid = guid{
            static_cast<uint32_t>((0x0000 << 16) | shortHash), 0x0000, 0x1000,
            {0x80, 0x00, 0x00, 0x80, 0x5F, 0x9B, 0x34, 0xFB}
        };
        
        // Characteristic: 0000FFF1-0000-1000-8000-00805F9B34FB
        m_characteristicUuid = guid{
            0x0000FFF1, 0x0000, 0x1000,
            {0x80, 0x00, 0x00, 0x80, 0x5F, 0x9B, 0x34, 0xFB}
        };
    }

public:
    BLEBridge(const std::string& clientID, const std::string& deviceID)
        : m_clientID(clientID), m_deviceID(deviceID) {
        generateServiceUuid();
    }

    void start() {
        if (m_isStarted) return;
        
        std::wcout << L"Starting BLE Bridge for Windows" << std::endl;
        
        // requires UWP manifest capabilities  
        // startPeripheral();
        
        // Start as Central (Scanner)
        startCentral();
        
        m_isStarted = true;
    }

    void stop() {
        if (!m_isStarted) return;
        
        std::wcout << L"Stopping BLE Bridge" << std::endl;
        
        // Stop advertising
        if (m_publisher) {
            m_publisher.Stop();
            m_publisher = nullptr;
        }
        
        // Stop scanning
        if (m_watcher) {
            m_watcher.Stop();
            m_watcher = nullptr;
        }
        
        // Disconnect from all devices
        m_discoveredDevices.clear();
        m_connectedServices.clear();
        m_remoteCharacteristics.clear();
        m_subscribedSessions.clear();
        
        if (m_serviceProvider) {
            m_serviceProvider = nullptr;
        }
        
        m_isStarted = false;
    }

    void sendData(const void* data, int length) {
        if (!m_localCharacteristic || length == 0) {
            std::wcout << L"No local characteristic or empty data" << std::endl;
            return;
        }
        
        try {
            // Create data buffer
            DataWriter writer;
            auto dataArray = winrt::array_view<const uint8_t>(
                static_cast<const uint8_t*>(data), length);
            writer.WriteBytes(dataArray);
            auto buffer = writer.DetachBuffer();
            
            // Notify all subscribed clients using correct API
            try {
                m_localCharacteristic.NotifyValueAsync(buffer);
                std::wcout << L"Sent " << length << L" bytes to all subscribed clients" << std::endl;
            } catch (...) {
                std::wcout << L"Failed to send to subscribed clients" << std::endl;
            }
        } catch (...) {
            std::wcout << L"Failed to send data" << std::endl;
        }
    }

private:
    void startPeripheral() {
        try {
            // Create GATT service provider
            auto serviceResult = GattServiceProvider::CreateAsync(m_serviceUuid).get();
            if (serviceResult.Error() != BluetoothError::Success) {
                std::wcout << L"Failed to create service provider" << std::endl;
                return;
            }
            
            m_serviceProvider = serviceResult.ServiceProvider();
            
            // Create local characteristic
            GattLocalCharacteristicParameters charParams;
            charParams.CharacteristicProperties(
                GattCharacteristicProperties::Read |
                GattCharacteristicProperties::Write |
                GattCharacteristicProperties::WriteWithoutResponse |
                GattCharacteristicProperties::Notify
            );
            charParams.ReadProtectionLevel(GattProtectionLevel::Plain);
            charParams.WriteProtectionLevel(GattProtectionLevel::Plain);
            
            auto charResult = m_serviceProvider.Service().CreateCharacteristicAsync(
                m_characteristicUuid, charParams).get();
            
            if (charResult.Error() != BluetoothError::Success) {
                std::wcout << L"Failed to create characteristic" << std::endl;
                return;
            }
            
            m_localCharacteristic = charResult.Characteristic();
            
            // Set up write request handler
            m_localCharacteristic.WriteRequested([this](
                GattLocalCharacteristic const&,
                GattWriteRequestedEventArgs const& args) {
                handleWriteRequest(args);
            });
            
            // Set up subscription handler
            m_localCharacteristic.SubscribedClientsChanged([this](
                GattLocalCharacteristic const&,
                IInspectable const&) {
                handleSubscriptionChanged();
            });
            
            
            // Start advertising
            startAdvertising();
            
            std::wcout << L"Peripheral mode started" << std::endl;
            
        } catch (...) {
            std::wcout << L"Failed to start peripheral mode" << std::endl;
        }
    }

    void startAdvertising() {
        try {
            // Create publisher and set advertisement data BEFORE setting up handlers
            m_publisher = BluetoothLEAdvertisementPublisher();
            
            // Set advertisement data immediately
            auto advertisement = m_publisher.Advertisement();
            advertisement.LocalName(winrt::to_hstring("DC-Win"));
            std::wcout << L"Set advertisement local name to: DC-Win" << std::endl;
            
            // Start advertising
            m_publisher.Start();
            std::wcout << L"Called Start() on advertising publisher for device: " 
                      << winrt::to_hstring(m_deviceID).c_str() << std::endl;
            
        } catch (winrt::hresult_error const& ex) {
            std::wcout << L"WinRT error in startAdvertising: 0x" << std::hex 
                      << ex.code() << std::dec << L" - " << ex.message().c_str() << std::endl;
        } catch (...) {
            std::wcout << L"Unknown error in startAdvertising" << std::endl;
        }
    }

    void startCentral() {
        try {
            m_watcher = BluetoothLEAdvertisementWatcher();
            
            // Only scan for our service UUID
            m_watcher.ScanningMode(BluetoothLEScanningMode::Active);
            
            // Set up advertisement received handler
            m_watcher.Received([this](
                BluetoothLEAdvertisementWatcher const&,
                BluetoothLEAdvertisementReceivedEventArgs const& args) {
                handleAdvertisementReceived(args);
            });
            
            // Start scanning
            m_watcher.Start();
            std::wcout << L"Central scanning started" << std::endl;
            
        } catch (...) {
            std::wcout << L"Failed to start central mode" << std::endl;
        }
    }

    void handleWriteRequest(GattWriteRequestedEventArgs const& args) {
        try {
            auto deferral = args.GetDeferral();
            auto request = args.GetRequestAsync().get();
            
            if (request.Value().Length() > 0) {
                // Read the data
                DataReader reader = DataReader::FromBuffer(request.Value());
                auto length = reader.UnconsumedBufferLength();
                std::vector<uint8_t> data(length);
                reader.ReadBytes(data);
                
                // Call Go callback
                std::string deviceName = "local"; // Could get actual device name
                GoOnBLEMessage(const_cast<char*>(deviceName.c_str()), data.data(), static_cast<int>(length));
                
                std::wcout << L"Received " << length << L" bytes" << std::endl;
            }
            
            // Respond with success
            if (request.Option() == GattWriteOption::WriteWithResponse) {
                request.Respond();
            }
            
            deferral.Complete();
        } catch (...) {
            std::wcout << L"Failed to handle write request" << std::endl;
        }
    }

    void handleSubscriptionChanged() {
        try {
            m_subscribedSessions.clear();
            for (auto session : m_localCharacteristic.SubscribedClients()) {
                std::string sessionId = winrt::to_string(winrt::to_hstring(session.Session().DeviceId().Id()));
                // Use emplace to avoid default constructor issues with WinRT types
                m_subscribedSessions.emplace(sessionId, session.Session());
                std::wcout << L"Client subscribed" << std::endl;
            }
        } catch (...) {
            std::wcout << L"Failed to handle subscription change" << std::endl;
        }
    }

    void handleAdvertisementReceived(BluetoothLEAdvertisementReceivedEventArgs const& args) {
        try {
            auto advertisement = args.Advertisement();
            auto localName = advertisement.LocalName();
            
            // Look for devices with our exact service UUID (like macOS does)
            bool hasOurService = false;
            for (auto uuid : advertisement.ServiceUuids()) {
                if (uuid == m_serviceUuid) {
                    hasOurService = true;
                    break;
                }
            }
            
            if (!hasOurService) {
                return;
            }
            
            // Don't connect to ourselves
            std::string deviceName = winrt::to_string(localName);
            if (!deviceName.empty() && deviceName == m_deviceID) {
                return;
            }
            
            std::wcout << L"Discovered device: " << localName.c_str() << std::endl;
            connectToDevice(args.BluetoothAddress());
            
        } catch (...) {
            std::wcout << L"Failed to handle advertisement" << std::endl;
        }
    }

    void connectToDevice(uint64_t bluetoothAddress) {
        try {
            // Check if already connected
            if (m_connectedServices.find(bluetoothAddress) != m_connectedServices.end()) {
                return;
            }
            
            // Get device from address
            auto deviceResult = BluetoothLEDevice::FromBluetoothAddressAsync(bluetoothAddress).get();
            if (!deviceResult) {
                return;
            }
            
            auto device = deviceResult;
            
            // Get GATT services
            auto servicesResult = device.GetGattServicesAsync().get();
            if (servicesResult.Status() != GattCommunicationStatus::Success) {
                return;
            }
            
            // Find our exact service (like macOS does)
            GattDeviceService targetService{nullptr};
            for (auto service : servicesResult.Services()) {
                if (service.Uuid() == m_serviceUuid) {
                    targetService = service;
                    break;
                }
            }
            
            if (!targetService) {
                return;
            }
            
            m_connectedServices.emplace(bluetoothAddress, targetService);
            
            // Get characteristics
            auto charsResult = targetService.GetCharacteristicsForUuidAsync(m_characteristicUuid).get();
            if (charsResult.Status() != GattCommunicationStatus::Success ||
                charsResult.Characteristics().Size() == 0) {
                return;
            }
            
            auto characteristic = charsResult.Characteristics().GetAt(0);
            m_remoteCharacteristics.emplace(bluetoothAddress, characteristic);
            
            // Subscribe to notifications
            auto status = characteristic.WriteClientCharacteristicConfigurationDescriptorAsync(
                GattClientCharacteristicConfigurationDescriptorValue::Notify).get();
            
            if (status == GattCommunicationStatus::Success) {
                characteristic.ValueChanged([this, bluetoothAddress](
                    GattCharacteristic const&,
                    GattValueChangedEventArgs const& args) {
                    handleValueChanged(bluetoothAddress, args);
                });
                
                std::wcout << L"Connected to BLE device" << std::endl;
            }
            
        } catch (...) {
            
        }
    }

    void handleValueChanged(uint64_t bluetoothAddress, GattValueChangedEventArgs const& args) {
        try {
            auto buffer = args.CharacteristicValue();
            if (buffer.Length() > 0) {
                DataReader reader = DataReader::FromBuffer(buffer);
                auto length = reader.UnconsumedBufferLength();
                std::vector<uint8_t> data(length);
                reader.ReadBytes(data);
                
                // Call Go callback with device address as ID
                std::string deviceId = std::to_string(bluetoothAddress);
                GoOnBLEMessage(const_cast<char*>(deviceId.c_str()), data.data(), static_cast<int>(length));
                
                std::wcout << L"Received " << length << L" bytes from remote device" << std::endl;
            }
        } catch (...) {
            std::wcout << L"Failed to handle value change" << std::endl;
        }
    }
};

static std::unique_ptr<BLEBridge> g_bridge = nullptr;

extern "C" {

void BLEBridgeStart(const char* clientID, const char* deviceID) {
    try {
        init_apartment();
        g_bridge = std::make_unique<BLEBridge>(clientID, deviceID);
        g_bridge->start();
    } catch (...) {
        std::wcout << L"Failed to start BLE bridge" << std::endl;
    }
}

void BLEBridgeStop(void) {
    if (g_bridge) {
        g_bridge->stop();
        g_bridge.reset();
    }
}

void BLEBridgeSend(const void* data, int length) {
    if (g_bridge) {
        g_bridge->sendData(data, length);
    }
}

} // extern "C"