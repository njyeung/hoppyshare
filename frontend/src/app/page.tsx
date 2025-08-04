'use client'

import LandingNavbar from "@/components/LandingNavbar";
import Navbar from "@/components/Navbar";
import { useAuth } from "@/contexts/AuthContext";

export default function Home() {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  return (
    <>
      {user ? <Navbar /> : <LandingNavbar />}
    </>
  )
}
