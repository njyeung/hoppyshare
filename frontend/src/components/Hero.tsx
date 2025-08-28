'use client';

import { motion } from 'framer-motion';
import { useRouter } from 'next/navigation';
import { useEffect, useRef, useState } from 'react';
export default function Hero() {
  const router = useRouter();
  const [spinnerIdx, setSpinnerIdx] = useState(1)
  const spinnerRef = useRef<null | NodeJS.Timeout>(null)
  
  useEffect(()=>{
    startSpinner()

    return ()=>{
      if (spinnerRef.current) {
        clearInterval(spinnerRef.current)
        spinnerRef.current = null
      }
    }
  }, [])

  const startSpinner = () => {
    spinnerRef.current = setInterval(()=>{
      setSpinnerIdx((prev)=> {
        console.log("asjkdajksd")
        if (prev >= 4) {
          return 1
        }
        return prev+1
      })
    }, 250)
  }
  return (
    <section className="relative bg-white flex justify-center overflow-hidden">
      
      <div className="container px-4 md:px-6 lg:px-12 relative z-10 mt-32 lg:mt-44">
        <div className="flex flex-col md:flex-row gap-6 justify-center">
          {/* Left Content */}
          <div className="text-left w-full md:max-w-md lg:max-w-lg xl:max-w-xl">
            
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
                onClick={() => router.push('/dashboard')}
                className="px-4 py-2 sm:px-5 sm:py-3 bg-secondary border-2 hover:bg-secondary-darker peer-hover:bg-white peer-hover:border-2 peer-hover:border-secondary peer-hover:text-secondary text-white rounded-lg text-lg font-semibold transition-all hover:cursor-pointer"
              >
                Start For Free
              </button>

            </div>
          </div>
          
          {/* Right Visual */}
          <div className='h-full flex justify-center items-center'>
            <motion.img className='pixel-art aspect-square
            w-full h-full max-w-80 max-h-80
            sm:max-w-none sm:max-h-none sm:h-72 sm:w-72
            md:w-64 md:h-64
            lg:w-72 lg:h-72
            xl:h-full
            ' src={`/loading-${spinnerIdx}.png`} alt="bunny mascot"></motion.img>
          </div>
        </div>
      </div>
    </section>
  );
}