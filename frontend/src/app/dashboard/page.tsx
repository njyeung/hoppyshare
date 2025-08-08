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
  const [expandedDevice, setExpandedDevice] = useState<string | null>(null);

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
      
      // Set first device as expanded by default
      if (devicesData.devices.length > 0 && !expandedDevice) {
        setExpandedDevice(devicesData.devices[0].deviceid);
      }
      
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

  const handleToggleExpansion = (deviceId: string) => {
    setExpandedDevice(expandedDevice === deviceId ? null : deviceId);
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
      <section className="relative w-full bg-white min-h-screen pb-24 flex justify-center">
        <WaveBackground />
        <div className="container mt-24 px-3 md:px-12 lg:px-28 relative z-10">
          {devices.length === 0 ? (
            // No devices - show centered connect button
            <div className="flex flex-col items-center justify-center min-h-[400px] space-y-6">
              <h2 className="text-3xl text-secondary-darker text-center">
                Welcome to HoppyShare
              </h2>
              <p className="text-secondary-muted text-center max-w-md">
                Get started by connecting your first device.
              </p>
              <button 
                onClick={() => router.push('/add-device')}
                className="group bg-white border-2 border-dashed border-secondary-light hover:border-solid hover:border-secondary hover:bg-secondary-light/10 rounded-lg px-8 py-6 transition-all duration-300 hover:scale-[102%]"
              >
                <div className="flex items-center space-x-4">
                  <div className="w-8 h-8 transition-transform group-hover:scale-110">
                    <img src="/connect.svg" alt="Connect device" className="w-full h-full" />
                  </div>
                  <div className="text-left">
                    <div className="text-lg font-semibold text-secondary-dark group-hover:text-secondary-darker transition-colors">Connect Your First Device</div>
                    <div className="text-sm text-secondary-muted group-hover:text-secondary transition-colors">Start building your HoppyShare network</div>
                  </div>
                </div>
              </button>
            </div>
          ) : (
            // Has devices - show normal layout
            <>
              <div className="flex justify-between items-center mb-4">
                <h2 className="text-3xl text-secondary-darker">
                  Devices
                </h2>
                <button 
                  onClick={() => router.push('/add-device')}
                  className="group bg-white border-2 border-dashed border-secondary-light hover:border-solid hover:border-secondary hover:bg-secondary-light/10 rounded-lg px-6 py-4 transition-all duration-300 hover:scale-[102%]"
                >
                  <div className="flex items-center space-x-3">
                    <div className="w-6 h-6 transition-transform group-hover:scale-110">
                      <img src="/connect.svg" alt="Connect device" className="w-full h-full" />
                    </div>
                    <div className="text-left">
                      <div className="text-sm font-semibold text-secondary-dark group-hover:text-secondary-darker transition-colors">Connect This Device</div>
                      <div className="text-xs text-secondary-muted group-hover:text-secondary transition-colors">Expand your HoppyShare network</div>
                    </div>
                  </div>
                </button>
              </div>
              
              <div className="space-y-2">
                {devices.map((device) => (
                  <DeviceAccordion
                    key={device.deviceid}
                    device={device}
                    onSettingsChange={handleSettingsChange}
                    onDeleteRequest={handleDeleteRequest}
                    isExpanded={expandedDevice === device.deviceid}
                    onToggleExpansion={handleToggleExpansion}
                  />
                ))}
              </div>
            </>
          )}
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
