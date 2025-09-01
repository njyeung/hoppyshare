package com.example.hoppyshare

import android.animation.ArgbEvaluator
import android.animation.ValueAnimator
import android.Manifest
import android.content.ClipboardManager
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.media.MediaPlayer
import android.net.Uri
import android.os.Build
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.os.VibrationEffect
import android.os.Vibrator
import android.os.VibratorManager
import android.widget.Button
import android.widget.ImageView
import android.widget.TextView
import android.widget.Toast
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AppCompatActivity
import androidx.core.view.ViewCompat
import androidx.core.view.WindowInsetsCompat
import androidx.lifecycle.lifecycleScope
import kotlinx.coroutines.launch

class MainActivity : AppCompatActivity() {
    
    private lateinit var titleText: TextView
    private lateinit var statusText: TextView
    private lateinit var mqttClient: MqttClient
    private lateinit var mascotIcon: ImageView
    private lateinit var bleToggleButton: Button
    private var bleManager: BLEManager? = null
    private var animationHandler: Handler? = null
    private var animationRunnable: Runnable? = null
    private var currentAnimationFrame = 0
    private var isShowingNotification = false
    private var isShowingError = false
    private var errorHandler: Handler? = null
    private var notificationSound: MediaPlayer? = null
    
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
        
        // Initialize settings
        Settings.initialize(this)
        
        // Set up settings change callback
        Settings.setOnSettingsChangedCallback { newSettings ->
            runOnUiThread {
                onSettingsChanged(newSettings)
            }
        }
        
        // Initialize MQTT client
        mqttClient = MqttManager.getInstance(this)
        
        setupUI()
        setupNotificationSound()
        connectToMqtt()
        
        // Check if this activity was started to show an error
        handleErrorIntent()
    }
    
    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        setIntent(intent) // Update the activity's intent
        handleErrorIntent()
    }
    
    private fun setupUI() {
        titleText = findViewById(R.id.titleText)
        statusText = findViewById(R.id.statusText)
        mascotIcon = findViewById(R.id.mascotIcon)
        bleToggleButton = findViewById<Button>(R.id.bleToggleButton)
        val pickFileButton = findViewById<Button>(R.id.pickFileButton)
        val sendClipboardButton = findViewById<Button>(R.id.sendClipboardButton)
        
        pickFileButton.setOnClickListener {
            // Launch file picker for any file type
            filePickerLauncher.launch("*/*")
        }
        
        sendClipboardButton.setOnClickListener {
            android.util.Log.d("MainActivity", "SENDING CLIPBOARD CONTENT")
            sendClipboardContent()
        }
        
        bleToggleButton.setOnClickListener {
            android.util.Log.d("MainActivity", "BLE toggle button clicked")
            toggleBLE()
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
        
        // Update device name in title
        updateDeviceName()
        
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
                // Initialize BLE after MQTT connects and Config is loaded
                initializeBLE()
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
                vibratePhone()
                playNotificationSound()
            }
        }
    }
    
    private fun vibratePhone() {
        val vibrator = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            val vibratorManager = getSystemService(Context.VIBRATOR_MANAGER_SERVICE) as VibratorManager
            vibratorManager.defaultVibrator
        } else {
            @Suppress("DEPRECATION")
            getSystemService(Context.VIBRATOR_SERVICE) as Vibrator
        }
        
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            // Create a short vibration pattern - two quick pulses
            val vibrationEffect = VibrationEffect.createWaveform(
                longArrayOf(0, 200),
                -1 // don't repeat
            )
            vibrator.vibrate(vibrationEffect)
        } else {
            @Suppress("DEPRECATION")
            vibrator.vibrate(400) // 400ms vibration for older devices
        }
    }
    
    private fun setupNotificationSound() {
        try {
            val assetFileDescriptor = assets.openFd("notification.wav")
            notificationSound = MediaPlayer().apply {
                setDataSource(assetFileDescriptor.fileDescriptor, assetFileDescriptor.startOffset, assetFileDescriptor.length)
                prepare()
                assetFileDescriptor.close()
            }
        } catch (e: Exception) {
            android.util.Log.e("MainActivity", "Failed to setup notification sound: ${e.message}")
        }
    }
    
    private fun playNotificationSound() {
        try {
            val settings = Settings.getCurrentSettings()
            if (!settings.muted) {
                notificationSound?.let { player ->
                    if (player.isPlaying) {
                        player.stop()
                        player.prepare()
                    }
                    player.start()
                }
            } else {
                android.util.Log.d("MainActivity", "Notification sound muted by settings")
            }
        } catch (e: Exception) {
            android.util.Log.e("MainActivity", "Failed to play notification sound: ${e.message}")
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
            val hasMessage = mqttClient.getLastMessage() != null
            if (!hasMessage && isShowingNotification) {
                runOnUiThread {
                    resumeLoadingAnimation()
                }
            } else if (hasMessage && !isShowingNotification) {
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
        errorHandler?.removeCallbacksAndMessages(null)
        
        // Clean up notification sound
        notificationSound?.release()
        notificationSound = null
        
        // Stop BLE if running
        bleManager?.stop()
    }
    
    private fun startLoadingAnimation() {
        animationHandler = Handler(Looper.getMainLooper())
        animationRunnable = object : Runnable {
            override fun run() {
                if (!isShowingNotification && !isShowingError) {
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
        isShowingError = false
        startLoadingAnimation()
        animateBackgroundColor(getColor(R.color.primary_light), getColor(R.color.brand_white))
    }
    
    private fun showErrorIcon() {
        android.util.Log.d("MainActivity", "showErrorIcon called")
        isShowingError = true
        isShowingNotification = false
        mascotIcon.setImageResource(R.drawable.error_icon)
        mascotIcon.drawable?.isFilterBitmap = false // Disable anti-aliasing
        
        // Clear error state after 3 seconds
        errorHandler?.removeCallbacksAndMessages(null)
        errorHandler = Handler(Looper.getMainLooper())
        errorHandler?.postDelayed({
            android.util.Log.d("MainActivity", "Error timeout - resuming loading animation")
            resumeLoadingAnimation()
        }, 3000)
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
                val textBytes = clipText.toString().toByteArray(Charsets.UTF_8)
                
                // Check 25MB size limit
                if (textBytes.size > 25 * 1024 * 1024) {
                    Toast.makeText(this, "Clipboard too large (>25MB). Operation cancelled.", Toast.LENGTH_LONG).show()
                    showErrorIcon()
                    return
                }
                
                // Send clipboard text via MQTT and BLE
                lifecycleScope.launch {
                    try {
                        val mqttSuccess = mqttClient.publish(textBytes, "text/plain", "clipboard.txt")
                        
                        // Also send via BLE if enabled (with 3MB size limit)
                        bleManager?.let { ble ->
                            if (ble.isStarted()) {
                                try {
                                    if (textBytes.size > 3 * 1024 * 1024) {
                                        Toast.makeText(this@MainActivity, "Clipboard too large for BLE (>3MB)", Toast.LENGTH_SHORT).show()
                                        android.util.Log.d("MainActivity", "Clipboard too large for BLE (>3MB), skipping BLE send")
                                    } else {
                                        val encoded = MqttCodec.encodeMessage("text/plain", "clipboard.txt", Config.deviceId, textBytes)
                                        ble.sendData(encoded)
                                        android.util.Log.d("MainActivity", "Clipboard sent via BLE")
                                    }
                                } catch (e: Exception) {
                                    android.util.Log.e("MainActivity", "Failed to send clipboard via BLE: ${e.message}", e)
                                }
                            }
                        }
                        
                        if (mqttSuccess) {
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
    
    private fun updateDeviceName() {
        val settings = Settings.getCurrentSettings()
        titleText.text = settings.nickname
    }
    
    private fun onSettingsChanged(newSettings: AppSettings) {
        android.util.Log.d("MainActivity", "Settings changed: $newSettings")
        
        // Update device name in title if it changed
        updateDeviceName()
        
        // Note: muted setting is checked at runtime in playNotificationSound()
        // send_to_self setting is handled in MQTT client
    }
    
    private fun initializeBLE() {
        android.util.Log.d("MainActivity", "initializeBLE called")
        
        if (!Config.isConfigured()) {
            android.util.Log.w("MainActivity", "Config not loaded, cannot initialize BLE")
            return
        }
        
        android.util.Log.d("MainActivity", "Creating BLE manager with clientId: ${mqttClient.getClientId()}, deviceId: ${Config.deviceId}")
        
        try {
            bleManager = BLEManager(this, mqttClient.getClientId(), Config.deviceId)
            bleManager?.setOnMessageCallback {
                lifecycleScope.launch {
                    handleIncomingBLEMessage()
                }
            }
            
            // Connect BLE manager to MQTT client for dual publishing
            bleManager?.let { mqttClient.setBLEManager(it) }
            
            android.util.Log.d("MainActivity", "BLE manager initialized: ${bleManager != null}")
        } catch (e: Exception) {
            android.util.Log.e("MainActivity", "Failed to initialize BLE: ${e.message}", e)
        }
    }
    
    private fun toggleBLE() {
        android.util.Log.d("MainActivity", "toggleBLE called")
        android.util.Log.d("MainActivity", "bleManager is null: ${bleManager == null}")
        
        bleManager?.let { ble ->
            if (ble.isStarted()) {
                android.util.Log.d("MainActivity", "Stopping BLE")
                ble.stop()
                Toast.makeText(this, "BLE stopped", Toast.LENGTH_SHORT).show()
                bleToggleButton.text = "Start BLE"
                bleToggleButton.backgroundTintList = getColorStateList(R.color.secondary)
            } else {
                android.util.Log.d("MainActivity", "Attempting to start BLE")
                
                // Check and request permissions first
                if (hasRequiredBLEPermissions()) {
                    android.util.Log.d("MainActivity", "Permissions granted, starting BLE")
                    try {
                        if (ble.start()) {
                            Toast.makeText(this, "BLE started", Toast.LENGTH_SHORT).show()
                            bleToggleButton.text = "Stop BLE"
                            bleToggleButton.backgroundTintList = getColorStateList(R.color.button_primary_background)
                        } else {
                            Toast.makeText(this, "Failed to start BLE", Toast.LENGTH_LONG).show()
                            bleToggleButton.text = "Start BLE"
                            bleToggleButton.backgroundTintList = getColorStateList(R.color.secondary)
                        }
                    } catch (e: Exception) {
                        android.util.Log.e("MainActivity", "BLE start crashed: ${e.message}", e)
                        Toast.makeText(this, "BLE start crashed: ${e.message}", Toast.LENGTH_LONG).show()
                        bleToggleButton.text = "Start BLE"
                        bleToggleButton.backgroundTintList = getColorStateList(R.color.secondary)
                    }
                } else {
                    android.util.Log.d("MainActivity", "Requesting BLE permissions")
                    requestBLEPermissions()
                    return
                }
            }
        }
    }
    
    private fun hasRequiredBLEPermissions(): Boolean {
        val requiredPermissions = mutableListOf<String>()
        
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            requiredPermissions.addAll(listOf(
                Manifest.permission.BLUETOOTH_SCAN,
                Manifest.permission.BLUETOOTH_ADVERTISE,
                Manifest.permission.BLUETOOTH_CONNECT
            ))
        } else {
            requiredPermissions.addAll(listOf(
                Manifest.permission.BLUETOOTH,
                Manifest.permission.BLUETOOTH_ADMIN,
                Manifest.permission.ACCESS_FINE_LOCATION
            ))
        }
        
        return requiredPermissions.all { permission ->
            ContextCompat.checkSelfPermission(this, permission) == PackageManager.PERMISSION_GRANTED
        }
    }
    
    private fun requestBLEPermissions() {
        val requiredPermissions = mutableListOf<String>()
        
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            requiredPermissions.addAll(listOf(
                Manifest.permission.BLUETOOTH_SCAN,
                Manifest.permission.BLUETOOTH_ADVERTISE,
                Manifest.permission.BLUETOOTH_CONNECT
            ))
        } else {
            requiredPermissions.addAll(listOf(
                Manifest.permission.BLUETOOTH,
                Manifest.permission.BLUETOOTH_ADMIN,
                Manifest.permission.ACCESS_FINE_LOCATION
            ))
        }
        
        ActivityCompat.requestPermissions(this, requiredPermissions.toTypedArray(), BLE_PERMISSION_REQUEST_CODE)
    }
    
    override fun onRequestPermissionsResult(requestCode: Int, permissions: Array<out String>, grantResults: IntArray) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)
        
        if (requestCode == BLE_PERMISSION_REQUEST_CODE) {
            val allGranted = grantResults.all { it == PackageManager.PERMISSION_GRANTED }
            android.util.Log.d("MainActivity", "BLE permissions result: allGranted = $allGranted")
            
            if (allGranted) {
                // Try to start BLE again
                bleManager?.let { ble ->
                    try {
                        if (ble.start()) {
                            Toast.makeText(this, "BLE started", Toast.LENGTH_SHORT).show()
                            bleToggleButton.text = "Stop BLE"
                            bleToggleButton.backgroundTintList = getColorStateList(R.color.button_primary_background)
                        } else {
                            Toast.makeText(this, "Failed to start BLE", Toast.LENGTH_LONG).show()
                            bleToggleButton.text = "Start BLE"
                            bleToggleButton.backgroundTintList = getColorStateList(R.color.secondary)
                        }
                    } catch (e: Exception) {
                        android.util.Log.e("MainActivity", "BLE start crashed in permissions callback: ${e.message}", e)
                        Toast.makeText(this, "BLE start crashed: ${e.message}", Toast.LENGTH_LONG).show()
                        bleToggleButton.text = "Start BLE"
                        bleToggleButton.backgroundTintList = getColorStateList(R.color.secondary)
                    }
                }
            } else {
                Toast.makeText(this, "BLE permissions denied", Toast.LENGTH_LONG).show()
            }
        }
    }
    
    private suspend fun handleIncomingBLEMessage() {
        bleManager?.getLastMessage()?.let { (filename, contentType, data) ->
            // Store BLE message in MQTT client for unified access
            mqttClient.cacheMessage(filename, contentType, data)
            
            // Show notification icon and play sound/vibrate
            runOnUiThread {
                showNotificationIcon()
                vibratePhone()
                playNotificationSound()
            }
        }
    }
    
    private fun handleErrorIntent() {
        val showError = intent.getBooleanExtra(EXTRA_SHOW_ERROR, false)
        val errorMessage = intent.getStringExtra(EXTRA_ERROR_MESSAGE)
        
        android.util.Log.d("MainActivity", "handleErrorIntent: showError=$showError, errorMessage=$errorMessage")
        
        if (showError) {
            android.util.Log.d("MainActivity", "Showing error icon and toast")
            showErrorIcon()
            if (errorMessage != null) {
                Toast.makeText(this, errorMessage, Toast.LENGTH_LONG).show()
            }
            
            // Clear the intent extras so error doesn't show again on config changes
            intent.removeExtra(EXTRA_SHOW_ERROR)
            intent.removeExtra(EXTRA_ERROR_MESSAGE)
        }
    }
    
    companion object {
        private const val EXTRA_SHOW_ERROR = "show_error"
        private const val EXTRA_ERROR_MESSAGE = "error_message"
        private const val BLE_PERMISSION_REQUEST_CODE = 123
        
        fun showError(context: Context, message: String) {
            android.util.Log.d("MainActivity", "showError called with message: $message")
            val intent = Intent(context, MainActivity::class.java).apply {
                putExtra(EXTRA_SHOW_ERROR, true)
                putExtra(EXTRA_ERROR_MESSAGE, message)
                flags = Intent.FLAG_ACTIVITY_CLEAR_TOP or Intent.FLAG_ACTIVITY_SINGLE_TOP
            }
            android.util.Log.d("MainActivity", "Starting activity with error intent")
            context.startActivity(intent)
        }
    }
}