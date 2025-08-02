interface DeleteConfirmModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  deviceName: string;
}

export default function DeleteConfirmModal({ isOpen, onClose, onConfirm, deviceName }: DeleteConfirmModalProps) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-secondary-darker bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
        <h3 className="text-lg font-bold text-secondary-darker mb-4">
          Delete Device
        </h3>
        
        <p className="text-secondary-dark mb-2">
          Are you sure you want to delete <strong>{deviceName}</strong>?
        </p>
        
        <p className="text-sm text-secondary-muted mb-6">
          This action cannot be undone. The device will be permanently removed from your account and the HoppyShare client will gracefully remove itself from your device.
        </p>
        
        <div className="flex gap-3 group">
          <button
            type="button"
            onClick={onClose}
            className="flex-1 px-4 py-2 border border-secondary-dark text-secondary-darker rounded-lg 
            hover:bg-secondary-darker hover:text-white transition-colors focus:outline-none focus:ring-2 focus:ring-secondary peer"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={onConfirm}
            className="flex-1 px-4 py-2 bg-secondary-dark hover:bg-secondary-darker text-white rounded-lg 
            transition-colors focus:outline-none focus:ring-2 focus:ring-secondary-dark font-medium peer-hover:bg-white peer-hover:text-secondary-darker peer-hover:border peer-hover:border-secondary-dark"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  );
}