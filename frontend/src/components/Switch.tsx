interface SwitchProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
  label: string;
  description: string;
  variant?: 'default' | 'danger';
}

export default function Switch({ checked, onChange, label, description, variant = 'default' }: SwitchProps) {
  const bgColor = variant === 'danger' 
    ? checked ? 'bg-secondary' : 'bg-primary-muted'
    : checked ? 'bg-primary' : 'bg-primary-muted';

  return (
    <div className="flex items-center gap-3">
      <button
        type="button"
        onClick={() => onChange(!checked)}
        className={`relative inline-flex h-6 w-11 flex-shrink-0 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-secondary ${bgColor}`}
      >
        <span
          className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
            checked ? 'translate-x-6' : 'translate-x-1'
          }`}
        />
      </button>
      <div className="flex-1 min-w-0">
        <span className="text-sm text-secondary-darker font-medium">{label}</span>
        <p className="text-xs text-secondary-muted">{description}</p>
      </div>
    </div>
  );
}