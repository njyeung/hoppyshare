'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import Navbar from "@/components/Navbar";
import DeviceAccordion from "@/components/DeviceAccordion";
import WaveBackground from "@/components/WaveBackground";
import DeleteConfirmModal from "@/components/DeleteConfirmModal";
import { Device, DeviceSettings } from "@/types/device";
import { apiGet, apiDelete, apiPut } from "@/lib/api";

export default function Dashboard() {
  const router = useRouter();
  const queryClient = useQueryClient();
  
  const { data: devices = [], isLoading, error } = useQuery<Device[]>({
    queryKey: ['devices'],
    queryFn: async () => {
      const response = await apiGet('https://en43r23fua.execute-api.us-east-2.amazonaws.com/prod/api/devices');
      
      if (!response.ok) {
        throw new Error('Failed to fetch devices');
      }
      
      const devicesData = await response.json();
      return devicesData.devices;
    },
    staleTime: 0,
    gcTime: 1000 * 60 * 5,
  });

  const deleteDeviceMutation = useMutation({
    mutationFn: async (deviceId: string) => {
      console.log('Attempting to delete device with ID:', deviceId);
      const url = `https://en43r23fua.execute-api.us-east-2.amazonaws.com/prod/api/devices/${deviceId}`;
      console.log('DELETE URL:', url);
      
      const response = await apiDelete(url);
      console.log('Delete response status:', response.status);
      
      if (!response.ok) {
        const errorText = await response.text();
        console.error('Delete error response:', errorText);
        
        // Handle case where device is already deleted/revoked
        if (response.status === 400 && errorText.includes('already revoked')) {
          console.log('Device already deleted, treating as success');
          return { success: true, message: 'Device already deleted' };
        }
        
        throw new Error(`Failed to delete device: ${response.status} - ${errorText}`);
      }
      return response.json();
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['devices'] });
    },
  });

  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [deviceToDelete, setDeviceToDelete] = useState<Device | null>(null);
  const [expandedDevice, setExpandedDevice] = useState<string | null>(null);

  // Set first device as expanded by default when devices first load
  useEffect(() => {
    if (devices.length > 0 && !expandedDevice) {
      setExpandedDevice(devices[0].deviceid);
    }
  }, [devices.length]);

  const handleDeleteRequest = (device: Device) => {
    setDeviceToDelete(device);
    setShowDeleteModal(true);
  };

  const handleDeleteCancel = () => {
    setShowDeleteModal(false);
    setDeviceToDelete(null);
  };

  const handleDeleteConfirm = async () => {
    if (deviceToDelete) {
      // Set destroy flag to trigger client self-destruction
      try {
        await apiPut(`https://en43r23fua.execute-api.us-east-2.amazonaws.com/prod/api/settings/${deviceToDelete.deviceid}`, {
          new_settings: { ...deviceToDelete.settings, destroy: true }
        });
      } catch (error) {
        console.error('Error setting destroy flag:', error);
      }
      
      // Delete the device record
      await deleteDeviceMutation.mutateAsync(deviceToDelete.deviceid);
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
            <div className="flex flex-col items-center justify-center min-h-[400px] space-y-3">
              <h2 className="text-3xl text-secondary-darker text-center font-semibold">
                Welcome to HoppyShare
              </h2>
              <p className="text-secondary-muted text-center max-w-md mb-16">
                Get started by connecting your first device.
              </p>
              <button 
                onClick={() => router.push('/add-device')}
                className="group bg-white border-2 border-dashed border-secondary-light
                hover:border-solid hover:border-secondary hover:bg-secondary-light/10 rounded-lg 
                transition-all duration-300 hover:scale-[102%] px-8 py-6 sm:px-8 sm:py-6 hover:cursor-pointer"
              >
                {/* Mobile: Just icon in square */}
                <div className="flex items-center justify-center sm:hidden">
                  <div className="w-8 h-8 transition-transform group-hover:scale-110">
                    <img src="/connect.svg" alt="Connect device" className="w-full h-full" />
                  </div>
                </div>
                
                {/* Desktop: Full button with text */}
                <div className="hidden sm:flex items-center space-x-4 ">
                  <div className="w-8 h-8 transition-transform group-hover:scale-110">
                    <img src="/connect.svg" alt="Connect device" className="w-full h-full" />
                  </div>
                  <div className="text-left">
                    <div className="text-lg font-semibold text-secondary-dark group-hover:text-secondary-darker transition-colors">Connect This Device</div>
                    <div className="text-sm text-secondary-muted group-hover:text-secondary transition-colors">Start building your HoppyShare network</div>
                  </div>
                </div>
              </button>
            </div>
          ) : (
            // Has devices - show normal layout
            <>
              <div className="flex justify-between items-center mb-4">
                <h2 className="font-bold text-3xl text-secondary-darker">
                  Devices
                </h2>
                <button 
                  onClick={() => router.push('/add-device')}
                  className="group bg-white border-2 border-dashed 
                  border-secondary-light hover:border-solid hover:border-secondary 
                  hover:bg-secondary-light/10 rounded-lg transition-all duration-300 
                  hover:scale-[102%] px-3 py-3 sm:px-6 sm:py-4 hover:cursor-pointer"
                >
                  {/* Mobile: Just icon */}
                  <div className="flex items-center justify-center sm:hidden">
                    <div className="w-6 h-6 transition-transform group-hover:scale-110">
                      <img src="/connect.svg" alt="Connect device" className="w-full h-full" />
                    </div>
                  </div>
                  
                  {/* Desktop: Full button with text */}
                  <div className="hidden sm:flex items-center space-x-3">
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
