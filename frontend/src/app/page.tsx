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