'use client';

import { useState } from 'react';
import { Device, DeviceSettings } from '@/types/device';
import Switch from '@/components/Switch';
import { apiPut } from '@/lib/api';
import BunnyEars from '@/components/svg/BunnyEars';

interface DeviceAccordionProps {
  device: Device;
  onSettingsChange?: (deviceId: string, settings: DeviceSettings) => void;
  onDeleteRequest?: (device: Device) => void;
  isExpanded?: boolean;
  onToggleExpansion?: (deviceId: string) => void;
}

export default function DeviceAccordion({ device, onSettingsChange, onDeleteRequest, isExpanded = false, onToggleExpansion }: DeviceAccordionProps) {
  const [settings, setSettings] = useState(device.settings);
  const [isSaving, setIsSaving] = useState(false);

  const handleSettingChange = (key: keyof DeviceSettings, value: any) => {
    // cache_time max of 300 seconds (5 minutes)
    if (key === 'cache_time') {
      value = Math.min(Math.max(1, value), 300);
    }
    
    const newSettings = { ...settings, [key]: value };
    setSettings(newSettings);
  };

  const handleSaveChanges = async () => {
    try {
      setIsSaving(true);
      console.log('Saving settings for device:', device.deviceid, settings);
      
      const response = await apiPut(`https://en43r23fua.execute-api.us-east-2.amazonaws.com/prod/api/settings/${device.deviceid}`, {
        new_settings: settings
      });
      
      if (!response.ok) {
        throw new Error('Failed to save settings');
      }
      
      const responseData = await response.json();
      console.log('Settings saved successfully:', responseData);
      
      // Call the parent callback if provided
      onSettingsChange?.(device.deviceid, settings);
      
    } catch (error) {
      console.error('Error saving settings:', error);
    } finally {
      setIsSaving(false);
    }
  };

  const handleDeleteClick = () => {
    onDeleteRequest?.(device);
  };

  return (
    <div className="border border-secondary-dark rounded-lg mb-3 bg-white transition-all duration-300 ease-in-out">
      <div 
        className={`relative flex items-center justify-between p-4 cursor-pointer transition-all duration-200 overflow-hidden ${
          isExpanded ? 'bg-primary-light/40' : 'hover:bg-primary-light/20'
        }`}
        onClick={() => onToggleExpansion?.(device.deviceid)}
      >
        

        <div className="flex items-center space-x-3 relative z-10">
          <div className={`w-3 h-3 rounded-full transition-all duration-300 ${
            settings.enabled ? 'bg-primary' : 'bg-primary-muted'
          }`} />
          <span className={`text-secondary-darker transition-all duration-200 ${
            isExpanded ? 'font-semibold' : ''
          }`}>{settings.nickname}</span>
        </div>
        
        {/* Arrow */}
        <svg 
          className={`w-5 h-5 text-secondary-muted transition-all duration-300 ease-in-out relative z-10 ${
            isExpanded ? 'rotate-180' : ''
          }`}
          fill="none" 
          viewBox="0 0 24 24" 
          stroke="currentColor"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </div>

      <div className={`overflow-hidden transition-all duration-500 ease-in-out ${
        isExpanded ? 'max-h-[2000px] opacity-100' : 'max-h-0 opacity-0'
      }`}>
        <div className={`border-t border-secondary-dark p-4 transition-all duration-300 delay-75 ${
          isExpanded ? 'translate-y-0 opacity-100' : 'translate-y-[-10px] opacity-0'
        }`}>
          {/* Bunny Ears Background */}
          <BunnyEars inside="#de94a1" 
          className='opacity-30 -rotate-12 -z-10
          scale-200 -bottom-2 right-6 
          sm:scale-[300%] sm:bottom-0 sm:right-16 
          md:scale-[400%] md:bottom-6 md:right-28' />
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="text-sm text-secondary-darker mb-1">
                Device Nickname
              </label>
              <input
                type="text"
                value={settings.nickname}
                onChange={(e) => handleSettingChange('nickname', e.target.value)}
                className="w-full px-3 py-2 border border-secondary-darker rounded-lg text-secondary-dark focus:outline-none focus:ring-2 focus:ring-secondary"
              />
              <p className="text-xs text-secondary-muted mt-1">
                Custom name to identify this device
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary-darker mb-1">
                Cache Time (seconds)
              </label>
              <input
                type="number"
                min="1"
                max="300"
                value={settings.cache_time}
                onChange={(e) => handleSettingChange('cache_time', parseInt(e.target.value))}
                className="w-full px-3 py-2 border border-secondary-darker text-secondary-dark rounded-lg focus:outline-none focus:ring-2 focus:ring-secondary"
              />
              <p className="text-xs text-secondary-muted mt-1">
                How long to keep clipboard items in memory. Max 5 mins.
              </p>
            </div>
          </div>

          <div className="grid grid-cols-1 gap-4 mt-4">
            <Switch
              checked={settings.enabled}
              onChange={(checked) => handleSettingChange('enabled', checked)}
              label="Device Enabled"
              description="Listen for incoming messages"
            />

            <Switch
              checked={settings.auto_copy}
              onChange={(checked) => handleSettingChange('auto_copy', checked)}
              label="Auto Copy"
              description="Automatically copy content to clipboard when received"
            />

            <Switch
              checked={settings.light_animations}
              onChange={(checked) => handleSettingChange('light_animations', checked)}
              label="Light Animations"
              description="Simple icon animations"
            />

            <Switch
              checked={settings.muted}
              onChange={(checked) => handleSettingChange('muted', checked)}
              label="Muted"
              description="Disable notification sounds for this device"
            />

            <Switch
              checked={settings.send_to_self}
              onChange={(checked) => handleSettingChange('send_to_self', checked)}
              label="Send to Self"
              description="Device listens to messages from itself"
            />

            <Switch
              checked={settings.auto_ble}
              onChange={(checked) => handleSettingChange('auto_ble', checked)}
              label="Auto BLE"
              description="Bluetooth Low Energy automatically turns on when network loss is detected"
            />

            <Switch
              checked={settings.startup}
              onChange={(checked) => handleSettingChange('startup', checked)}
              label="Start with System"
              description="Automatically starts HoppyShare client when this system boots"
            />

            <div className="flex flex-col justify-between items-center gap-3 mt-7">
              <button 
                onClick={handleSaveChanges}
                disabled={isSaving}
                className={`rounded-lg transition-all p-3 text-white w-full max-w-[400px] font-medium hover:cursor-pointer ${
                  isSaving 
                    ? 'bg-primary text-white cursor-not-allowed' 
                    : 'bg-secondary-light hover:bg-secondary'
                }`}
              >
                {isSaving ? 'Saving...' : 'Save Changes'}
              </button>
              <div className="text-xs text-secondary-muted text-end w-full">
                Device ID: {device.deviceid}
              </div>
            </div>
          </div>

          <div className="mt-6 pt-6 pb-2 border-t border-secondary-dark">
            <button
              type="button"
              onClick={handleDeleteClick}
              className="hover:cursor-pointer px-6 py-2 bg-secondary-dark hover:bg-secondary-darker text-white rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-secondary-dark text-sm"
            >
              Delete Device
            </button>
            <p className="text-xs text-secondary-muted mt-2">
              This action cannot be undone. The device will be permanently removed from your account and the HoppyShare client will gracefully remove itself from your device.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}