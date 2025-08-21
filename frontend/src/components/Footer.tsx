export default function Footer() {
  return (
    <footer className="bg-secondary-darker py-12">
      <div className="container px-6 md:px-12 lg:px-28 mx-auto">
        <div className="max-w-6xl mx-auto">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {/* Brand */}
            <div>
              <h3 className="text-2xl font-bold text-primary mb-4">HoppyShare</h3>
              <p className="text-primary-light mb-4">
                Privacy-first file and clipboard sharing across all your devices.
              </p>
              <p className="text-primary-muted text-sm">
                Built with enterprise-grade security and cross-platform compatibility.
              </p>
            </div>
            
            {/* Quick Links */}
            <div>
              <h4 className="text-lg font-semibold text-primary mb-4">Quick Links</h4>
              <ul className="space-y-2">
                <li><a href="#features" className="text-primary-light hover:text-primary transition-colors">Features</a></li>
                <li><a href="#howitworks" className="text-primary-light hover:text-primary transition-colors">How it works</a></li>
                <li><a href="/dashboard" className="text-primary-light hover:text-primary transition-colors">Dashboard</a></li>
                <li><a href="https://github.com/njyeung/hoppyshare" className="text-primary-light hover:text-primary transition-colors">Github</a></li>
                <li><a href="/faq" className="text-primary-light hover:text-primary transition-colors">FAQ</a></li>
              </ul>
            </div>
            
            {/* Technical */}
            <div>
              <h4 className="text-lg font-semibold text-primary mb-4">Technical</h4>
              <ul className="space-y-2">
                <li><span className="text-primary-light">End-to-End Encryption</span></li>
                <li><span className="text-primary-light">Self-Hosted Infrastructure</span></li>
                <li><span className="text-primary-light">Cross-Platform Native</span></li>
                <li><span className="text-primary-light">Offline BLE Support</span></li>
              </ul>
            </div>
          </div>
          
          {/* Bottom Border */}
          <div className="border-t border-secondary-dark mt-8 pt-8 text-center">
            <p className="text-primary-muted">
              HoppyShare. Built for snappy sharing with privacy and security.
            </p>
          </div>
        </div>
      </div>
    </footer>
  );
}