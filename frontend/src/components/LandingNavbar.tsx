import Link from 'next/link';

export default function LandingNavbar(){
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
            Github
          </Link>
          <Link
            href={"/auth"}
            className="text-secondary-darker hover:underline"
          >
            Login
          </Link>
        </div>
      </nav>
    </div>
  )
}