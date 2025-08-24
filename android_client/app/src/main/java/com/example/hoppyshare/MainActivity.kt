package com.example.hoppyshare

import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.widget.Button
import android.widget.TextView
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
        // Launch the confirmation activity with the selected file
        val intent = Intent(this, FileConfirmActivity::class.java)
        intent.putExtra("file_uri", uri.toString())
        startActivity(intent)
    }
}