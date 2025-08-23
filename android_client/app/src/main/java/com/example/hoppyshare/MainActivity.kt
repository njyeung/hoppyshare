package com.example.hoppyshare

import android.net.Uri
import android.os.Bundle
import android.widget.Button
import android.widget.TextView
import android.widget.Toast
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AppCompatActivity
import androidx.core.view.ViewCompat
import androidx.core.view.WindowInsetsCompat

class MainActivity : AppCompatActivity() {
    
    private lateinit var statusText: TextView
    
    // Modern file picker using ActivityResultContracts
    private val filePickerLauncher = registerForActivityResult(
        ActivityResultContracts.GetContent()
    ) { uri: Uri? ->
        uri?.let { handleSelectedFile(it) }
    }
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContentView(R.layout.activity_main)
        
        ViewCompat.setOnApplyWindowInsetsListener(findViewById(R.id.main)) { v, insets ->
            val systemBars = insets.getInsets(WindowInsetsCompat.Type.systemBars())
            v.setPadding(systemBars.left, systemBars.top, systemBars.right, systemBars.bottom)
            insets
        }
        
        setupUI()
    }
    
    private fun setupUI() {
        statusText = findViewById(R.id.statusText)
        val pickFileButton = findViewById<Button>(R.id.pickFileButton)
        
        pickFileButton.setOnClickListener {
            // Launch file picker for any file type
            filePickerLauncher.launch("*/*")
        }
    }
    
    private fun handleSelectedFile(uri: Uri) {
        val fileName = getFileName(uri)
        val fileSize = getFileSize(uri)
        
        statusText.text = "Selected: $fileName (${formatFileSize(fileSize)})"
        
        Toast.makeText(
            this,
            "Would send: $fileName via HoppyShare MQTT",
            Toast.LENGTH_LONG
        ).show()
        
        // TODO: Implement actual file sending via MQTT
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