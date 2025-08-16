'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Image from 'next/image';
import Navbar from "@/components/Navbar";
import WaveBackground from "@/components/WaveBackground";
import { apiPost } from "@/lib/api";

type Platform = 'WINDOWS' | 'MACOS' | 'LINUX' | null;

export default function AddDevice() {
  const router = useRouter();
  const [selectedPlatform, setSelectedPlatform] = useState<Platform>(null);
  const [isLoading, setIsLoading] = useState(false);

  const detectOS = (): Platform => {
    if (typeof window === 'undefined') return 'MACOS'; // Default for SSR
    
    const userAgent = window.navigator.userAgent;
    const platform = window.navigator.platform;
    
    if (platform.toUpperCase().indexOf('MAC') >= 0 || userAgent.indexOf('Mac') >= 0) {
      return 'MACOS';
    } else if (platform.toUpperCase().indexOf('WIN') >= 0 || userAgent.indexOf('Windows') >= 0) {
      return 'WINDOWS';
    } else if (platform.toUpperCase().indexOf('LINUX') >= 0 || userAgent.indexOf('Linux') >= 0) {
      return 'LINUX';
    }
    
    return 'MACOS'; // Default fallback
  };

  useEffect(() => {
    setSelectedPlatform(detectOS());
  }, []);

  const handlePlatformSelect = (platform: Platform) => {
    setSelectedPlatform(platform);
  };

  async function downloadAndRename(url: string, filename: string) {
    const response = await fetch(url)
    if (!response.ok) throw new Error("Failed to fetch blob");

    const blob = await response.blob();
    
    const blobUrl = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = blobUrl
    link.download = filename
    document.body.appendChild(link)
    link.click()
  }

  const handleAddDevice = async () => {
    if (!selectedPlatform) return;

    try {
      setIsLoading(true);
      
      const response = await apiPost('https://en43r23fua.execute-api.us-east-2.amazonaws.com/prod/api/devices', {
        platform: selectedPlatform
      });
      
      if (!response.ok) {
        throw new Error(`Failed to add device: ${response.status} ${response.statusText}`);
      }
      
      const responseData = await response.json();
      
      // Download the binary with proper filename
      if (responseData.download_url) {
        const downloadUrl = responseData.download_url;
        const filename = getFilenameForPlatform(selectedPlatform);
        
        downloadAndRename(downloadUrl, filename)
      } else {
        console.log('No download_url in response');
      }
      
      // Redirect back to dashboard after download starts
      setTimeout(() => {
        router.push('/dashboard');
      }, 1000);
      
    } catch (err) {
      console.error('Error details:', err);
      console.error('Error message:', err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setIsLoading(false);
    }
  };

  const getFilenameForPlatform = (platform: Platform): string => {
    switch (platform) {
      case 'WINDOWS':
        return 'HoppyShare.exe';
      case 'MACOS':
        return 'HoppyShare';
      case 'LINUX':
        return 'HoppyShare';
      default:
        return 'HoppyShare';
    }
  };

  const getSetupInstructions = () => {
    switch (selectedPlatform) {
      case 'MACOS':
        return (
          <div className="bg-gray-50/90 p-6 rounded-lg">
            <h3 className="text-lg font-semibold text-secondary-darker mb-4">MacOS Setup Instructions</h3>
            <ol className="list-decimal list-inside space-y-2 text-secondary-dark">
              <li>Download the HoppyShare client for MacOS</li>
              <li>Open Terminal to the downloaded directory</li> 
              <li>Remove quarantine and start the executable:</li>
              <div className="bg-gray-900 mb-4 text-primary-light p-3 rounded mt-2 font-mono text-sm">
                xattr -d com.apple.quarantine ./HoppyShare && chmod +x ./HoppyShare && ./HoppyShare
              </div>
              <li>The application will start and appear in your system tray</li>
            </ol>
            <div className="mt-6 p-4 bg-blue-50 border-l-4 border-blue-400 rounded">
              <h4 className="font-semibold text-blue-800 mb-2">Installation Location</h4>
              <p className="text-blue-700 text-sm">
                The binary will relocate itself to: 
                <code className="bg-blue-100 px-2 py-1 rounded text-xs ml-2">
                  ~/Library/Application Support/HoppyShare/HoppyShare
                </code>
              </p>
            </div>
          </div>
        );
      case 'LINUX':
        return (
          <div className="bg-gray-50/90 p-6 rounded-lg">
            <h3 className="text-lg font-semibold text-secondary-darker mb-4">Linux Setup Instructions</h3>
            <ol className="list-decimal list-inside space-y-2 text-secondary-dark">
              <li>Download the HoppyShare client for Linux</li>
              <li>Extract the downloaded file to a directory of your choice</li>
              <li>Make the binary executable:</li>
              <div className="bg-gray-900 mb-4 text-primary-light p-3 rounded mt-2 font-mono text-sm">
                chmod +x HoppyShare
              </div>
              <li>Run the application:</li>
              <div className="bg-gray-900 mb-4 text-primary-light p-3 rounded mt-2 font-mono text-sm">
                ./HoppyShare
              </div>
              <li>The application will start and appear in your system tray</li>
              <p className='pl-7 text-sm text-primary-muted'>Note: System tray support requires a desktop environment like GNOME, KDE, or XFCE. Window managers like Hyprland may not display the system tray icon.</p>

            </ol>
            <div className="mt-6 p-4 bg-blue-50 border-l-4 border-blue-400 rounded">
              <h4 className="font-semibold text-blue-800 mb-2">Installation Location</h4>
              <p className="text-blue-700 text-sm">
                The binary will relocate itself to: 
                <code className="bg-blue-100 px-2 py-1 rounded text-xs ml-2">
                  ~/.local/bin/hoppyshare
                </code>
              </p>
            </div>
          </div>
        );
      case 'WINDOWS':
        return (
          <div className="bg-gray-50/90 p-6 rounded-lg">
            <h3 className="text-lg font-semibold text-secondary-darker mb-4">Windows Setup Instructions</h3>
            <ol className="list-decimal list-inside space-y-2 text-secondary-dark">
              <li>Download the HoppyShare client for Windows</li>
              <li>Open PowerShell as Administrator in the downloads folder and run:</li>
              <div className="bg-gray-900 mb-4 text-primary-light p-3 rounded mt-2 font-mono text-sm">
                Unblock-File -Path ".\HoppyShare.exe"<br/>
              </div>
              <li><strong>Important:</strong> Add antivirus exclusions for both locations:</li>
              <div className="bg-red-50 border-l-4 border-red-400 p-3 rounded mt-2 mb-2">
                <p className="text-red-800 text-sm font-semibold mb-2">Required Antivirus Exclusions:</p>
                <ul className="text-red-700 text-xs space-y-1 ml-4">
                  <li><strong>File:</strong><code className="bg-red-100 px-2 py-1 rounded">C:\Users\[YourUsername]\Downloads\HoppyShare.exe</code></li>
                  <li><strong>Folder:</strong> <code className="bg-red-100 px-2 py-1 rounded">%LOCALAPPDATA%\Local</code></li>
                </ul>
                <p className="text-red-600 text-xs mt-2 italic">Without these exclusions, the app will fail to install or run properly.</p>
              </div>
              <li>Right click the binary and run as administrator</li>
              <li>The application will start and appear in your system tray</li>
            </ol>
            <div className="mt-6 p-4 bg-blue-50 border-l-4 border-blue-400 rounded">
              <h4 className="font-semibold text-blue-800 mb-2">Installation Location</h4>
              <p className="text-blue-700 text-sm">
                The binary will relocate itself to: 
                <code className="bg-blue-100 px-2 py-1 rounded text-xs ml-2">
                  %LOCALAPPDATA%\HoppyShare\HoppyShare.exe
                </code>
              </p>
            </div>
            <div className="mt-4 p-4 bg-yellow-50 border-l-4 border-yellow-400 rounded">
              <h4 className="font-semibold text-yellow-800 mb-2">Antivirus Issues</h4>
              <p className="text-yellow-700 text-xs mb-2">
                If the file disappears or the app doesn't start, your antivirus may have blocked it:
              </p>
              <ul className="text-yellow-700 text-xs list-disc list-inside space-y-1">
                <li><strong>Windows Defender:</strong> Add an exclusion for %LOCALAPPDATA%\HoppyShare directory</li>
                <li><strong>Malwarebytes/other AV:</strong> Whitelist or exclude %LOCALAPPDATA%\HoppyShare directory</li>
              </ul>
            </div>
          </div>
        );
      default:
        return null;
    }
  };

  return (
    <>
      <Navbar />
      <section className="relative w-full bg-white min-h-screen flex justify-center">
        <WaveBackground />
        <div className="container mt-24 px-3 md:px-12 lg:px-28 relative z-10">
          <div className="max-w-4xl mx-auto">
            <div className="mb-8">
              <h1 className="text-3xl font-bold text-secondary-darker mb-2">New Device</h1>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 mb-8">
              {/* macOS */}
              <button
                onClick={() => handlePlatformSelect('MACOS')}
                className={`hover:cursor-pointer flex flex-col items-center p-6 rounded-xl border-2 transition-all duration-200 ${
                  selectedPlatform === 'MACOS'
                    ? 'border-primary bg-primary-light/90 text-secondary-darker'
                    : 'border-secondary-light hover:border-secondary-dark bg-white text-secondary-dark'
                }`}
              >
                <div className="w-24 h-24 mb-4 flex items-center justify-center">
                  <Image
                    src="/apple.svg"
                    alt="macOS"
                    width={48}
                    height={48}
                    className="w-full h-full"
                  />
                </div>
                <h3 className="text-xl font-semibold">macOS</h3>
                <p className="text-sm mt-2 opacity-80">Mac computers and laptops</p>
              </button>

              {/* Linux */}
              <button
                onClick={() => handlePlatformSelect('LINUX')}
                className={`hover:cursor-pointer flex flex-col items-center p-6 rounded-xl border-2 transition-all duration-200 ${
                  selectedPlatform === 'LINUX'
                    ? 'border-primary bg-primary-light/90 text-secondary-darker'
                    : 'border-secondary-light hover:border-secondary-dark bg-white text-secondary-dark'
                }`}
              >
                <div className="w-24 h-24 mb-4 flex items-center justify-center">
                  <Image
                    src="/linux.svg"
                    alt="Linux"
                    width={48}
                    height={48}
                    className="w-full h-full"
                  />
                </div>
                <h3 className="text-xl font-semibold">Linux</h3>
                <p className="text-sm mt-2 opacity-80">Ubuntu, Debian, Fedora, etc.</p>
              </button>

              {/* Windows */}
              <button
                onClick={() => handlePlatformSelect('WINDOWS')}
                className={`hover:cursor-pointer flex flex-col items-center p-6 rounded-xl border-2 transition-all duration-200 ${
                  selectedPlatform === 'WINDOWS'
                    ? 'border-primary bg-primary-light/90 text-secondary-darker'
                    : 'border-secondary-light hover:border-secondary-dark bg-white text-secondary-dark'
                }`}
              >
                <div className="w-24 h-24 mb-4 flex items-center justify-center">
                  <Image
                    src="/windows.svg"
                    alt="Windows"
                    width={48}
                    height={48}
                    className="w-full h-full"
                  />
                </div>
                <h3 className="text-xl font-semibold">Windows</h3>
                <p className="text-sm mt-2 opacity-80">Windows 10, 11, and later</p>
              </button>
            </div>

            {/* Setup Instructions */}
            {selectedPlatform && (
              <div className="mb-8">
                {getSetupInstructions()}
              </div>
            )}

            {/* Action Buttons */}
            <div className="flex justify-between items-center mb-12">
              <button
                onClick={() => router.push('/dashboard')}
                className="hover:cursor-pointer px-4 py-2 rounded-lg bg-primary-muted/80 hover:underline text-white transition-colors"
              >
                Back to Dashboard
              </button>

              <button
                onClick={handleAddDevice}
                disabled={!selectedPlatform || isLoading}
                className={`hover:cursor-pointer px-8 py-3 rounded-lg transition-all ${
                  selectedPlatform && !isLoading
                    ? 'bg-secondary hover:bg-secondary-dark text-white'
                    : 'bg-primary text-secondary-darker cursor-not-allowed'
                }`}
              >
                {isLoading ? 'Creating Binary...' : 'Download'}
              </button>
            </div>
          </div>
        </div>
      </section>
    </>
  );
}