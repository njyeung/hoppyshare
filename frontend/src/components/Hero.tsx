'use client';

import { useRouter } from 'next/navigation';
export default function Hero() {
  const router = useRouter();

  return (
    <section className="relative bg-white flex justify-center overflow-hidden">
      
      <div className="container px-4 lg:px-12 relative z-10 mt-32 lg:mt-44">
        <div className="flex flex-col md:flex-row gap-6">
          {/* Left Content */}
          <div className="text-left sm:max-w-sm md:max-w-md lg:max-w-lg xl:max-w-xl">
            
            {/* Main Heading */}
            <h1 className="text-4xl sm:text-5xl lg:text-6xl xl:text-7xl font-bold text-secondary-darker mb-4 sm:mb-6">
              One clipboard,
              <span className="block text-primary">all your devices.</span>
            </h1>
            
            {/* Subtitle */}
            <p className="text-sm sm:text-base xl:text-lg text-secondary-dark mb-8 ">
              HoppyShare gives you a seamless way to share text and files across Windows, MacOS, and Linux. Powered by direct connections and a secure MQTT backbone, it works instantly, even when WiFi drops.
            </p>
            
            
            {/* CTA Buttons */}
            <div className="flex flex-col sm:flex-row gap-2 justify-center md:justify-start">
              
              <button
                onClick={() => router.push('https://github.com/njyeung/hoppyshare')}
                className="peer order-2 px-4 py-2 sm:px-5 sm:py-3 border-2 border-secondary text-secondary hover:bg-secondary hover:text-white rounded-lg text-lg font-semibold transition-all hover:cursor-pointer"
              >
                Open Source
              </button>
              
              <button
                onClick={() => router.push('/auth')}
                className="px-4 py-2 sm:px-5 sm:py-3 bg-secondary border-2 hover:bg-secondary-darker peer-hover:bg-white peer-hover:border-2 peer-hover:border-secondary peer-hover:text-secondary text-white rounded-lg text-lg font-semibold transition-all hover:cursor-pointer"
              >
                Start For Free
              </button>

            </div>
          </div>
          
          {/* Right Visual */}

        </div>
      </div>
    </section>
  );
}