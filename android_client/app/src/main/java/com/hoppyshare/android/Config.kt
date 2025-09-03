package com.hoppyshare.android

import android.content.Context
import android.util.Base64
import android.util.Log
import java.security.interfaces.RSAPrivateKey
import javax.crypto.spec.OAEPParameterSpec

object Config {
    private const val PREFS_NAME = "hoppyshare_certs"
    
    var deviceId: String = ""
    var certPem: String = ""
    var keyPem: String = ""
    var caPem: String = ""
    var groupKey: String = ""
    
    fun loadFromPreferences(context: Context): Boolean {
        return try {
            val prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
            
            if (!prefs.getBoolean("is_configured", false)) {
                return false
            }
            
            deviceId = prefs.getString("device_id", "") ?: ""
            certPem = prefs.getString("client_cert", "") ?: ""
            keyPem = prefs.getString("client_key", "") ?: ""
            caPem = prefs.getString("ca_cert", "") ?: ""
            groupKey = prefs.getString("group_key", "") ?: ""
            
            if (deviceId.isEmpty() || certPem.isEmpty() || keyPem.isEmpty() || 
                caPem.isEmpty() || groupKey.isEmpty()) {
                return false
            }
            
            Log.d("Config", "Successfully loaded config from preferences")
            true
        } catch (e: Exception) {
            Log.e("Config", "Failed to load config from preferences: ${e.message}")
            false
        }
    }
    
    fun isConfigured(): Boolean {
        return deviceId.isNotEmpty() && 
               certPem.isNotEmpty() && 
               keyPem.isNotEmpty() && 
               caPem.isNotEmpty() && 
               groupKey.isNotEmpty()
    }
    
    fun getGroupKeyBytes(): ByteArray {
        Log.d("Config", "getGroupKeyBytes: group key string length = ${groupKey.length}")
        Log.d("Config", "getGroupKeyBytes: group key = ${groupKey.take(50)}...") // Show first 50 chars
        
        return try {
            // Based on the backend code, for Android the group_key is sent as encrypted_group_key.hex()
            // This means it's a hex-encoded encrypted group key that needs RSA decryption
            val encryptedKeyBytes = hexStringToByteArray(groupKey)
            Log.d("Config", "getGroupKeyBytes: hex decoded size = ${encryptedKeyBytes.size}")
            
            val decryptedKey = decryptGroupKey(encryptedKeyBytes, keyPem)
            Log.d("Config", "getGroupKeyBytes: decrypted key size = ${decryptedKey.size}")
            decryptedKey
        } catch (e: Exception) {
            Log.e("Config", "Failed to decrypt group key, trying as direct hex: ${e.message}")
            // Fallback: maybe it's just a hex-encoded raw key
            try {
                val directKey = hexStringToByteArray(groupKey)
                Log.d("Config", "getGroupKeyBytes: direct hex key size = ${directKey.size}")
                directKey
            } catch (e2: Exception) {
                Log.e("Config", "All group key parsing methods failed", e2)
                throw e2
            }
        }
    }
    
    private fun hexStringToByteArray(s: String): ByteArray {
        val len = s.length
        val data = ByteArray(len / 2)
        var i = 0
        while (i < len) {
            data[i / 2] = ((Character.digit(s[i], 16) shl 4) + Character.digit(s[i + 1], 16)).toByte()
            i += 2
        }
        return data
    }
    
    private fun decryptGroupKey(encryptedKey: ByteArray, privateKeyPem: String): ByteArray {
        try {
            // Parse the private key
            val keyContent = privateKeyPem
                .replace("-----BEGIN PRIVATE KEY-----", "")
                .replace("-----END PRIVATE KEY-----", "")
                .replace("-----BEGIN RSA PRIVATE KEY-----", "")
                .replace("-----END RSA PRIVATE KEY-----", "")
                .replace("\n", "")
                .replace("\r", "")
                .trim()
            
            val keyBytes = Base64.decode(keyContent, Base64.DEFAULT)
            
            // Try PKCS8 format first
            val keySpec = java.security.spec.PKCS8EncodedKeySpec(keyBytes)
            val keyFactory = java.security.KeyFactory.getInstance("RSA")
            val privateKey = keyFactory.generatePrivate(keySpec) as java.security.interfaces.RSAPrivateKey
            
            // Decrypt using RSA-OAEP
            val cipher = javax.crypto.Cipher.getInstance("RSA/ECB/OAEPPadding")
            val oaepParams = javax.crypto.spec.OAEPParameterSpec(
                "SHA-256", "MGF1",
                java.security.spec.MGF1ParameterSpec.SHA256,
                javax.crypto.spec.PSource.PSpecified.DEFAULT
            )
            cipher.init(javax.crypto.Cipher.DECRYPT_MODE, privateKey, oaepParams)
            
            return cipher.doFinal(encryptedKey)
        } catch (e: Exception) {
            android.util.Log.e("Config", "Failed to decrypt group key: ${e.message}", e)
            throw e
        }
    }
}