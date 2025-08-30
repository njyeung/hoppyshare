package com.example.hoppyshare

import android.Manifest
import android.annotation.SuppressLint
import android.bluetooth.*
import android.bluetooth.le.*
import android.content.Context
import android.content.pm.PackageManager
import android.os.Build
import android.os.ParcelUuid
import android.util.Log
import androidx.core.app.ActivityCompat
import kotlinx.coroutines.*
import java.nio.ByteBuffer
import java.nio.ByteOrder
import java.util.*
import java.util.concurrent.ConcurrentHashMap
import kotlin.random.Random

class BLEManager(
    private val context: Context,
    private val clientId: String,
    private val deviceId: String
) {
    companion object {
        private const val TAG = "BLEManager"
        private const val MAX_CHUNK_SIZE = 500
        
        // Standard UUIDs (matching Darwin implementation)
        private const val CHARACTERISTIC_UUID = "0000FFF1-0000-1000-8000-00805F9B34FB"
    }

    private val bluetoothManager = context.getSystemService(Context.BLUETOOTH_SERVICE) as BluetoothManager
    private val bluetoothAdapter = bluetoothManager.adapter
    
    // Service UUID based on clientId hash (same as Darwin implementation)
    private val serviceUUID: UUID = generateServiceUUID(clientId)
    
    // BLE Central (Scanner) components
    private var bluetoothLeScanner: BluetoothLeScanner? = null
    private var scanCallback: ScanCallback? = null
    
    // BLE Peripheral (Server) components
    private var bluetoothGattServer: BluetoothGattServer? = null
    private var advertiser: BluetoothLeAdvertiser? = null
    private var advertiseCallback: AdvertiseCallback? = null
    
    // Connected devices
    private val connectedDevices = ConcurrentHashMap<String, BluetoothGatt>()
    private val subscribedCentrals = ConcurrentHashMap<String, BluetoothDevice>()
    
    // Message handling
    private var onMessageCallback: (() -> Unit)? = null
    private var lastMessage: ByteArray? = null
    private val messageLock = Any()
    
    // Chunk assembly (following Go implementation)
    private val assemblyBuffers = ConcurrentHashMap<String, ChunkBuffer>()
    
    data class ChunkBuffer(
        val chunks: MutableMap<Int, ByteArray> = mutableMapOf(),
        val received: MutableSet<Int> = mutableSetOf(),
        var total: Int = -1
    )
    
    private var isStarted = false
    private val startLock = Any()
    
    private fun generateServiceUUID(clientId: String): UUID {
        // Same hash algorithm as Darwin implementation
        var hash = 0u
        for (char in clientId) {
            hash = hash * 31u + char.code.toUInt()
        }
        val shortHash = (hash and 0xFFFFu).toInt()
        val uuidString = String.format("0000%04X-0000-1000-8000-00805F9B34FB", shortHash)
        Log.d(TAG, "Generated service UUID: $uuidString (hash: $hash, shortHash: $shortHash)")
        return UUID.fromString(uuidString)
    }
    
    @SuppressLint("MissingPermission")
    fun start(): Boolean {
        synchronized(startLock) {
            Log.d(TAG, "start() called, isStarted: $isStarted")
            
            if (isStarted) {
                Log.d(TAG, "BLE already started")
                return true
            }
            
            Log.d(TAG, "Bluetooth adapter enabled: ${bluetoothAdapter.isEnabled}")
            if (!bluetoothAdapter.isEnabled) {
                Log.e(TAG, "Bluetooth not enabled")
                return false
            }
            
            val hasPermissions = hasRequiredPermissions()
            Log.d(TAG, "Has required permissions: $hasPermissions")
            if (!hasPermissions) {
                Log.e(TAG, "Missing required BLE permissions")
                logMissingPermissions()
                return false
            }
            
            Log.d(TAG, "Starting BLE with serviceUUID: $serviceUUID, deviceId: $deviceId")
            
            try {
                Log.d(TAG, "About to set up GATT server...")
                setupGattServer()
                Log.d(TAG, "GATT server setup completed, about to start advertising...")
                startAdvertising()
                Log.d(TAG, "Advertising started, about to start scanning...")
                startScanning()
                Log.d(TAG, "Scanning started, marking as started...")
                isStarted = true
                Log.d(TAG, "BLE started successfully")
                return true
            } catch (e: Exception) {
                Log.e(TAG, "Failed to start BLE: ${e.message}", e)
                stop()
                return false
            }
        }
    }
    
    @SuppressLint("MissingPermission") 
    fun stop() {
        synchronized(startLock) {
            if (!isStarted) return
            
            Log.d(TAG, "Stopping BLE")
            
            try {
                // Stop advertising
                advertiser?.stopAdvertising(advertiseCallback)
                advertiseCallback = null
                advertiser = null
                
                // Stop scanning
                bluetoothLeScanner?.stopScan(scanCallback)
                scanCallback = null
                bluetoothLeScanner = null
                
                // Disconnect all connected devices
                for (gatt in connectedDevices.values) {
                    gatt.disconnect()
                    gatt.close()
                }
                connectedDevices.clear()
                subscribedCentrals.clear()
                
                // Close GATT server
                bluetoothGattServer?.close()
                bluetoothGattServer = null
                
                // Clear message buffers
                assemblyBuffers.clear()
                
                isStarted = false
                Log.d(TAG, "BLE stopped")
            } catch (e: Exception) {
                Log.e(TAG, "Error stopping BLE: ${e.message}", e)
            }
        }
    }
    
    private fun hasRequiredPermissions(): Boolean {
        val requiredPermissions = mutableListOf<String>()
        
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            requiredPermissions.addAll(listOf(
                Manifest.permission.BLUETOOTH_SCAN,
                Manifest.permission.BLUETOOTH_ADVERTISE,
                Manifest.permission.BLUETOOTH_CONNECT
            ))
        } else {
            requiredPermissions.addAll(listOf(
                Manifest.permission.BLUETOOTH,
                Manifest.permission.BLUETOOTH_ADMIN,
                Manifest.permission.ACCESS_FINE_LOCATION
            ))
        }
        
        return requiredPermissions.all { permission ->
            ActivityCompat.checkSelfPermission(context, permission) == PackageManager.PERMISSION_GRANTED
        }
    }
    
    private fun logMissingPermissions() {
        val requiredPermissions = mutableListOf<String>()
        
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            requiredPermissions.addAll(listOf(
                Manifest.permission.BLUETOOTH_SCAN,
                Manifest.permission.BLUETOOTH_ADVERTISE,
                Manifest.permission.BLUETOOTH_CONNECT
            ))
        } else {
            requiredPermissions.addAll(listOf(
                Manifest.permission.BLUETOOTH,
                Manifest.permission.BLUETOOTH_ADMIN,
                Manifest.permission.ACCESS_FINE_LOCATION
            ))
        }
        
        for (permission in requiredPermissions) {
            val granted = ActivityCompat.checkSelfPermission(context, permission) == PackageManager.PERMISSION_GRANTED
            Log.d(TAG, "Permission $permission: ${if (granted) "GRANTED" else "DENIED"}")
        }
    }
    
    @SuppressLint("MissingPermission")
    private fun setupGattServer() {
        Log.d(TAG, "setupGattServer: Creating callback...")
        val gattServerCallback = object : BluetoothGattServerCallback() {
            override fun onConnectionStateChange(device: BluetoothDevice, status: Int, newState: Int) {
                super.onConnectionStateChange(device, status, newState)
                when (newState) {
                    BluetoothProfile.STATE_CONNECTED -> {
                        Log.d(TAG, "Device connected: ${device.address}")
                    }
                    BluetoothProfile.STATE_DISCONNECTED -> {
                        Log.d(TAG, "Device disconnected: ${device.address}")
                        subscribedCentrals.remove(device.address)
                    }
                }
            }
            
            override fun onCharacteristicWriteRequest(
                device: BluetoothDevice,
                requestId: Int,
                characteristic: BluetoothGattCharacteristic,
                preparedWrite: Boolean,
                responseNeeded: Boolean,
                offset: Int,
                value: ByteArray?
            ) {
                super.onCharacteristicWriteRequest(device, requestId, characteristic, preparedWrite, responseNeeded, offset, value)
                
                if (characteristic.uuid.toString().equals(CHARACTERISTIC_UUID, ignoreCase = true)) {
                    value?.let { data ->
                        Log.d(TAG, "Received ${data.size} bytes from ${device.address}")
                        handleReceivedChunk(data)
                    }
                }
                
                if (responseNeeded) {
                    bluetoothGattServer?.sendResponse(device, requestId, BluetoothGatt.GATT_SUCCESS, offset, null)
                }
            }
            
            override fun onDescriptorWriteRequest(
                device: BluetoothDevice,
                requestId: Int,
                descriptor: BluetoothGattDescriptor,
                preparedWrite: Boolean,
                responseNeeded: Boolean,
                offset: Int,
                value: ByteArray?
            ) {
                super.onDescriptorWriteRequest(device, requestId, descriptor, preparedWrite, responseNeeded, offset, value)
                
                // Handle subscription/unsubscription
                if (descriptor.uuid == UUID.fromString("00002902-0000-1000-8000-00805f9b34fb")) { // Client Characteristic Configuration
                    if (Arrays.equals(value, BluetoothGattDescriptor.ENABLE_NOTIFICATION_VALUE)) {
                        Log.d(TAG, "Central subscribed: ${device.address}")
                        subscribedCentrals[device.address] = device
                    } else if (Arrays.equals(value, BluetoothGattDescriptor.DISABLE_NOTIFICATION_VALUE)) {
                        Log.d(TAG, "Central unsubscribed: ${device.address}")
                        subscribedCentrals.remove(device.address)
                    }
                }
                
                if (responseNeeded) {
                    bluetoothGattServer?.sendResponse(device, requestId, BluetoothGatt.GATT_SUCCESS, offset, null)
                }
            }
        }
        
        Log.d(TAG, "setupGattServer: Opening GATT server...")
        bluetoothGattServer = bluetoothManager.openGattServer(context, gattServerCallback)
        Log.d(TAG, "setupGattServer: GATT server opened: ${bluetoothGattServer != null}")
        
        Log.d(TAG, "setupGattServer: Creating characteristic...")
        // Create the characteristic
        val characteristic = BluetoothGattCharacteristic(
            UUID.fromString(CHARACTERISTIC_UUID),
            BluetoothGattCharacteristic.PROPERTY_READ or
            BluetoothGattCharacteristic.PROPERTY_WRITE or
            BluetoothGattCharacteristic.PROPERTY_WRITE_NO_RESPONSE or
            BluetoothGattCharacteristic.PROPERTY_NOTIFY,
            BluetoothGattCharacteristic.PERMISSION_READ or
            BluetoothGattCharacteristic.PERMISSION_WRITE
        )
        
        // Add Client Characteristic Configuration Descriptor for notifications
        val descriptor = BluetoothGattDescriptor(
            UUID.fromString("00002902-0000-1000-8000-00805f9b34fb"),
            BluetoothGattDescriptor.PERMISSION_READ or BluetoothGattDescriptor.PERMISSION_WRITE
        )
        characteristic.addDescriptor(descriptor)
        Log.d(TAG, "setupGattServer: Descriptor added to characteristic")
        
        // Create the service
        Log.d(TAG, "setupGattServer: Creating service with UUID: $serviceUUID")
        val service = BluetoothGattService(serviceUUID, BluetoothGattService.SERVICE_TYPE_PRIMARY)
        service.addCharacteristic(characteristic)
        Log.d(TAG, "setupGattServer: Characteristic added to service")
        
        // Add service to GATT server
        Log.d(TAG, "setupGattServer: Adding service to GATT server...")
        val result = bluetoothGattServer?.addService(service)
        Log.d(TAG, "setupGattServer: Service added result: $result")
    }
    
    @SuppressLint("MissingPermission")
    private fun startAdvertising() {
        advertiser = bluetoothAdapter.bluetoothLeAdvertiser
        if (advertiser == null) {
            Log.e(TAG, "BLE advertising not supported")
            return
        }
        
        val settings = AdvertiseSettings.Builder()
            .setAdvertiseMode(AdvertiseSettings.ADVERTISE_MODE_BALANCED)
            .setTxPowerLevel(AdvertiseSettings.ADVERTISE_TX_POWER_MEDIUM)
            .setConnectable(true)
            .build()
            
        val data = AdvertiseData.Builder()
            .setIncludeDeviceName(false)
            .setIncludeTxPowerLevel(false)
            .addServiceUuid(ParcelUuid(serviceUUID))
            .build()
            
        // Add device ID as service data (since setLocalName doesn't exist in Android)
        val deviceIdBytes = deviceId.toByteArray(Charsets.UTF_8)
        val scanResponse = AdvertiseData.Builder()
            .setIncludeDeviceName(false)
            .setIncludeTxPowerLevel(false)
            .addServiceData(ParcelUuid(serviceUUID), deviceIdBytes)
            .build()
            
        advertiseCallback = object : AdvertiseCallback() {
            override fun onStartSuccess(settingsInEffect: AdvertiseSettings) {
                super.onStartSuccess(settingsInEffect)
                Log.d(TAG, "Advertising started as $deviceId")
            }
            
            override fun onStartFailure(errorCode: Int) {
                super.onStartFailure(errorCode)
                Log.e(TAG, "Advertising failed with error: $errorCode")
            }
        }
        
        advertiser?.startAdvertising(settings, data, scanResponse, advertiseCallback)
    }
    
    @SuppressLint("MissingPermission")
    private fun startScanning() {
        bluetoothLeScanner = bluetoothAdapter.bluetoothLeScanner
        if (bluetoothLeScanner == null) {
            Log.e(TAG, "BLE scanning not supported")
            return
        }
        
        val scanSettings = ScanSettings.Builder()
            .setScanMode(ScanSettings.SCAN_MODE_BALANCED)
            .build()
            
        val scanFilters = listOf(
            ScanFilter.Builder()
                .setServiceUuid(ParcelUuid(serviceUUID))
                .build()
        )
        
        scanCallback = object : ScanCallback() {
            override fun onScanResult(callbackType: Int, result: ScanResult) {
                super.onScanResult(callbackType, result)
                
                val device = result.device
                val scanRecord = result.scanRecord
                
                // Get device ID from service data (Android approach)
                val advertisedDeviceId = scanRecord?.getServiceData(ParcelUuid(serviceUUID))?.let { data ->
                    String(data, Charsets.UTF_8)
                } ?: scanRecord?.deviceName ?: device.name ?: "Unknown"
                
                // Skip self-advertisements (same as Darwin implementation)
                if (advertisedDeviceId == deviceId) {
                    Log.d(TAG, "Ignoring self-advertisement")
                    return
                }
                
                // Skip devices that are already connected
                if (connectedDevices.containsKey(device.address)) {
                    return
                }
                
                // Connect to any device advertising our service UUID
                Log.d(TAG, "Discovered device: ${device.address}, deviceId: $advertisedDeviceId")
                connectToDevice(device)
            }
            
            override fun onScanFailed(errorCode: Int) {
                super.onScanFailed(errorCode)
                Log.e(TAG, "Scan failed with error: $errorCode")
            }
        }
        
        bluetoothLeScanner?.startScan(scanFilters, scanSettings, scanCallback)
        Log.d(TAG, "Scanning for $serviceUUID")
    }
    
    @SuppressLint("MissingPermission")
    private fun connectToDevice(device: BluetoothDevice) {
        val gattCallback = object : BluetoothGattCallback() {
            override fun onConnectionStateChange(gatt: BluetoothGatt, status: Int, newState: Int) {
                super.onConnectionStateChange(gatt, status, newState)
                
                when (newState) {
                    BluetoothProfile.STATE_CONNECTED -> {
                        Log.d(TAG, "Connected to ${device.address}")
                        connectedDevices[device.address] = gatt
                        gatt.discoverServices()
                    }
                    BluetoothProfile.STATE_DISCONNECTED -> {
                        Log.d(TAG, "Disconnected from ${device.address}")
                        connectedDevices.remove(device.address)
                        gatt.close()
                    }
                }
            }
            
            override fun onServicesDiscovered(gatt: BluetoothGatt, status: Int) {
                super.onServicesDiscovered(gatt, status)
                
                if (status == BluetoothGatt.GATT_SUCCESS) {
                    val service = gatt.getService(serviceUUID)
                    service?.let { svc ->
                        val characteristic = svc.getCharacteristic(UUID.fromString(CHARACTERISTIC_UUID))
                        characteristic?.let { char ->
                            // Enable notifications
                            gatt.setCharacteristicNotification(char, true)
                            
                            // Write to descriptor to enable notifications
                            val descriptor = char.getDescriptor(UUID.fromString("00002902-0000-1000-8000-00805f9b34fb"))
                            descriptor?.let { desc ->
                                desc.value = BluetoothGattDescriptor.ENABLE_NOTIFICATION_VALUE
                                gatt.writeDescriptor(desc)
                            }
                        }
                    }
                }
            }
            
            override fun onCharacteristicChanged(gatt: BluetoothGatt, characteristic: BluetoothGattCharacteristic) {
                super.onCharacteristicChanged(gatt, characteristic)
                
                if (characteristic.uuid.toString().equals(CHARACTERISTIC_UUID, ignoreCase = true)) {
                    val data = characteristic.value
                    Log.d(TAG, "Received ${data.size} bytes from ${gatt.device.address}")
                    handleReceivedChunk(data)
                }
            }
        }
        
        device.connectGatt(context, false, gattCallback)
    }
    
    fun sendData(data: ByteArray) {
        if (!isStarted) {
            Log.w(TAG, "BLE not started, cannot send data")
            return
        }
        
        if (subscribedCentrals.isEmpty()) {
            Log.w(TAG, "No subscribers, cannot send data")
            return
        }
        
        CoroutineScope(Dispatchers.IO).launch {
            publishChunked(data)
        }
    }
    
    @SuppressLint("MissingPermission")
    private suspend fun publishChunked(payload: ByteArray) {
        // Generate message ID (2 bytes, same as Go implementation)
        val msgId = ByteArray(2)
        Random.Default.nextBytes(msgId)
        
        val totalChunks = (payload.size + MAX_CHUNK_SIZE - 1) / MAX_CHUNK_SIZE
        Log.d(TAG, "Sending ${payload.size} bytes in $totalChunks chunks")
        
        for (i in 0 until totalChunks) {
            val start = i * MAX_CHUNK_SIZE
            val end = minOf(start + MAX_CHUNK_SIZE, payload.size)
            val chunkData = payload.sliceArray(start until end)
            
            // Create chunk header (4 bytes total)
            val buffer = ByteBuffer.allocate(4 + chunkData.size).order(ByteOrder.BIG_ENDIAN)
            
            // msgID (2 bytes)
            buffer.put(msgId)
            
            // seq (2 bytes) - sequence number with last bit as LAST flag
            var seq = (i shl 1).toShort()
            if (i == totalChunks - 1) {
                seq = (seq.toInt() or 1).toShort()
            }
            buffer.putShort(seq)
            
            // data
            buffer.put(chunkData)
            
            val chunkPacket = buffer.array()
            
            // Send to all subscribed centrals
            val service = bluetoothGattServer?.getService(serviceUUID)
            val characteristic = service?.getCharacteristic(UUID.fromString(CHARACTERISTIC_UUID))
            
            characteristic?.let { char ->
                char.value = chunkPacket
                for (device in subscribedCentrals.values) {
                    bluetoothGattServer?.notifyCharacteristicChanged(device, char, false)
                }
            }
            
            Log.d(TAG, "Sent chunk $i/${totalChunks - 1} (${chunkData.size} bytes)")
            delay(20) // Small delay between chunks (same as Go implementation)
        }
    }
    
    private fun handleReceivedChunk(chunk: ByteArray) {
        if (chunk.size < 4) {
            Log.w(TAG, "Received chunk too small: ${chunk.size} bytes")
            return
        }
        
        val buffer = ByteBuffer.wrap(chunk).order(ByteOrder.BIG_ENDIAN)
        
        // Parse header
        val msgId = ByteArray(2)
        buffer.get(msgId)
        
        val seqRaw = buffer.getShort().toInt() and 0xFFFF
        val seqIndex = seqRaw shr 1
        val isLast = (seqRaw and 1) == 1
        
        val chunkData = ByteArray(buffer.remaining())
        buffer.get(chunkData)
        
        val msgKey = msgId.contentToString()
        
        Log.d(TAG, "Received chunk seq=$seqIndex, isLast=$isLast, size=${chunkData.size}")
        
        synchronized(assemblyBuffers) {
            val chunkBuffer = assemblyBuffers.getOrPut(msgKey) { ChunkBuffer() }
            
            if (isLast) {
                chunkBuffer.total = seqIndex + 1
                Log.d(TAG, "Message has ${chunkBuffer.total} total chunks")
            }
            
            chunkBuffer.chunks[seqIndex] = chunkData
            chunkBuffer.received.add(seqIndex)
            
            // Check if message is complete
            if (chunkBuffer.total > 0 && chunkBuffer.received.size == chunkBuffer.total) {
                // Reassemble message
                val fullMessage = ByteArray(chunkBuffer.chunks.values.sumOf { it.size })
                var offset = 0
                
                for (i in 0 until chunkBuffer.total) {
                    val chunk = chunkBuffer.chunks[i]!!
                    System.arraycopy(chunk, 0, fullMessage, offset, chunk.size)
                    offset += chunk.size
                }
                
                assemblyBuffers.remove(msgKey)
                
                // Store message and trigger callback
                synchronized(messageLock) {
                    lastMessage = fullMessage
                }
                
                Log.d(TAG, "Message reassembled: ${fullMessage.size} bytes")
                onMessageCallback?.invoke()
            }
        }
    }
    
    fun setOnMessageCallback(callback: () -> Unit) {
        onMessageCallback = callback
    }
    
    fun getLastMessage(): Triple<String, String, ByteArray>? {
        synchronized(messageLock) {
            lastMessage?.let { msg ->
                return try {
                    val decoded = MqttCodec.decodeMessage(msg)
                    Triple(decoded.filename, decoded.type, decoded.payload)
                } catch (e: Exception) {
                    Log.e(TAG, "Failed to decode BLE message: ${e.message}", e)
                    null
                }
            }
            return null
        }
    }
    
    fun clearMessage() {
        synchronized(messageLock) {
            lastMessage = null
        }
        synchronized(assemblyBuffers) {
            assemblyBuffers.clear()
        }
    }
    
    fun isStarted(): Boolean = isStarted
}