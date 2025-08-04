'use client';

import { useState, useEffect } from 'react';
import Navbar from "@/components/Navbar";
import DeviceAccordion from "@/components/DeviceAccordion";
import { Device, DeviceSettings } from "@/types/device";
import { useApiQuery } from "@/lib/api";

export default function Dashboard() {
  const [devices, setDevices] = useState<Device[]>([
    {
      deviceid: "12345678-1234-1234-1234-123456789abc",
      settings: {
        nickname: "MacBook Pro",
        enabled: true,
        auto_copy: true,
        auto_paste: false,
        cache_time: 30,
        hotkey: "Cmd+Shift+V",
        enable_hotkey: true,
        notification_vol: 0.8,
        muted: false,
        send_to_self: true,
        auto_ble: false,
        startup: true,
        destroy: false
      }
    },
    {
      deviceid: "87654321-4321-4321-4321-cba987654321",
      settings: {
        nickname: "iPhone 15",
        enabled: true,
        auto_copy: false,
        auto_paste: true,
        cache_time: 60,
        hotkey: "",
        enable_hotkey: false,
        notification_vol: 1.0,
        muted: false,
        send_to_self: false,
        auto_ble: true,
        startup: false,
        destroy: false
      }
    },
    {
      deviceid: "abcdef12-3456-7890-abcd-ef1234567890",
      settings: {
        nickname: "Windows Desktop",
        enabled: false,
        auto_copy: true,
        auto_paste: true,
        cache_time: 15,
        hotkey: "Ctrl+Alt+V",
        enable_hotkey: true,
        notification_vol: 0.5,
        muted: true,
        send_to_self: true,
        auto_ble: false,
        startup: true,
        destroy: false
      }
    }
  ]);

  // Just test the query and log to console
  const { data: apiDevices, isLoading, error } = useApiQuery<Device[]>(
    ['devices'],
    'https://en43r23fua.execute-api.us-east-2.amazonaws.com/prod/api/devices'
  );

  useEffect(() => {
    console.log('API Query Results:', { apiDevices, isLoading, error });
  }, [apiDevices, isLoading, error]);

  const handleSettingsChange = (deviceId: string, newSettings: DeviceSettings) => {
    setDevices(devices.map(device => 
      device.deviceid === deviceId 
        ? { ...device, settings: newSettings }
        : device
    ));
  };

  return (
    <>
      <Navbar />
      <section className="w-full bg-white min-h-screen flex justify-center">
        <div className="container mt-24 px-3 md:px-12 lg:px-28">
          <div className="flex justify-between items-end mb-4">
            <h2 className="text-3xl text-secondary-darker">
              Devices
            </h2>
            <button className="bg-secondary hover:bg-secondary-dark text-white font-bold px-4 py-2 rounded-lg transition-colors">
              +
            </button>
          </div>
          
          <div className="space-y-2">
            {devices.map((device) => (
              <DeviceAccordion
                key={device.deviceid}
                device={device}
                onSettingsChange={handleSettingsChange}
              />
            ))}
          </div>
        </div>
      </section>
    </>
  )
}
