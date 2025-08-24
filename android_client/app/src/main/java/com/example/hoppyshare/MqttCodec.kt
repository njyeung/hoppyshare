package com.example.hoppyshare

import android.util.Base64
import java.io.ByteArrayInputStream
import java.io.ByteArrayOutputStream
import java.security.MessageDigest
import java.security.SecureRandom
import javax.crypto.Cipher
import javax.crypto.spec.GCMParameterSpec
import javax.crypto.spec.SecretKeySpec

data class DecodedPayload(
    val type: String,
    val filename: String,
    val deviceId: ByteArray,
    val payload: ByteArray
) {
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (javaClass != other?.javaClass) return false

        other as DecodedPayload

        if (type != other.type) return false
        if (filename != other.filename) return false
        if (!deviceId.contentEquals(other.deviceId)) return false
        if (!payload.contentEquals(other.payload)) return false

        return true
    }

    override fun hashCode(): Int {
        var result = type.hashCode()
        result = 31 * result + filename.hashCode()
        result = 31 * result + deviceId.contentHashCode()
        result = 31 * result + payload.contentHashCode()
        return result
    }
}

object MqttCodec {
    
    fun encodeMessage(mimeType: String, filename: String, deviceId: String, payload: ByteArray): ByteArray {
        android.util.Log.d("MqttCodec", "encodeMessage() called:")
        android.util.Log.d("MqttCodec", "  mimeType: $mimeType (length: ${mimeType.length})")
        android.util.Log.d("MqttCodec", "  filename: $filename (length: ${filename.length})")
        android.util.Log.d("MqttCodec", "  deviceId: $deviceId")
        android.util.Log.d("MqttCodec", "  payload size: ${payload.size}")
        
        val groupKey = Config.getGroupKeyBytes()
        android.util.Log.d("MqttCodec", "  groupKey size: ${groupKey.size}")
        
        if (mimeType.length > 255 || filename.length > 255) {
            throw IllegalArgumentException("MIME type or filename too long")
        }
        
        val buffer = ByteArrayOutputStream()
        
        // Write header
        buffer.write(mimeType.length)
        buffer.write(mimeType.toByteArray())
        
        buffer.write(filename.length)
        buffer.write(filename.toByteArray())
        
        val hashedId = hashDeviceId(deviceId)
        android.util.Log.d("MqttCodec", "  hashed device ID size: ${hashedId.size}")
        buffer.write(hashedId)
        
        val header = buffer.toByteArray()
        android.util.Log.d("MqttCodec", "  header size: ${header.size}")
        
        // Encrypt payload with AES-GCM
        android.util.Log.d("MqttCodec", "Encrypting payload...")
        val (nonce, ciphertext) = encryptAESGCM(groupKey, payload, header)
        android.util.Log.d("MqttCodec", "  nonce size: ${nonce.size}")
        android.util.Log.d("MqttCodec", "  ciphertext size: ${ciphertext.size}")
        
        buffer.write(nonce)
        buffer.write(ciphertext)
        
        val result = buffer.toByteArray()
        android.util.Log.d("MqttCodec", "Final encoded message size: ${result.size}")
        return result
    }
    
    fun decodeMessage(data: ByteArray): DecodedPayload {
        val input = ByteArrayInputStream(data)
        
        // Read header
        val typeLen = input.read()
        val typeBytes = ByteArray(typeLen)
        input.read(typeBytes)
        
        val nameLen = input.read()
        val nameBytes = ByteArray(nameLen)
        input.read(nameBytes)
        
        val devId = ByteArray(32)
        input.read(devId)
        
        val headerLen = 1 + typeBytes.size + 1 + nameBytes.size + 32
        val header = data.sliceArray(0 until headerLen)
        
        // Read nonce and ciphertext
        val nonce = ByteArray(12)
        input.read(nonce)
        
        val ciphertext = ByteArray(input.available())
        input.read(ciphertext)
        
        val groupKey = Config.getGroupKeyBytes()
        val plaintext = decryptAESGCM(groupKey, nonce, ciphertext, header)
        
        return DecodedPayload(
            type = String(typeBytes),
            filename = String(nameBytes),
            deviceId = devId,
            payload = plaintext
        )
    }
    
    private fun encryptAESGCM(key: ByteArray, plaintext: ByteArray, aad: ByteArray): Pair<ByteArray, ByteArray> {
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        val secretKey = SecretKeySpec(key, "AES")
        
        // Generate random nonce
        val nonce = ByteArray(12)
        SecureRandom().nextBytes(nonce)
        
        val gcmSpec = GCMParameterSpec(128, nonce)
        cipher.init(Cipher.ENCRYPT_MODE, secretKey, gcmSpec)
        
        // Add AAD (Additional Authenticated Data)
        cipher.updateAAD(aad)
        
        val ciphertext = cipher.doFinal(plaintext)
        return Pair(nonce, ciphertext)
    }
    
    private fun decryptAESGCM(key: ByteArray, nonce: ByteArray, ciphertext: ByteArray, aad: ByteArray): ByteArray {
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        val secretKey = SecretKeySpec(key, "AES")
        val gcmSpec = GCMParameterSpec(128, nonce)
        
        cipher.init(Cipher.DECRYPT_MODE, secretKey, gcmSpec)
        cipher.updateAAD(aad)
        
        return cipher.doFinal(ciphertext)
    }
    
    fun hashDeviceId(id: String): ByteArray {
        val digest = MessageDigest.getInstance("SHA-256")
        return digest.digest(id.toByteArray())
    }
}