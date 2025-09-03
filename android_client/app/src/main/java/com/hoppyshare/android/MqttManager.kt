package com.hoppyshare.android

import android.content.Context
import android.util.Log

object MqttManager {
    private var mqttClient: MqttClient? = null
    private var isInitialized = false
    
    fun getInstance(context: Context): MqttClient {
        if (mqttClient == null || !isInitialized) {
            Log.d("MqttManager", "Creating new MQTT client instance")
            mqttClient = MqttClient(context.applicationContext)
            isInitialized = true
        } else {
            Log.d("MqttManager", "Reusing existing MQTT client instance")
        }
        return mqttClient!!
    }
    
    suspend fun ensureConnected(context: Context): Boolean {
        return try {
            val client = getInstance(context)
            if (!client.isConnected()) {
                Log.d("MqttManager", "MQTT not connected, reconnecting...")
                val clientId = client.connect()
                if (clientId != null) {
                    Log.d("MqttManager", "Reconnected successfully as $clientId")
                    true
                } else {
                    Log.e("MqttManager", "Failed to reconnect")
                    false
                }
            } else {
                Log.d("MqttManager", "MQTT already connected")
                true
            }
        } catch (e: Exception) {
            Log.e("MqttManager", "Error ensuring connection: ${e.message}", e)
            false
        }
    }
}