package com.example.hoppyshare

import android.content.Intent
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
import androidx.lifecycle.lifecycleScope
import kotlinx.coroutines.launch

class MainActivity : AppCompatActivity() {
    
    private lateinit var statusText: TextView
    private lateinit var mqttClient: MqttClient
    
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
        
        // Check if certificates are configured
        if (!isCertificatesConfigured()) {
            redirectToSetup()
            return
        }
        
        // Initialize MQTT client
        mqttClient = MqttClient(this)
        
        setupUI()
        connectToMqtt()
    }
    
    private fun setupUI() {
        statusText = findViewById(R.id.statusText)
        val pickFileButton = findViewById<Button>(R.id.pickFileButton)
        
        pickFileButton.setOnClickListener {
            // Launch file picker for any file type
            filePickerLauncher.launch("*/*")
        }
        
        // Set up message callback
        mqttClient.setOnMessageCallback {
            lifecycleScope.launch {
                handleIncomingMessage()
            }
        }
    }
    
    private fun connectToMqtt() {
        lifecycleScope.launch {
            statusText.text = "Connecting to MQTT..."
            val clientId = mqttClient.connect()
            if (clientId != null) {
                statusText.text = "Connected"
            } else {
                statusText.text = "Failed to connect"
            }
        }
    }
    
    private suspend fun handleIncomingMessage() {
        val lastMessage = mqttClient.getLastMessage()
        if (lastMessage != null) {
            statusText.text = "Received: ${lastMessage.filename} (${lastMessage.contentType})"
            
            // If it's text content, show it in a toast
            if (lastMessage.contentType.startsWith("text/")) {
                val textContent = String(lastMessage.payload)
                Toast.makeText(this, "Text: $textContent", Toast.LENGTH_LONG).show()
            } else {
                Toast.makeText(this, "Received file: ${lastMessage.filename}", Toast.LENGTH_LONG).show()
            }
        }
    }
    
    override fun onDestroy() {
        super.onDestroy()
        if (::mqttClient.isInitialized) {
            mqttClient.disconnect()
        }
    }
    
    private fun handleSelectedFile(uri: Uri) {
        // Launch the confirmation activity with the selected file
        val intent = Intent(this, FileConfirmActivity::class.java)
        intent.putExtra("file_uri", uri.toString())
        startActivity(intent)
    }
    
    private fun isCertificatesConfigured(): Boolean {
        val prefs = getSharedPreferences("hoppyshare_certs", MODE_PRIVATE)
        return prefs.getBoolean("is_configured", false)
    }
    
    private fun redirectToSetup() {
        try {
            val intent = Intent(Intent.ACTION_VIEW, Uri.parse("https://hoppyshare.vercel.app/add-device/mobile"))
            startActivity(intent)
            finish()
        } catch (e: Exception) {
            // If no browser available, show error
            statusText?.text = "Please visit http://hoppyshare.vercel.app/add-device/mobile to set up your device"
        }
    }
}