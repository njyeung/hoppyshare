'use client'

import { useAuth } from "@/contexts/AuthContext";
import { apiPost } from "@/lib/api";
import { useRouter } from "next/navigation";
import { userAgent } from "next/server";
import { useState, useEffect } from "react";

type os = "ANDROID" | "IOS"

export default function Mobile() {

  const [os, setOs] = useState<os>("ANDROID")
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const {user, loading, signInWithGoogle} = useAuth()
  const router = useRouter()
  function detectOS() {
    if (typeof window === 'undefined') return 'MACOS';
    
    const userAgent = window.navigator.userAgent;
    
    if (/Android/i.test(userAgent)) {
      setOs("ANDROID")
    }
    else if (/iPhone|iPad|iPod/i.test(userAgent)) {
      setOs("IOS")
    }
    else {
      router.push("/")
    }
  }

  const fetchDeviceCerts = async () => {
    if (!user || isLoading) return
    
    setIsLoading(true)
    setError(null)
    
    try {
      const response = await apiPost('https://en43r23fua.execute-api.us-east-2.amazonaws.com/prod/api/devices', {
        platform: os
      });
      
      console.log(response.status)
      const data = await response.json()
      
      if (response.status == 200) {
        const params = new URLSearchParams({
          cert: data.cert,
          key: data.key,
          ca: data.ca,
          group_key: data.group_key,
          device_id: data.device_id
        })
        
        console.log(params)
        window.location.href = `hoppyshare://setup?${params.toString()}`
      } else {
        setError(data.error || 'Failed to get device certificates')
      }
    } catch (err) {
      setError('Network error occurred')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(()=>{
    detectOS()
    if (!loading && !user) {
      sessionStorage.setItem('mobileSetup', 'true');
      signInWithGoogle()
    }
    else if (!loading && user) {
      // User is authenticated, fetch certificates
      fetchDeviceCerts()
    }

  }, [loading, user])
  
  if (loading || isLoading) {
    return (
      <div className="w-screen h-screen bg-white flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-secondary-darker mx-auto mb-4"></div>
          <p className="text-secondary-darker">
            {loading ? 'Checking authentication...' : 'Setting up device...'}
          </p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="w-screen h-screen bg-white flex items-center justify-center">
        <div className="text-center p-6">
          <div className="text-secondary mb-4">
            <svg className="w-12 h-12 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <h2 className="text-xl font-semibold">Setup Failed</h2>
          </div>
          <p className="text-secondary-dark mb-4">{error}</p>
          <button 
            onClick={() => {
              setError(null)
              fetchDeviceCerts()
            }}
            className="px-4 py-2 bg-secondary-darker text-white rounded hover:bg-opacity-80"
          >
            Try Again
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="w-screen h-screen bg-white flex items-center justify-center">
      <div className="text-center p-6">
        <div className="text-green-600 mb-4">
          <svg className="w-12 h-12 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <h2 className="text-xl font-semibold">Setup Complete</h2>
        </div>
        <p className="text-gray-600">
          Redirecting to HoppyShare app...
        </p>
      </div>
    </div>
  )
}