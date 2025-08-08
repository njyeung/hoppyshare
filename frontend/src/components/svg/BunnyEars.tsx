interface BunnyEarsProps {
  className?: string;
  opacity?: number;
  inside?: string;
  outside?: string;
}

export default function BunnyEars({ className = "", inside="#f0b77e", outside="#f9f1f0"}: BunnyEarsProps) {
  return (
    <div className={`absolute pointer-events-none ${className}`}>
      <svg width="120" height="96" viewBox="0 0 120 96" fill="none">
        {/* Left ear */}
        <ellipse cx="36" cy="48" rx="16" ry="40" fill={outside} stroke="#603d3e" strokeWidth="1.5" transform="rotate(-25 36 48)" />
        <ellipse cx="36" cy="48" rx="8" ry="28" fill={inside} stroke="#603d3e" strokeWidth="1" transform="rotate(-25 36 48)" />
        
        {/* Right ear */}
        <ellipse cx="84" cy="48" rx="16" ry="40" fill={outside} stroke="#603d3e" strokeWidth="1.5" transform="rotate(25 84 48)" />
        <ellipse cx="84" cy="48" rx="8" ry="28" fill={inside} stroke="#603d3e" strokeWidth="1" transform="rotate(25 84 48)" />
      </svg>
    </div>
  );
}