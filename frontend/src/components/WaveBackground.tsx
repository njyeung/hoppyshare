'use client';

export default function WaveBackground() {
  return (
    <div className="fixed bottom-0 left-0 w-full h-1/3 overflow-hidden pointer-events-none z-0">
      <svg
        className="absolute bottom-0 w-full h-full"
        viewBox="0 0 1200 120"
        preserveAspectRatio="none"
        style={{ width: '105%', left: '-5%' }}
      >
        <defs>
          <linearGradient id="waveGradient" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" className="text-secondary-darker" stopOpacity="0.8" />
            <stop offset="100%" className="text-secondary-darker" stopOpacity="1" />
          </linearGradient>
        </defs>
        
        {/* First wave layer */}
        <path
          d="M-100,60 C300,120 900,0 1300,60 L1300,120 L-100,120 Z"
          fill="url(#waveGradient)"
          className="animate-wave1"
        />
        
        {/* Second wave layer */}
        <path
          d="M-100,80 C400,140 800,20 1300,80 L1300,120 L-100,120 Z"
          fill="currentColor"
          className="text-secondary-darker opacity-60 animate-wave2"
        />
        
        {/* Third wave layer */}
        <path
          d="M-100,100 C200,160 1000,40 1300,100 L1300,120 L-100,120 Z"
          fill="currentColor"
          className="text-secondary-darker opacity-40 animate-wave3"
        />
      </svg>

      <style jsx>{`
        @keyframes wave1 {
          0%, 100% {
            transform: translateX(0);
          }
          50% {
            transform: translateX(-25px);
          }
        }

        @keyframes wave2 {
          0%, 100% {
            transform: translateX(0);
          }
          50% {
            transform: translateX(25px);
          }
        }

        @keyframes wave3 {
          0%, 100% {
            transform: translateX(0);
          }
          50% {
            transform: translateX(-15px);
          }
        }

        .animate-wave1 {
          animation: wave1 3s ease-in-out infinite;
        }

        .animate-wave2 {
          animation: wave2 5s ease-in-out infinite;
        }

        .animate-wave3 {
          animation: wave3 7s ease-in-out infinite;
        }
      `}</style>
    </div>
  );
}