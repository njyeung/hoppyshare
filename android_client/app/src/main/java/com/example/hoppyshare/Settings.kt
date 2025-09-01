package com.example.hoppyshare

import android.content.Context
import android.content.SharedPreferences
import android.util.Log
import org.json.JSONArray
import org.json.JSONObject

data class AppSettings(
    val nickname: String = "Unnamed Device",
    val muted: Boolean = false,
    val sendToSelf: Boolean = true
)

object Settings {
    private const val PREFS_NAME = "hoppyshare_settings"
    private const val KEY_NICKNAME = "nickname"
    private const val KEY_MUTED = "muted"
    private const val KEY_SEND_TO_SELF = "send_to_self"
    
    private var currentSettings = AppSettings()
    private lateinit var sharedPrefs: SharedPreferences
    private var onSettingsChangedCallback: ((AppSettings) -> Unit)? = null
    
    fun initialize(context: Context) {
        sharedPrefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
        loadSettings()
    }
    
    fun getCurrentSettings(): AppSettings {
        return currentSettings
    }

    fun setOnSettingsChangedCallback(callback: (AppSettings) -> Unit) {
        onSettingsChangedCallback = callback
    }
    
    private fun loadSettings() {
        currentSettings = AppSettings(
            nickname = sharedPrefs.getString(KEY_NICKNAME, "Unnamed Device") ?: "Unnamed Device",
            muted = sharedPrefs.getBoolean(KEY_MUTED, false),
            sendToSelf = sharedPrefs.getBoolean(KEY_SEND_TO_SELF, true)
        )
        Log.d("Settings", "Loaded settings: $currentSettings")
    }
    
    private fun saveSettings() {
        sharedPrefs.edit().apply {
            putString(KEY_NICKNAME, currentSettings.nickname)
            putBoolean(KEY_MUTED, currentSettings.muted)
            putBoolean(KEY_SEND_TO_SELF, currentSettings.sendToSelf)
            apply()
        }
        Log.d("Settings", "Saved settings: $currentSettings")
    }
    
    fun parseSettingsFromMqtt(payload: String, deviceId: String): Boolean {
        return try {
            val jsonArray = JSONArray(payload)
            
            for (i in 0 until jsonArray.length()) {
                val deviceObj = jsonArray.getJSONObject(i)
                val deviceIdFromPayload = deviceObj.getString("deviceid")
                
                if (deviceIdFromPayload == deviceId) {
                    val settingsObj = deviceObj.getJSONObject("settings")
                    
                    var updated = false
                    
                    // Extract only the settings we care about for mobile
                    if (settingsObj.has("nickname")) {
                        val nickname = settingsObj.getString("nickname")
                        if (nickname != currentSettings.nickname) {
                            currentSettings = currentSettings.copy(nickname = nickname)
                            updated = true
                        }
                    }
                    
                    if (settingsObj.has("muted")) {
                        val muted = settingsObj.getBoolean("muted")
                        if (muted != currentSettings.muted) {
                            currentSettings = currentSettings.copy(muted = muted)
                            updated = true
                        }
                    }
                    
                    if (settingsObj.has("send_to_self")) {
                        val sendToSelf = settingsObj.getBoolean("send_to_self")
                        if (sendToSelf != currentSettings.sendToSelf) {
                            currentSettings = currentSettings.copy(sendToSelf = sendToSelf)
                            updated = true
                        }
                    }
                    
                    if (updated) {
                        saveSettings()
                        onSettingsChangedCallback?.invoke(currentSettings)
                        Log.d("Settings", "Updated settings for device $deviceId: $currentSettings")
                        return true
                    }
                    
                    break
                }
            }
            false
        } catch (e: Exception) {
            Log.e("Settings", "Failed to parse settings: ${e.message}", e)
            false
        }
    }
}