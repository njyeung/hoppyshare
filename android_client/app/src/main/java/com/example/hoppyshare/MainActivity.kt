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
    private lateinit var viewMessageButton: Button
    
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
        mqttClient = MqttManager.getInstance(this)
        
        setupUI()
        connectToMqtt()
    }
    
    private fun setupUI() {
        statusText = findViewById(R.id.statusText)
        viewMessageButton = findViewById(R.id.viewMessageButton)
        val pickFileButton = findViewById<Button>(R.id.pickFileButton)
        
        pickFileButton.setOnClickListener {
            // Launch file picker for any file type
            filePickerLauncher.launch("*/*")
        }
        
        viewMessageButton.setOnClickListener {
            lifecycleScope.launch {
                viewReceivedMessage()
            }
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
            
            // Enable the view message button
            runOnUiThread {
                viewMessageButton.isEnabled = true
                viewMessageButton.backgroundTintList =
                    getColorStateList(R.color.button_primary_background)
            }
        }
    }
    
    private suspend fun viewReceivedMessage() {
        val lastMessage = mqttClient.getLastMessage()
        if (lastMessage != null) {
            // Launch MessageViewActivity
            val intent = Intent(this, MessageViewActivity::class.java)
            startActivity(intent)
        }
    }
    
    override fun onResume() {
        super.onResume()
        
        lifecycleScope.launch {
            // Check if message was cleared by MessageViewActivity
            val hasMessage = mqttClient.getLastMessage() != null
            if (!hasMessage && viewMessageButton.isEnabled) {
                // Message was viewed/cleared, update UI
                viewMessageButton.isEnabled = false
                viewMessageButton.backgroundTintList = getColorStateList(R.color.secondary)
            }
            
            // Check and reconnect MQTT if needed
            statusText.text = "Reconnecting..."
            val connected = MqttManager.ensureConnected(this@MainActivity)
            if (connected) {
                statusText.text = "Connected"
            } else {
                statusText.text = "Failed to reconnect. Restart app"
            }
        }
    }
    
    override fun onDestroy() {
        super.onDestroy()
        // Don't disconnect here - let other activities use the same connection
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
            Toast.makeText(this, "Please visit http://hoppyshare.vercel.app/add-device/mobile to set up your device", Toast.LENGTH_SHORT).show()
        }
    }
}