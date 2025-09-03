package com.hoppyshare.android

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import android.content.Intent
import android.graphics.BitmapFactory
import android.net.Uri
import android.os.Bundle
import android.os.Environment
import android.provider.MediaStore
import android.widget.Button
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import android.widget.Toast
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.FileProvider
import androidx.lifecycle.lifecycleScope
import kotlinx.coroutines.launch
import java.io.File
import java.io.FileOutputStream

class MessageViewActivity : AppCompatActivity() {
    
    private lateinit var messageInfoText: TextView
    private lateinit var textContentView: TextView
    private lateinit var imagePreview: ImageView
    private lateinit var copyButton: Button
    private lateinit var saveButton: Button
    private lateinit var closeButton: Button
    
    private var messageData: MqttClient.LastMessage? = null
    
    // Request permission for file writing
    private val requestPermissionLauncher = registerForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { isGranted ->
        if (isGranted) {
            saveFileToStorage()
        } else {
            Toast.makeText(this, "Storage permission required to save file", Toast.LENGTH_LONG).show()
        }
    }
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_message_view)
        
        initViews()
        loadMessageData()
        setupClickListeners()
    }
    
    private fun initViews() {
        messageInfoText = findViewById(R.id.messageInfoText)
        textContentView = findViewById(R.id.textContentView)
        imagePreview = findViewById(R.id.imagePreview)
        copyButton = findViewById(R.id.copyButton)
        saveButton = findViewById(R.id.saveButton)
        closeButton = findViewById(R.id.closeButton)
    }
    
    private fun loadMessageData() {
        lifecycleScope.launch {
            val mqttClient = MqttManager.getInstance(this@MessageViewActivity)
            messageData = mqttClient.getLastMessage()
            
            if (messageData != null) {
                displayMessage(messageData!!)
            } else {
                Toast.makeText(this@MessageViewActivity, "No message to display", Toast.LENGTH_SHORT).show()
                finish()
            }
        }
    }
    
    private fun displayMessage(message: MqttClient.LastMessage) {
        // Update info text
        val sizeText = formatFileSize(message.payload.size.toLong())
        messageInfoText.text = "${message.filename} • ${message.contentType} • $sizeText"
        
        when {
            message.contentType.startsWith("text/") -> {
                showTextContent(message)
            }
            message.contentType.startsWith("image/") -> {
                showImageContent(message)
            }
            else -> {
                showFileInfo(message)
            }
        }
    }
    
    private fun showTextContent(message: MqttClient.LastMessage) {
        textContentView.visibility = TextView.VISIBLE
        imagePreview.visibility = ImageView.GONE
        
        textContentView.text = String(message.payload)
        copyButton.visibility = Button.VISIBLE
    }
    
    private fun showImageContent(message: MqttClient.LastMessage) {
        textContentView.visibility = TextView.GONE
        imagePreview.visibility = ImageView.VISIBLE
        
        try {
            val bitmap = BitmapFactory.decodeByteArray(message.payload, 0, message.payload.size)
            if (bitmap != null) {
                imagePreview.setImageBitmap(bitmap)
            } else {
                // Fallback to file info if image can't be decoded
                showFileInfo(message)
                return
            }
        } catch (e: Exception) {
            showFileInfo(message)
            return
        }
        
        copyButton.visibility = Button.GONE
    }
    
    private fun showFileInfo(message: MqttClient.LastMessage) {
        textContentView.visibility = TextView.GONE
        imagePreview.visibility = ImageView.GONE
        
        copyButton.visibility = Button.GONE
    }
    
    private fun setupClickListeners() {
        copyButton.setOnClickListener {
            copyToClipboard()
        }
        
        saveButton.setOnClickListener {
            checkPermissionAndSaveFile()
        }
        
        closeButton.setOnClickListener {
            clearMessageAndFinish()
        }
    }
    
    private fun copyToClipboard() {
        val message = messageData ?: return
        
        if (message.contentType.startsWith("text/")) {
            val clipboard = getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
            val clip = ClipData.newPlainText("Received Text", String(message.payload))
            clipboard.setPrimaryClip(clip)
            Toast.makeText(this, "Text copied to clipboard", Toast.LENGTH_SHORT).show()
        }
    }
    
    private fun checkPermissionAndSaveFile() {
        // For Android 10+ (API 29+), we don't need WRITE_EXTERNAL_STORAGE for app-specific directories
        saveFileToStorage()
    }
    
    private fun saveFileToStorage() {
        val message = messageData ?: return
        
        try {
            // Save to Downloads folder using MediaStore
            val fileName = message.filename
            val mimeType = message.contentType
            
            // Create content values for MediaStore
            val contentValues = android.content.ContentValues().apply {
                put(MediaStore.MediaColumns.DISPLAY_NAME, fileName)
                put(MediaStore.MediaColumns.MIME_TYPE, mimeType)
                put(MediaStore.MediaColumns.RELATIVE_PATH, Environment.DIRECTORY_DOWNLOADS)
            }
            
            val resolver = contentResolver
            val uri = resolver.insert(MediaStore.Downloads.EXTERNAL_CONTENT_URI, contentValues)
            
            if (uri != null) {
                resolver.openOutputStream(uri)?.use { outputStream ->
                    outputStream.write(message.payload)
                }
                Toast.makeText(this, "File saved to Downloads folder", Toast.LENGTH_SHORT).show()
            } else {
                // Fallback to app-specific directory
                saveToAppDirectory(message)
            }
            
        } catch (e: Exception) {
            Toast.makeText(this, "Failed to save file: ${e.message}", Toast.LENGTH_LONG).show()
        }
    }
    
    private fun saveToAppDirectory(message: MqttClient.LastMessage) {
        try {
            val downloadsDir = File(getExternalFilesDir(Environment.DIRECTORY_DOWNLOADS), "HoppyShare")
            if (!downloadsDir.exists()) {
                downloadsDir.mkdirs()
            }
            
            val file = File(downloadsDir, message.filename)
            FileOutputStream(file).use { fos ->
                fos.write(message.payload)
            }
            
            Toast.makeText(this, "File saved to HoppyShare folder", Toast.LENGTH_SHORT).show()
        } catch (e: Exception) {
            Toast.makeText(this, "Failed to save file: ${e.message}", Toast.LENGTH_LONG).show()
        }
    }
    
    private fun formatFileSize(bytes: Long): String {
        return when {
            bytes >= 1024 * 1024 -> "${bytes / (1024 * 1024)}MB"
            bytes >= 1024 -> "${bytes / 1024}KB"
            else -> "${bytes}B"
        }
    }
    
    private fun clearMessageAndFinish() {
        lifecycleScope.launch {
            val mqttClient = MqttManager.getInstance(this@MessageViewActivity)
            mqttClient.clearLastMessage()
            finish()
        }
    }
    
    override fun onBackPressed() {
        super.onBackPressed()
        clearMessageAndFinish()
    }
}