import Auth from '@/components/Auth'
import LandingNavbar from '@/components/LandingNavbar'

export default function AuthPage() {
  return (
    <div className='min-w-screen min-h-screen bg-white'>
      <LandingNavbar />
      <Auth />
    </div>
  )
}