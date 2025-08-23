package com.example.hoppyshare

import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity

class ShareReceiveActivity : AppCompatActivity() {
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        when (intent?.action) {
            Intent.ACTION_SEND -> handleSingleFile()
            Intent.ACTION_SEND_MULTIPLE -> handleMultipleFiles()
            else -> {
                Toast.makeText(this, "Unsupported action", Toast.LENGTH_SHORT).show()
                finish()
            }
        }
    }
    
    private fun handleSingleFile() {
        val uri = intent.getParcelableExtra<Uri>(Intent.EXTRA_STREAM)
        if (uri != null) {
            val fileName = getFileName(uri)
            val fileSize = getFileSize(uri)
            
            Toast.makeText(
                this, 
                "Would send: $fileName (${formatFileSize(fileSize)} bytes)\nVia HoppyShare MQTT", 
                Toast.LENGTH_LONG
            ).show()
            
            // TODO: Implement actual file sending via MQTT
        } else {
            Toast.makeText(this, "No file received", Toast.LENGTH_SHORT).show()
        }
        
        // Close this activity and return to the sharing app
        finish()
    }
    
    private fun handleMultipleFiles() {
        val uris = intent.getParcelableArrayListExtra<Uri>(Intent.EXTRA_STREAM)
        if (!uris.isNullOrEmpty()) {
            Toast.makeText(
                this, 
                "Would send ${uris.size} files via HoppyShare MQTT", 
                Toast.LENGTH_LONG
            ).show()
            
            // TODO: Implement multiple file sending via MQTT
        } else {
            Toast.makeText(this, "No files received", Toast.LENGTH_SHORT).show()
        }
        
        finish()
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