import BunnyEars from "./svg/BunnyEars";
import WaveBackground from "./WaveBackground";

export default function HowItWorks() {
  const steps = [
    {
      step: "1",
      title: "Download & Install",
      description: "Get the native desktop client for your platform. Just one file — a lightweight, self-contained binary. Follow the quick steps and it sets itself up automatically."
    },
    {
      step: "2", 
      title: "Connect Your Devices",
      description: "Install on all your devices. Each device gets unique certificates for secure authentication."
    },
    {
      step: "3",
      title: "Start Sharing",
      description: "Copy, paste, and share files instantly. System tray integration makes it effortless."
    }
  ];

  return (
    <section id="howitworks" className="bg-primary-light/20 py-20 pb-36 relative overflow-clip">
      
      <div className="container px-6 md:px-12 lg:px-28 mx-auto relative">
        <div className="max-w-4xl mx-auto">

          <BunnyEars className="absolute scale-[250%] sm:scale-[300%] -bottom-42 left-20 xl:left-36 rotate-6 opacity-45" />

          {/* Section Header */}
          <div className="text-center mb-16">
            <h2 className="text-3xl sm:text-4xl font-bold text-secondary-darker mb-4">
              How It Works
            </h2>
            <p className="text-lg text-secondary-dark">
              Get up and running in minutes with a simple setup process.
            </p>
          </div>
          
          {/* Steps */}
          <div className="space-y-12">
            {steps.map((step, index) => (
              <div 
                key={index}
                className="flex flex-col sm:flex-row items-center gap-8"
              >
                {/* Step Number */}
                <div className="flex-shrink-0 w-16 h-16 bg-secondary rounded-full flex items-center justify-center">
                  <span className="text-2xl font-bold text-white">{step.step}</span>
                </div>
                
                {/* Step Content */}
                <div className="text-center sm:text-left">
                  <h3 className="text-2xl font-semibold text-secondary-darker mb-3">
                    {step.title}
                  </h3>
                  <p className="text-lg text-secondary-dark">
                    {step.description}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}