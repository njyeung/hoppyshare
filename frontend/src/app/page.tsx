'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';
import LandingNavbar from '@/components/LandingNavbar';
import Hero from '@/components/Hero';
import Features from '@/components/Features';
import HowItWorks from '@/components/HowItWorks';
import Footer from '@/components/Footer';
import WaveBackground from '@/components/WaveBackground';

export default function HomePage() {
  const router = useRouter();
  const { user, loading } = useAuth();

  useEffect(() => {
    if (!loading && user) {
      router.replace('/dashboard');
    }
  }, [user, loading, router]);

  if (loading) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="text-lg text-secondary-darker">Loading...</div>
      </div>
    );
  }

  // Only show landing page if user is not authenticated
  if (user) {
    return null; // Will redirect to dashboard
  }

  return (
    <div className="bg-white min-h-screen">
      <LandingNavbar />
      <div className="relative">
        <Hero />
      </div>
      <Features />
      <HowItWorks />
      <Footer />
    </div>
  );
}