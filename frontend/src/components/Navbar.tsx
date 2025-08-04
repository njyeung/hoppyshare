'use client'

import Link from 'next/link';
import { useAuth } from '@/contexts/AuthContext';

export default function Navbar(){
  const { user, signOut } = useAuth();

  const handleLogout = async () => {
    await signOut();
  };

  return (
    <div className="bg-gradient-to-b from-primary/69 to-transparent backdrop-blur-md fixed top-0 h-[80px] w-full z-50">
      <nav className="container mx-auto px-6 h-full flex items-center justify-between">
        <Link href="/" className="flex items-center space-x-2">
          <span className="text-secondary-darker font-bold text-2xl">HoppyShare</span>
        </Link>
        
        <div className="flex items-center space-x-6">
          <Link 
            href="/dashboard" 
            className="text-secondary-darker hover:underline"
          >
            Dashboard
          </Link>
          <div className="flex items-center space-x-2">
            {user?.user_metadata?.avatar_url && (
              <img
                src={user.user_metadata.avatar_url}
                alt="Profile"
                className="w-6 h-6 rounded-full"
              />
            )}
            <button
              onClick={handleLogout}
              className="text-secondary-darker hover:underline"
            >
              Logout
            </button>
          </div>
        </div>
      </nav>
    </div>
  )
}