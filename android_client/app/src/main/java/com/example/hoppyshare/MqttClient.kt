package com.example.hoppyshare

import android.content.Context
import android.util.Log
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import org.eclipse.paho.client.mqttv3.*
import org.eclipse.paho.client.mqttv3.persist.MemoryPersistence
import java.io.ByteArrayInputStream
import java.security.KeyStore
import java.security.cert.CertificateFactory
import java.security.cert.X509Certificate
import java.security.spec.PKCS8EncodedKeySpec
import java.util.*
import javax.net.ssl.KeyManagerFactory
import javax.net.ssl.SSLContext
import javax.net.ssl.TrustManagerFactory
import java.security.KeyFactory
import java.security.PrivateKey
import android.util.Base64
import android.widget.Toast

class MqttClient(private val context: Context) {
    private var client: IMqttClient? = null
    private var clientId: String = ""

    private val lastMessageMutex = Mutex()
    private var lastMessage: LastMessage? = null

    private var onMessageCallback: (() -> Unit)? = null

    data class LastMessage(
        val filename: String,
        val contentType: String,
        val payload: ByteArray
    ) {
        override fun equals(other: Any?): Boolean {
            if (this === other) return true
            if (javaClass != other?.javaClass) return false

            other as LastMessage

            if (filename != other.filename) return false
            if (contentType != other.contentType) return false
            if (!payload.contentEquals(other.payload)) return false

            return true
        }

        override fun hashCode(): Int {
            var result = filename.hashCode()
            result = 31 * result + contentType.hashCode()
            result = 31 * result + payload.contentHashCode()
            return result
        }
    }

    suspend fun connect(): String? {
        if (!Config.loadFromPreferences(context)) {
            Log.e("MqttClient", "Config not loaded, cannot connect")
            return null
        }

        return try {
            val sslContext = createSSLContext()

            // Extract client ID from certificate
            val certFactory = CertificateFactory.getInstance("X.509")
            val cert = certFactory.generateCertificate(
                ByteArrayInputStream(Config.certPem.toByteArray())
            ) as X509Certificate
            clientId = cert.subjectX500Principal.name.substringAfter("CN=").substringBefore(",")
            Log.d("MqttClient", "Base client ID from cert: $clientId")

            val fullClientId = "${clientId}-${System.currentTimeMillis() / 1000}"
            Log.d("MqttClient", "Full client ID: $fullClientId")

            client = org.eclipse.paho.client.mqttv3.MqttClient("ssl://18.188.110.246:8883", fullClientId, MemoryPersistence())
            
            val options = MqttConnectOptions().apply {
                isCleanSession = true
                connectionTimeout = 30
                keepAliveInterval = 60
                isAutomaticReconnect = true
                socketFactory = sslContext.socketFactory
            }
            
            client?.setCallback(object : MqttCallback {
                override fun connectionLost(cause: Throwable?) {
                    Log.w("MqttClient", "Connection lost: ${cause?.message}")
                    // The MQTT client should automatically reconnect due to isAutomaticReconnect = true
                    // But we can add additional handling here if needed
                }
                
                override fun messageArrived(topic: String?, message: MqttMessage?) {
                    message?.let { handleMessage(topic, it) }
                }
                
                override fun deliveryComplete(token: IMqttDeliveryToken?) {
                    Log.d("MqttClient", "Message delivered")
                }
            })
            
            client?.connect(options)

            subscribe()
            
            Log.d("MqttClient", "Connected to MQTT as $clientId")
            clientId
        } catch (e: Exception) {
            Log.e("MqttClient", "Failed to connect: ${e.message}", e)
            null
        }
    }
    
    private fun createSSLContext(): SSLContext {
        val sslContext = SSLContext.getInstance("TLS")
        
        // Load CA certificate for trust store
        val trustStore = KeyStore.getInstance(KeyStore.getDefaultType())
        trustStore.load(null, null)
        
        val certFactory = CertificateFactory.getInstance("X.509")
        val caCert = certFactory.generateCertificate(
            ByteArrayInputStream(Config.caPem.toByteArray())
        )
        trustStore.setCertificateEntry("ca", caCert)
        
        val trustManagerFactory = TrustManagerFactory.getInstance(TrustManagerFactory.getDefaultAlgorithm())
        trustManagerFactory.init(trustStore)
        
        // Load client certificate and key for key store
        val keyStore = KeyStore.getInstance("PKCS12")
        keyStore.load(null, null)
        
        val clientCert = certFactory.generateCertificate(
            ByteArrayInputStream(Config.certPem.toByteArray())
        ) as X509Certificate
        
        // Parse the private key
        val privateKey = parsePrivateKey(Config.keyPem)
        
        keyStore.setKeyEntry(
            "client",
            privateKey,
            "".toCharArray(),
            arrayOf(clientCert)
        )
        
        val keyManagerFactory = KeyManagerFactory.getInstance(KeyManagerFactory.getDefaultAlgorithm())
        keyManagerFactory.init(keyStore, "".toCharArray())
        
        sslContext.init(keyManagerFactory.keyManagers, trustManagerFactory.trustManagers, null)
        return sslContext
    }
    
    private fun parsePrivateKey(keyPem: String): PrivateKey {
        val keyContent = keyPem
            .replace("-----BEGIN PRIVATE KEY-----", "")
            .replace("-----END PRIVATE KEY-----", "")
            .replace("-----BEGIN RSA PRIVATE KEY-----", "")
            .replace("-----END RSA PRIVATE KEY-----", "")
            .replace("\n", "")
            .replace("\r", "")
        
        val keyBytes = Base64.decode(keyContent, Base64.DEFAULT)
        
        // Try PKCS8 first, then PKCS1 if it fails
        return try {
            val keySpec = PKCS8EncodedKeySpec(keyBytes)
            KeyFactory.getInstance("RSA").generatePrivate(keySpec)
        } catch (e: Exception) {
            // If PKCS8 fails, the key might be in PKCS1 format
            // We would need to convert it, but for now assume PKCS8
            throw IllegalArgumentException("Unsupported private key format", e)
        }
    }
    
    private fun subscribe() {
        client?.let { mqttClient ->
            val notesTopic = "users/$clientId/notes"
            val settingsTopic = "users/$clientId/settings"
            
            Log.d("MqttClient", "Subscribing to $notesTopic and $settingsTopic")
            
            try {
                mqttClient.subscribe(notesTopic, 1) { topic, message ->
                    Log.d("MqttClient", "Received message on $topic")
                    handleNotesMessage(message)
                }
                
                mqttClient.subscribe(settingsTopic, 1) { topic, message ->
                    Log.d("MqttClient", "Received settings message on $topic")
                    handleSettingsMessage(message)
                }
            } catch (e: Exception) {
                Log.e("MqttClient", "Failed to subscribe: ${e.message}", e)
            }
        }
    }
    
    private fun handleMessage(topic: String?, message: MqttMessage) {
        when {
            topic?.contains("/notes") == true -> handleNotesMessage(message)
            topic?.contains("/settings") == true -> handleSettingsMessage(message)
        }
    }
    
    private fun handleNotesMessage(message: MqttMessage) {
        Log.d("MqttClient", "NEW MESSAGE")
        try {
            val decoded = MqttCodec.decodeMessage(message.payload)
            
            // Check if message is from self (filter based on send_to_self setting)
            val ownDeviceHash = MqttCodec.hashDeviceId(Config.deviceId)
            if (decoded.deviceId.contentEquals(ownDeviceHash)) {
                val settings = Settings.getCurrentSettings()
                if (!settings.sendToSelf) {
                    Log.d("MqttClient", "Ignoring message from self (send_to_self = false)")
                    return
                } else {
                    Log.d("MqttClient", "Processing message from self (send_to_self = true)")
                }
            }
            
            CoroutineScope(Dispatchers.Main).launch {
                cacheMessage(decoded.filename, decoded.type, decoded.payload)
            }

        } catch (e: Exception) {
            Log.e("MqttClient", "Failed to decode message: ${e.message}", e)
        }
    }
    
    private fun handleSettingsMessage(message: MqttMessage) {
        val settingsData = String(message.payload)
        Log.d("MqttClient", "Settings message: $settingsData")
        
        try {
            // Parse settings using the device ID from Config
            val deviceId = Config.deviceId
            if (deviceId.isNotEmpty()) {
                val updated = Settings.parseSettingsFromMqtt(settingsData, deviceId)
                if (updated) {
                    Log.d("MqttClient", "Settings updated successfully")
                    // Settings have been updated, the app should respond to these changes
                    // The Settings object will handle persistence automatically
                } else {
                    Log.d("MqttClient", "No settings changes for this device")
                }
            } else {
                Log.w("MqttClient", "Device ID not available, cannot parse settings")
            }
        } catch (e: Exception) {
            Log.e("MqttClient", "Failed to handle settings message: ${e.message}", e)
        }
    }
    
    private suspend fun cacheMessage(filename: String, contentType: String, payload: ByteArray) {
        lastMessageMutex.withLock {
            lastMessage = LastMessage(filename, contentType, payload)
        }
        
        onMessageCallback?.invoke()
    }
    
    suspend fun getLastMessage(): LastMessage? {
        return lastMessageMutex.withLock {
            lastMessage
        }
    }
    
    suspend fun clearLastMessage() {
        lastMessageMutex.withLock {
            lastMessage = null
        }
    }
    
    fun setOnMessageCallback(callback: () -> Unit) {
        onMessageCallback = callback
    }
    
    suspend fun publish(data: ByteArray, contentType: String, filename: String): Boolean {
        val mqttClient = client
        if (mqttClient == null) {
            Log.e("MqttClient", "Cannot publish: client is null")
            return false
        }
        
        if (!mqttClient.isConnected) {
            Log.e("MqttClient", "Cannot publish: client not connected")
            return false
        }
        
        return try {
            val encoded = MqttCodec.encodeMessage(contentType, filename, Config.deviceId, data)
            
            val message = MqttMessage(encoded).apply {
                qos = 1
                isRetained = false
            }

            val topic = "users/$clientId/notes"
            mqttClient.publish(topic, message)
            true
        } catch (e: Exception) {
            Log.e("MqttClient", "Failed to publish: ${e.message}", e)
            false
        }
    }
    
    fun disconnect() {
        try {
            client?.disconnect()
            Log.d("MqttClient", "Disconnected from MQTT")
        } catch (e: Exception) {
            Log.e("MqttClient", "Error during disconnect: ${e.message}", e)
        }
    }
    
    fun isConnected(): Boolean {
        return client?.isConnected ?: false
    }
}