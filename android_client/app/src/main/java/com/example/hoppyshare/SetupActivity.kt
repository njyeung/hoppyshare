package com.example.hoppyshare

import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.widget.Button
import android.widget.TextView
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.core.content.ContextCompat

class SetupActivity : ComponentActivity() {
    
    private lateinit var statusText: TextView
    private lateinit var finishButton: Button
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_setup)
        
        statusText = findViewById(R.id.statusText)
        finishButton = findViewById(R.id.finishButton)
        
        finishButton.setOnClickListener {
            // Navigate to main activity
            val intent = Intent(this, MainActivity::class.java)
            startActivity(intent)
            finish()
        }
        
        // Handle the incoming deep link
        handleDeepLink(intent)
    }
    
    private fun handleDeepLink(intent: Intent?) {
        val action = intent?.action
        val data = intent?.data
        
        if (Intent.ACTION_VIEW == action && data != null) {
            processCertificateData(data)
        } else {
            showError("Invalid setup link")
        }
    }
    
    private fun processCertificateData(uri: Uri) {
        try {
            val cert = uri.getQueryParameter("cert")
            val key = uri.getQueryParameter("key") 
            val ca = uri.getQueryParameter("ca")
            val groupKey = uri.getQueryParameter("group_key")
            val deviceId = uri.getQueryParameter("device_id")
            
            if (cert == null || key == null || ca == null || groupKey == null || deviceId == null) {
                showError("Missing certificate data")
                return
            }
            
            // Store certificates securely
            val success = storeCertificates(cert, key, ca, groupKey, deviceId)
            
            if (success) {
                showSuccess()
            } else {
                showError("Failed to store certificates")
            }
            
        } catch (e: Exception) {
            showError("Error processing certificate data: ${e.message}")
        }
    }
    
    private fun storeCertificates(cert: String, key: String, ca: String, groupKey: String, deviceId: String): Boolean {
        return try {
            // Store in SharedPreferences for now (should use Android Keystore for production)
            val prefs = getSharedPreferences("hoppyshare_certs", MODE_PRIVATE)
            with(prefs.edit()) {
                putString("client_cert", cert)
                putString("client_key", key)
                putString("ca_cert", ca)
                putString("group_key", groupKey)
                putString("device_id", deviceId)
                putBoolean("is_configured", true)
                apply()
            }
            true
        } catch (e: Exception) {
            false
        }
    }
    
    private fun showSuccess() {
        statusText.text = "Device setup complete!"
        statusText.setTextColor(ContextCompat.getColor(this, R.color.success_green))
        finishButton.text = "Continue to App"
        finishButton.isEnabled = true
        
        Toast.makeText(this, "HoppyShare is now configured!", Toast.LENGTH_LONG).show()
    }
    
    private fun showError(message: String) {
        statusText.text = "Setup failed: $message"
        statusText.setTextColor(ContextCompat.getColor(this, R.color.error_red))
        finishButton.text = "Close"
        finishButton.isEnabled = true
        
        Toast.makeText(this, message, Toast.LENGTH_LONG).show()
    }
}