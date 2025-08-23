'use client'

import Link from 'next/link';
import { useAuth } from '@/contexts/AuthContext';

export default function LandingNavbar(){
  const { user, signOut } = useAuth();

  return (
    <div className="bg-gradient-to-b from-primary/69 to-transparent backdrop-blur-md fixed top-0 h-[80px] w-full z-40">
      <nav className="container mx-auto px-6 h-full flex items-center justify-between">
        <Link href="/" className="flex items-center space-x-2">
          <span className="text-secondary-darker font-bold text-2xl">HoppyShare</span>
        </Link>       
        <div className="flex items-center space-x-6">
          {user && 
            <Link 
              href="/dashboard" 
              className="text-secondary-darker hover:underline"
            >
              Dashboard
            </Link>
          } 
          <Link 
            href="/faq" 
            className="text-secondary-darker hover:underline"
          >
            Issues
          </Link>
          <Link 
            href="https://github.com/njyeung/hoppyshare" 
            className="text-secondary-darker hover:underline"
          >
            Github
          </Link>
          {user ? (
            <div className="flex items-center space-x-2">
              {user?.user_metadata?.avatar_url ? (
                <img
                  src={user.user_metadata.avatar_url}
                  alt="Profile"
                  className="w-6 h-6 rounded-full"
                  referrerPolicy="no-referrer"
                  onError={(e) => {
                    e.currentTarget.style.display = 'none';
                    const fallback = e.currentTarget.nextElementSibling as HTMLElement;
                    if (fallback) {
                      fallback.style.display = 'flex';
                    }
                  }}
                />
              ) : null}
              <div 
                className="w-6 h-6 rounded-full bg-secondary-darker text-white flex items-center justify-center text-xs font-medium"
                style={{ display: user?.user_metadata?.avatar_url ? 'none' : 'flex' }}
              >
                {user?.email?.charAt(0)?.toUpperCase() || 'U'}
              </div>
              <button
                onClick={signOut}
                className="text-secondary-darker hover:underline"
              >
                Logout
              </button>
            </div>
          ) : (
            <Link
              href={"/auth"}
              className="text-secondary-darker hover:underline"
            >
              Login
            </Link>
          )}
        </div>
      </nav>
    </div>
  )
}