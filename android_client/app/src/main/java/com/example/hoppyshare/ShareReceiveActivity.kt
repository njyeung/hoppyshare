package com.example.hoppyshare

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

class ShareReceiveActivity : AppCompatActivity() {
    
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
        connectToMqtt()
        
        when (intent?.action) {
            Intent.ACTION_SEND -> handleSingleFile()
            Intent.ACTION_SEND_MULTIPLE -> handleMultipleFiles()
            else -> {
                Toast.makeText(this, "Unsupported action", Toast.LENGTH_SHORT).show()
                finish()
            }
        }
    }
    
    private fun setupUI() {
        sharePreview = findViewById(R.id.sharePreview)
        shareFileName = findViewById(R.id.shareFileName)
        shareFileSize = findViewById(R.id.shareFileSize)
        
        val cancelButton = findViewById<Button>(R.id.cancelButton)
        val sendButton = findViewById<Button>(R.id.sendButton)
        
        cancelButton.setOnClickListener {
            finish()
        }
        
        sendButton.setOnClickListener {
            sendFileViaMqtt()
        }
    }
    
    private fun handleSingleFile() {
        val uri = intent.getParcelableExtra<Uri>(Intent.EXTRA_STREAM)
        if (uri != null) {
            currentFileUri = uri
            val fileName = getFileName(uri)
            val fileSize = getFileSize(uri)
            
            // Show file preview
            showFilePreview(uri, fileName, fileSize)
            
        } else {
            // Handle text sharing
            val text = intent.getStringExtra(Intent.EXTRA_TEXT)
            if (text != null) {
                handleSharedText(text)
            } else {
                Toast.makeText(this, "No file received", Toast.LENGTH_SHORT).show()
                finish()
            }
        }
    }
    
    private fun handleMultipleFiles() {
        val uris = intent.getParcelableArrayListExtra<Uri>(Intent.EXTRA_STREAM)
        if (!uris.isNullOrEmpty()) {
            // For now, just show the first file (we can enhance this later)
            val firstUri = uris[0]
            val fileName = "${uris.size} files selected"
            val fileSize = getFileSize(firstUri)
            
            showFilePreview(firstUri, fileName, fileSize)
            
        } else {
            Toast.makeText(this, "No files received", Toast.LENGTH_SHORT).show()
            finish()
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
    
    private fun connectToMqtt() {
        lifecycleScope.launch {
            MqttManager.ensureConnected(this@ShareReceiveActivity)
        }
    }
    
    private fun handleSharedText(text: String) {
        shareFileName.text = "Shared Text"
        shareFileSize.text = "${text.length} characters"
        sharePreview.setBackgroundColor(android.graphics.Color.LTGRAY)
        sharePreview.setImageDrawable(null)
        
        // Store text as "file" data for sending
        currentFileUri = null // No URI for text
    }
    
    private fun sendFileViaMqtt() {
        lifecycleScope.launch {
            try {
                if (!MqttManager.ensureConnected(this@ShareReceiveActivity)) {
                    Toast.makeText(this@ShareReceiveActivity, "Failed to connect to MQTT", Toast.LENGTH_LONG).show()
                    return@launch
                }
                
                // Handle text sharing
                val sharedText = intent.getStringExtra(Intent.EXTRA_TEXT)
                if (sharedText != null && currentFileUri == null) {
                    val success = mqttClient.publish(
                        sharedText.toByteArray(),
                        "text/plain",
                        "clipboard.txt"
                    )
                    
                    if (success) {
                        Toast.makeText(this@ShareReceiveActivity, "Text shared successfully!", Toast.LENGTH_SHORT).show()
                        finish()
                    } else {
                        Toast.makeText(this@ShareReceiveActivity, "Failed to share text", Toast.LENGTH_LONG).show()
                    }
                    return@launch
                }
                
                // Handle file sharing
                val uri = currentFileUri ?: return@launch
                val fileName = getFileName(uri)
                val mimeType = contentResolver.getType(uri) ?: "application/octet-stream"
                
                // Read file content
                val fileData = readFileData(uri)
                if (fileData == null) {
                    Toast.makeText(this@ShareReceiveActivity, "Failed to read file", Toast.LENGTH_LONG).show()
                    return@launch
                }
                
                val success = mqttClient.publish(fileData, mimeType, fileName)
                
                if (success) {
                    Toast.makeText(this@ShareReceiveActivity, "File shared successfully!", Toast.LENGTH_SHORT).show()
                    finish()
                } else {
                    Toast.makeText(this@ShareReceiveActivity, "Failed to share file", Toast.LENGTH_LONG).show()
                }
                
            } catch (e: Exception) {
                Toast.makeText(this@ShareReceiveActivity, "Error: ${e.message}", Toast.LENGTH_LONG).show()
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
    
    override fun onDestroy() {
        super.onDestroy()
        // Don't disconnect here - let other activities use the same connection
    }
}