'use client'

import Link from 'next/link';
import Image from 'next/image';
import { useAuth } from '@/contexts/AuthContext';
import { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useRouter } from 'next/navigation';

export default function Navbar(){
  const { user, loading, signOut } = useAuth();
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const router = useRouter()
  
  useEffect(()=>{
    if (!loading && !user) {
      router.push("/auth")
    }
  }, [loading, user])

  const handleLogout = async () => {
    await signOut();
    setIsMobileMenuOpen(false);
  };

  const toggleMobileMenu = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen);
  };

  return (
    <>
      <div className="bg-gradient-to-b from-primary/69 to-transparent backdrop-blur-md fixed top-0 h-[80px] w-full z-40">
        <nav className="container mx-auto px-6 h-full flex items-center justify-between">
          <Link href="/" className="flex items-center space-x-2">
            <Image 
              src="/logo.png" 
              alt="HoppyShare" 
              width={180} 
              height={48}
              className="h-10 w-auto"
            />
          </Link>
          
          {/* Desktop Menu */}
          <div className="hidden sm:flex items-center space-x-6">
            <Link 
              href="/dashboard" 
              className="text-secondary-darker hover:underline"
            >
              Dashboard
            </Link>
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
                onClick={handleLogout}
                className="text-secondary-darker hover:underline hover:cursor-pointer"
              >
                Logout
              </button>
            </div>
          </div>

          {/* Mobile Menu Button */}
          <button
            onClick={toggleMobileMenu}
            className="sm:hidden flex flex-col justify-center items-center w-6 h-6 space-y-1"
          >
            <span className="block w-5 h-0.5 bg-secondary-darker transition-transform duration-300"></span>
            <span className="block w-5 h-0.5 bg-secondary-darker transition-opacity duration-300"></span>
            <span className="block w-5 h-0.5 bg-secondary-darker transition-transform duration-300"></span>
          </button>
        </nav>
      </div>

      {/* Mobile Menu Overlay & Menu */}
      <AnimatePresence>
        {isMobileMenuOpen && (
          <>
            {/* Overlay */}
            <motion.div 
              className="fixed inset-0 bg-secondary-muted z-50 sm:hidden"
              initial={{ opacity: 0 }}
              animate={{ opacity: 0.8 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.3 }}
              onClick={() => setIsMobileMenuOpen(false)}
            />

            {/* Mobile Menu */}
            <motion.div 
              className="fixed top-0 right-0 h-full w-56 bg-white shadow-lg z-50 sm:hidden"
              initial={{ x: '100%' }}
              animate={{ x: 0 }}
              exit={{ x: '100%' }}
              transition={{ type: 'tween', duration: 0.3 }}
            >
        <div className="flex flex-col h-full">
          {/* Header */}
          <div className="flex items-center justify-between p-6 border-b border-primary-muted">
            <h2 className="text-xl font-semibold text-secondary-darker">Menu</h2>
            <button
              onClick={() => setIsMobileMenuOpen(false)}
              className="text-secondary-darker hover:text-secondary-dark"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {/* Menu Items */}
          <div className="flex-1 py-6">
            <Link 
              href="/dashboard" 
              className="block px-6 py-3 text-secondary-darker transition-colors"
              onClick={() => setIsMobileMenuOpen(false)}
            >
              Dashboard
            </Link>
          </div>

          {/* User Profile Section */}
          <div className="border-t border-primary-muted p-6">
            <div className="flex items-center space-x-2 mb-3">
              {user?.user_metadata?.avatar_url ? (
                <img
                  src={user.user_metadata.avatar_url}
                  alt="Profile"
                  className="w-8 h-8 rounded-full"
                  referrerPolicy="no-referrer"
                />
              ) : (
                <div className="w-8 h-8 rounded-full bg-secondary-darker text-white flex items-center justify-center text-sm font-medium">
                  {user?.email?.charAt(0)?.toUpperCase() || 'U'}
                </div>
              )}
              <div>
                <p className="text-xs font-medium text-secondary-darker">{user?.email}</p>
              </div>
            </div>
            <button
              onClick={handleLogout}
              className="w-full text-left px-2 py-2 text-secondary-darker underline hover:bg-gray-50 rounded transition-colors"
            >
              Logout
            </button>
          </div>
        </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </>
  )
}