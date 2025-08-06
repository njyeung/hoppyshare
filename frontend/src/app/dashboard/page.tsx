'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Navbar from "@/components/Navbar";
import DeviceAccordion from "@/components/DeviceAccordion";
import WaveBackground from "@/components/WaveBackground";
import DeleteConfirmModal from "@/components/DeleteConfirmModal";
import { Device, DeviceSettings } from "@/types/device";
import { apiGet, apiPost } from "@/lib/api";

export default function Dashboard() {
  const router = useRouter();
  const [devices, setDevices] = useState<Device[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [deviceToDelete, setDeviceToDelete] = useState<Device | null>(null);

  const fetchDevices = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const response = await apiGet('https://en43r23fua.execute-api.us-east-2.amazonaws.com/prod/api/devices');
      
      if (!response.ok) {
        throw new Error('Failed to fetch devices');
      }
      
      const devicesData = await response.json();
      setDevices(devicesData.devices);
      
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
      console.error('Error fetching devices:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchDevices();
  }, []);

  const handleAddDevice = async () => {
    try {
      const response = await apiPost('https://en43r23fua.execute-api.us-east-2.amazonaws.com/prod/api/devices', {
        platform: 'WINDOWS'
      });
      
      if (!response.ok) {
        throw new Error('Failed to add device');
      }
      
      const responseData = await response.json();
      console.log('Add device API response:', responseData);
      
      // Refresh the devices list
      await fetchDevices();
      
    } catch (err) {
      console.error('Error adding device:', err);
      setError(err instanceof Error ? err.message : 'Failed to add device');
    }
  };

  const handleSettingsChange = (deviceId: string, newSettings: DeviceSettings) => {
    console.log('Settings change requested for device:', deviceId, newSettings);
  };

  const handleDeleteRequest = (device: Device) => {
    setDeviceToDelete(device);
    setShowDeleteModal(true);
  };

  const handleDeleteCancel = () => {
    setShowDeleteModal(false);
    setDeviceToDelete(null);
  };

  const handleDeleteConfirm = () => {
    if (deviceToDelete) {
      handleSettingsChange(deviceToDelete.deviceid, { ...deviceToDelete.settings, destroy: true });
    }
    setShowDeleteModal(false);
    setDeviceToDelete(null);
  };

  if (isLoading) {
    return (
      <>
        <Navbar />
        <section className="w-full bg-white min-h-screen flex justify-center">
          <div className="container mt-24 px-3 md:px-12 lg:px-28">
            <div className="flex justify-center items-center h-64">
              <div className="text-lg text-secondary-darker">Loading devices...</div>
            </div>
          </div>
        </section>
      </>
    );
  }

  if (error) {
    return (
      <>
        <Navbar />
        <section className="w-full bg-white min-h-screen flex justify-center">
          <div className="container mt-24 px-3 md:px-12 lg:px-28">
            <div className="flex justify-center items-center h-64">
              <div className="text-lg text-secondary-darker">Error loading devices. Please try again.</div>
            </div>
          </div>
        </section>
      </>
    );
  }

  return (
    <>
      <Navbar />
      <section className="relative w-full bg-white min-h-screen flex justify-center">
        <WaveBackground />
        <div className="container mt-24 px-3 md:px-12 lg:px-28 relative z-10">
          <div className="flex justify-between items-end mb-4">
            <h2 className="text-3xl text-secondary-darker">
              Devices
            </h2>
            <button 
              onClick={() => router.push('/add-device')}
              className="bg-secondary-light hover:bg-secondary text-white font-bold px-4 py-2 rounded-lg transition-colors"
            >
              +
            </button>
          </div>
          
          <div className="space-y-2">
            {devices.map((device) => (
              <DeviceAccordion
                key={device.deviceid}
                device={device}
                onSettingsChange={handleSettingsChange}
                onDeleteRequest={handleDeleteRequest}
              />
            ))}
          </div>
        </div>
      </section>
      
      <DeleteConfirmModal
        isOpen={showDeleteModal}
        onClose={handleDeleteCancel}
        onConfirm={handleDeleteConfirm}
        deviceName={deviceToDelete?.settings.nickname || ''}
      />
    </>
  )
}
