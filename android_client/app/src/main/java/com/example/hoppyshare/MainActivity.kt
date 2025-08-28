package com.example.hoppyshare

import android.animation.ArgbEvaluator
import android.animation.ValueAnimator
import android.content.ClipboardManager
import android.content.Context
import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.widget.Button
import android.widget.ImageView
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
    private lateinit var mascotIcon: ImageView
    private var animationHandler: Handler? = null
    private var animationRunnable: Runnable? = null
    private var currentAnimationFrame = 0
    private var isShowingNotification = false
    
    private val loadingIcons = arrayOf(
        R.drawable.loading_1,
        R.drawable.loading_2,
        R.drawable.loading_3,
        R.drawable.loading_4
    )
    
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
        mascotIcon = findViewById(R.id.mascotIcon)
        val pickFileButton = findViewById<Button>(R.id.pickFileButton)
        val sendClipboardButton = findViewById<Button>(R.id.sendClipboardButton)
        
        pickFileButton.setOnClickListener {
            // Launch file picker for any file type
            filePickerLauncher.launch("*/*")
        }
        
        sendClipboardButton.setOnClickListener {
            sendClipboardContent()
        }
        
        mascotIcon.setOnClickListener {
            if (isShowingNotification) {
                lifecycleScope.launch {
                    viewReceivedMessage()
                }
            }
        }
        
        // Disable anti-aliasing for pixel art
        mascotIcon.drawable?.isFilterBitmap = false
        
        // Start loading animation
        startLoadingAnimation()
        
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
                statusText.text = "Failed to connect. Restart app"
            }
        }
    }
    
    private suspend fun handleIncomingMessage() {
        val lastMessage = mqttClient.getLastMessage()
        if (lastMessage != null) {
            // Switch to notification icon
            runOnUiThread {
                showNotificationIcon()
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
            if (!hasMessage && isShowingNotification) {
                // Message was viewed/cleared, resume loading animation
                runOnUiThread {
                    resumeLoadingAnimation()
                }
            } else if (hasMessage && !isShowingNotification) {
                // Message arrived while we were away, show notification
                runOnUiThread {
                    showNotificationIcon()
                }
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
        stopLoadingAnimation()
    }
    
    private fun startLoadingAnimation() {
        animationHandler = Handler(Looper.getMainLooper())
        animationRunnable = object : Runnable {
            override fun run() {
                if (!isShowingNotification) {
                    mascotIcon.setImageResource(loadingIcons[currentAnimationFrame])
                    mascotIcon.drawable?.isFilterBitmap = false // Disable anti-aliasing
                    currentAnimationFrame = (currentAnimationFrame + 1) % loadingIcons.size
                    animationHandler?.postDelayed(this, 200)
                }
            }
        }
        animationHandler?.post(animationRunnable!!)
    }
    
    private fun stopLoadingAnimation() {
        animationHandler?.removeCallbacks(animationRunnable!!)
        animationHandler = null
        animationRunnable = null
    }
    
    private fun showNotificationIcon() {
        isShowingNotification = true
        mascotIcon.setImageResource(R.drawable.notification_icon)
        mascotIcon.drawable?.isFilterBitmap = false // Disable anti-aliasing
        animateBackgroundColor(getColor(R.color.brand_white), getColor(R.color.primary_light))
    }
    
    private fun resumeLoadingAnimation() {
        isShowingNotification = false
        startLoadingAnimation()
        animateBackgroundColor(getColor(R.color.primary_light), getColor(R.color.brand_white))
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
            val intent = Intent(Intent.ACTION_VIEW, Uri.parse("https://hoppyshare.com/add-device/mobile"))
            startActivity(intent)
            finish()
        } catch (e: Exception) {
            // If no browser available, show error
            Toast.makeText(this, "Please visit http://hoppyshare.com/add-device/mobile to set up your device", Toast.LENGTH_SHORT).show()
        }
    }
    
    private fun sendClipboardContent() {
        val clipboardManager = getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
        val clipData = clipboardManager.primaryClip
        
        if (clipData != null && clipData.itemCount > 0) {
            val clipText = clipData.getItemAt(0).text
            if (clipText != null && clipText.isNotEmpty()) {
                // Send clipboard text via MQTT
                lifecycleScope.launch {
                    try {
                        val textBytes = clipText.toString().toByteArray(Charsets.UTF_8)
                        val success = mqttClient.publish(textBytes, "text/plain", "clipboard.txt")
                        if (success) {
                            Toast.makeText(this@MainActivity, "Clipboard content sent!", Toast.LENGTH_SHORT).show()
                        } else {
                            Toast.makeText(this@MainActivity, "Failed to send clipboard", Toast.LENGTH_SHORT).show()
                        }
                    } catch (e: Exception) {
                        Toast.makeText(this@MainActivity, "Failed to send clipboard: ${e.message}", Toast.LENGTH_SHORT).show()
                    }
                }
            } else {
                Toast.makeText(this, "Clipboard is empty or contains no text", Toast.LENGTH_SHORT).show()
            }
        } else {
            Toast.makeText(this, "Clipboard is empty", Toast.LENGTH_SHORT).show()
        }
    }
    
    private fun animateBackgroundColor(fromColor: Int, toColor: Int) {
        val mainLayout = findViewById<androidx.constraintlayout.widget.ConstraintLayout>(R.id.main)
        val colorAnimator = ValueAnimator.ofObject(ArgbEvaluator(), fromColor, toColor)
        colorAnimator.duration = 200
        colorAnimator.addUpdateListener { animator ->
            mainLayout.setBackgroundColor(animator.animatedValue as Int)
        }
        colorAnimator.start()
    }
}