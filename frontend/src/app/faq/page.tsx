import LandingNavbar from "@/components/LandingNavbar";

export default function FAQ() {
  return <section className="flex w-full h-full bg-white justify-center">
    <LandingNavbar />

    <div className="mt-12 py-20 max-w-4xl w-full text-secondary-darker px-6">
      <h1 className="text-5xl w-full mb-16">Common Issues</h1>
      
      <div className="space-y-8">
        <div className="border-b border-gray-200 pb-6">
          <h2 className="text-2xl font-semibold mb-3">What if HoppyShare doesn't relocate itself from the downloads directory?</h2>
          <p className=" leading-relaxed mb-6">
            If the setup steps aren't followed correctly, HoppyShare may get stuck in a "half-installed" state. The fastest fix is to completely remove it and reinstall fresh.
          </p>
          
          <div className="ml-5 space-y-6">
            <div>
              <h3 className="text-lg font-semibold mb-3 text-secondary-darker">1. Clear stored credentials</h3>
              <ul className="space-y-2 ml-3">
                <li><strong>Windows:</strong> Open the windows search bar, type credential manager. Scroll until you find the HoppyShare entries and remove them all.</li>
                <li><strong>macOS:</strong> Open Keychain Access, search hoppyshare, find and delete the entry.</li>
                <li><strong>Linux:</strong> Delete the relevant .keyring from <code className="bg-gray-100 px-2 py-1 rounded">~/.local/share/keyrings/</code>. If it is locked, delete and recreate the keyring.</li>
              </ul>
            </div>

            <div>
              <h3 className="text-lg font-semibold mb-3 text-secondary-darker">2. Delete application files manually</h3>
              <ul className="space-y-2 ml-3">
                <li><strong>Windows:</strong> Remove <code className="bg-gray-100 px-2 py-1 rounded">%LOCALAPPDATA%\HoppyShare\HoppyShare.exe</code>.</li>
                <li><strong>macOS:</strong> Delete <code className="bg-gray-100 px-2 py-1 rounded">~/Library/Application Support/HoppyShare/HoppyShare</code>.</li>
                <li><strong>Linux:</strong> Remove <code className="bg-gray-100 px-2 py-1 rounded">~/.local/bin/hoppyshare</code>.</li>
              </ul>
            </div>

            <div>
              <h3 className="text-lg font-semibold mb-3 text-secondary-darker">3. Disable autostart entries</h3>
              <ul className="space-y-2 ml-3">
                <li><strong>Windows:</strong> Delete the shortcut in the Startup folder.</li>
                <li><strong>macOS:</strong> Remove the .plist in LaunchAgents or disable via launchctl.</li>
                <li><strong>Linux:</strong> Remove the .desktop file from <code className="bg-gray-100 px-2 py-1 rounded">~/.config/autostart/</code>.</li>
              </ul>
            </div>

            <div>
              <h3 className="text-lg font-semibold mb-3 text-secondary-darker">4. Reinstall cleanly</h3>
              <ul className="space-y-2 ml-3">
                <li>Download the latest HoppyShare binary, follow the setup steps for your OS closely, and you should be good to go.</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
}