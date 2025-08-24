package com.example.hoppyshare

import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.widget.Button
import android.widget.ImageView
import android.widget.TextView
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity

class FileConfirmActivity : AppCompatActivity() {
    
    private lateinit var sharePreview: ImageView
    private lateinit var shareFileName: TextView
    private lateinit var shareFileSize: TextView
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_share_receive)
        
        setupUI()
        handleFileUri()
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
            Toast.makeText(this, "Would send via HoppyShare MQTT", Toast.LENGTH_SHORT).show()
            // TODO: Implement actual file sending via MQTT
            finish()
        }
    }
    
    private fun handleFileUri() {
        val uriString = intent.getStringExtra("file_uri")
        if (uriString != null) {
            val uri = Uri.parse(uriString)
            val fileName = getFileName(uri)
            val fileSize = getFileSize(uri)
            
            showFilePreview(uri, fileName, fileSize)
        } else {
            Toast.makeText(this, "No file selected", Toast.LENGTH_SHORT).show()
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
}