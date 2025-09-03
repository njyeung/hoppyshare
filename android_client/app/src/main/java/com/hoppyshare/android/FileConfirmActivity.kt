package com.hoppyshare.android

import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.widget.Button
import android.widget.ImageView
import android.widget.TextView
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import kotlinx.coroutines.launch
import java.io.ByteArrayOutputStream

class FileConfirmActivity : AppCompatActivity() {
    
    private lateinit var sharePreview: ImageView
    private lateinit var shareFileName: TextView
    private lateinit var shareFileSize: TextView
    private lateinit var mqttClient: MqttClient
    private var currentFileUri: Uri? = null
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_share_receive)
        
        mqttClient = MqttManager.getInstance(this)
        setupUI()
        handleFileUri()
        connectToMqtt()
    }
    
    private fun setupUI() {
        sharePreview = findViewById(R.id.sharePreview)
        shareFileName = findViewById(R.id.shareFileName)
        shareFileSize = findViewById(R.id.shareFileSize)
        
        // Update title for file picker context
        val titleText = findViewById<TextView>(R.id.titleText)
        titleText.text = "Send via HoppyShare"
        
        val cancelButton = findViewById<Button>(R.id.cancelButton)
        val sendButton = findViewById<Button>(R.id.sendButton)
        
        cancelButton.setOnClickListener {
            finish()
        }
        
        sendButton.setOnClickListener {
            sendFileViaMqtt()
        }
    }
    
    private fun handleFileUri() {
        val uriString = intent.getStringExtra("file_uri")
        if (uriString != null) {
            val uri = Uri.parse(uriString)
            currentFileUri = uri
            val fileName = getFileName(uri)
            val fileSize = getFileSize(uri)
            
            showFilePreview(uri, fileName, fileSize)
        } else {
            Toast.makeText(this, "No file selected", Toast.LENGTH_SHORT).show()
            finish()
        }
    }
    
    private fun connectToMqtt() {
        lifecycleScope.launch {
            MqttManager.ensureConnected(this@FileConfirmActivity)
        }
    }
    
    private fun sendFileViaMqtt() {
        val uri = currentFileUri ?: return
        val fileName = getFileName(uri)
        val mimeType = contentResolver.getType(uri) ?: "application/octet-stream"
        
        android.util.Log.d("FileConfirm", "Starting file send:")
        android.util.Log.d("FileConfirm", "  fileName: $fileName")
        android.util.Log.d("FileConfirm", "  mimeType: $mimeType")
        
        lifecycleScope.launch {
            try {
                // Load config first
                if (!Config.loadFromPreferences(this@FileConfirmActivity)) {
                    Toast.makeText(this@FileConfirmActivity, "Config not loaded", Toast.LENGTH_LONG).show()
                    android.util.Log.e("FileConfirm", "Config not loaded")
                    return@launch
                }
                
                android.util.Log.d("FileConfirm", "Config loaded, deviceId: ${Config.deviceId}")
                
                if (!MqttManager.ensureConnected(this@FileConfirmActivity)) {
                    Toast.makeText(this@FileConfirmActivity, "Failed to connect to MQTT", Toast.LENGTH_LONG).show()
                    android.util.Log.e("FileConfirm", "Failed to connect to MQTT")
                    return@launch
                }
                
                // Read file content
                android.util.Log.d("FileConfirm", "Reading file data...")
                val fileData = readFileData(uri)
                if (fileData == null) {
                    Toast.makeText(this@FileConfirmActivity, "Failed to read file", Toast.LENGTH_LONG).show()
                    android.util.Log.e("FileConfirm", "Failed to read file data")
                    return@launch
                }
                
                // Check 25MB size limit
                if (fileData.size > 25 * 1024 * 1024) {
                    android.util.Log.e("FileConfirm", "File too large: ${fileData.size} bytes")
                    MainActivity.showError(this@FileConfirmActivity, "File too large (>25MB). Operation cancelled.")
                    finish()
                    return@launch
                }
                
                android.util.Log.d("FileConfirm", "File data read successfully, size: ${fileData.size} bytes")

                val success = mqttClient.publish(fileData, mimeType, fileName)
                
                android.util.Log.d("FileConfirm", "Publish result: $success")
                
                if (success) {
                    Toast.makeText(this@FileConfirmActivity, "File sent successfully!", Toast.LENGTH_SHORT).show()
                    finish()
                } else {
                    Toast.makeText(this@FileConfirmActivity, "Failed to send file via MQTT", Toast.LENGTH_LONG).show()
                }
                
            } catch (e: Exception) {
                android.util.Log.e("FileConfirm", "Error sending file: ${e.message}", e)
                Toast.makeText(this@FileConfirmActivity, "Error: ${e.message}", Toast.LENGTH_LONG).show()
            }
        }
    }
    
    private suspend fun readFileData(uri: Uri): ByteArray? {
        return try {
            contentResolver.openInputStream(uri)?.use { inputStream ->
                val buffer = ByteArrayOutputStream()
                val data = ByteArray(1024)
                var nRead: Int
                while (inputStream.read(data, 0, data.size).also { nRead = it } != -1) {
                    buffer.write(data, 0, nRead)
                }
                buffer.toByteArray()
            }
        } catch (e: Exception) {
            null
        }
    }
    
    private fun showFilePreview(uri: Uri, fileName: String, fileSize: Long) {
        shareFileName.text = fileName
        shareFileSize.text = formatFileSize(fileSize)
        
        // Try to show thumbnail for images
        val mimeType = contentResolver.getType(uri)
        if (mimeType?.startsWith("image/") == true) {
            try {
                sharePreview.setImageURI(uri)
            } catch (e: Exception) {
                showFileIcon(fileName)
            }
        } else {
            showFileIcon(fileName)
        }
    }
    
    private fun showFileIcon(fileName: String) {
        val extension = fileName.substringAfterLast('.', "").lowercase()
        val color = when (extension) {
            "pdf" -> android.graphics.Color.RED
            "doc", "docx" -> android.graphics.Color.BLUE
            "txt" -> android.graphics.Color.GREEN
            "zip", "rar" -> android.graphics.Color.YELLOW
            else -> android.graphics.Color.GRAY
        }
        
        sharePreview.setBackgroundColor(color)
        sharePreview.setImageDrawable(null)
    }
    
    private fun getFileName(uri: Uri): String {
        var fileName = "Unknown file"
        contentResolver.query(uri, null, null, null, null)?.use { cursor ->
            val nameIndex = cursor.getColumnIndex(android.provider.OpenableColumns.DISPLAY_NAME)
            if (nameIndex != -1 && cursor.moveToFirst()) {
                fileName = cursor.getString(nameIndex) ?: fileName
            }
        }
        return fileName
    }
    
    private fun getFileSize(uri: Uri): Long {
        var fileSize = 0L
        contentResolver.query(uri, null, null, null, null)?.use { cursor ->
            val sizeIndex = cursor.getColumnIndex(android.provider.OpenableColumns.SIZE)
            if (sizeIndex != -1 && cursor.moveToFirst()) {
                fileSize = cursor.getLong(sizeIndex)
            }
        }
        return fileSize
    }
    
    private fun formatFileSize(bytes: Long): String {
        return when {
            bytes >= 1024 * 1024 -> "${bytes / (1024 * 1024)}MB"
            bytes >= 1024 -> "${bytes / 1024}KB" 
            else -> "${bytes}B"
        }
    }
}