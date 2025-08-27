'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { supabase } from '@/lib/supabase'

export default function AuthCallback() {
  const router = useRouter()

  useEffect(() => {
    const handleAuthCallback = async () => {
      try {
        // Handle the OAuth callback
        const { data, error } = await supabase.auth.getSession()
        
        if (error) {
          console.error('Auth callback error:', error)
          router.push('/auth')
          return
        }

        if (data.session) {
          // Check if this was a mobile setup login
          const isMobileSetup = sessionStorage.getItem('mobileSetup');
          if (isMobileSetup) {
            alert("IS MOBILE")
            sessionStorage.removeItem('mobileSetup');
            router.push('/faq');
          } else {
            router.push('/dashboard');
          }
        } else {
          router.push('/auth')
        }
      } catch (error) {
        console.error('Unexpected error in auth callback:', error)
        router.push('/auth')
      }
    }

    handleAuthCallback()
  }, [router])

  return (
    <div className="min-h-screen flex items-center justify-center bg-white">
      <div className="text-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-secondary-darker mx-auto mb-4"></div>
        <p className="text-secondary-darker">Completing sign in...</p>
      </div>
    </div>
  )
}