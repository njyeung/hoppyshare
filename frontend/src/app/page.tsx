'use client'

import LandingNavbar from "@/components/LandingNavbar";
import Navbar from "@/components/Navbar";
import { useAuth } from "@/contexts/AuthContext";

export default function Home() {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-white">
        <div className="text-lg text-secondary-darker">Loading...</div>
      </div>
    );
  }

  return (
    <>
      {user ? <Navbar /> : <LandingNavbar />}
    </>
  )
}
