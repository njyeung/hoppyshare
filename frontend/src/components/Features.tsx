
'use client';

import { motion, useScroll, useTransform } from 'framer-motion';
import { Zap, ShieldCheck, MonitorSmartphone, Bluetooth, FileUp, Code2 } from 'lucide-react';
import { useRef } from 'react';

export default function Features() {
  const containerRef = useRef<HTMLDivElement>(null);
  const features = [
    {
      title: "Blazing Fast",
      blurb:
        "Instant handoff for text, screenshots, and files using a low-latency MQTT connection.",
      icon: Zap,
      kicker: "Sub-second feel, no cloud hop",
    },
    {
      title: "End-to-End Encrypted",
      blurb:
        "Your data is sealed before ever leaving your device. Only your devices hold the keys.",
      icon: ShieldCheck,
      kicker: "mTLS + E2EE by default",
    },
    {
      title: "True Cross-Platform",
      blurb:
        "Windows to macOS to Linux. Send a file on one device, save it on another. Copy an image on one device, paste natively on another.",
      icon: MonitorSmartphone,
      kicker: "Your clipboard, everywhere",
    },
    {
      title: "Offline via BLE",
      blurb:
        "No Wi-Fi? HoppyShare falls back to Bluetooth Low Energy for local, resilient transfer.",
      icon: Bluetooth,
      kicker: "Keeps working when networks don't",
    },
    {
      title: "Large File Support",
      blurb:
        "Send files up to 25MB — perfect for images, screen recordings, logs, decks, and artifacts.",
      icon: FileUp,
      kicker: "Optimized chunking, minimal wait",
    },
    {
      title: "Open Source",
      blurb:
        "Transparent and auditable. Self-host the broker, keep control of your infrastructure.",
      icon: Code2,
      kicker: "No lock-in, just code",
    },
  ];

  return (
    <section id="features" className="relative mt-10 md:mt-20 mb-20" ref={containerRef}>
      {/* Subtle grid background */}
      <div
        className="pointer-events-none absolute inset-0 [mask-image:radial-gradient(ellipse_at_center,black,transparent_70%)]"
        aria-hidden
      >
        <div className="absolute inset-0 bg-[linear-gradient(to_right,#bca4a4_1.5px,transparent_1.5px),linear-gradient(to_bottom,#bca4a4_1.5px,transparent_1.5px)] bg-[size:28px_28px] opacity-30" />
      </div>

      <div className="container mx-auto px-6 md:px-12 lg:px-28">
        {/* Header */}
        <div className="mx-auto max-w-3xl text-center mb-14">
          <h2 className="mt-4 text-2xl md:text-3xl font-bold tracking-tight text-secondary-darker">
            Why choose <span className="underline decoration-primary">HoppyShare</span>?
          </h2>
          <p className="mt-3 text-sm sm:text-base text-secondary-dark">
            A lightweight, cross‑platform handoff for the things you move all day — code, screenshots,
            snippets, and small files — all without the cloud detour.
          </p>
        </div>

        {/* Feature grid */}
        <div
          className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3"
        >
          {features.map(({ title, blurb, icon: Icon, kicker }, i) => {
            const cardRef = useRef<HTMLDivElement>(null);
            const { scrollYProgress: cardScrollProgress } = useScroll({
              target: cardRef,
              offset: ["center start", "center end"]
            });
            
            const gradientOpacity = useTransform(
              cardScrollProgress,
              [0.1, 0.5, 0.9],
              [0.1, 0.5, 0.1]
            );
            
            return (
            <motion.div
              key={title}
              ref={cardRef}
              className="group relative overflow-hidden rounded-2xl border border-primary-muted p-5 transition backdrop-blur-sm"
              style={{
                background: useTransform(
                  gradientOpacity,
                  (opacity) => `linear-gradient(135deg, rgba(240, 183, 126, ${opacity}) 0%, rgba(222, 148, 161, ${opacity}) 50%, rgba(255, 255, 255, ${opacity}) 100%)`
                )
              }}
            >

              <div className="relative flex items-start gap-4">
                {/* icon */}
                <div className="shrink-0">
                  <div className="grid h-12 w-12 place-items-center rounded-full border border-primary-muted bg-gradient-to-br from-secondary via-primary to-primary-light">
                    <Icon className="h-6 w-6 text-white" aria-hidden />
                  </div>
                </div>
                {/* text */}
                <div className="min-w-0">
                  <div className="text-xs font-mono uppercase text-secondary-muted">{kicker}</div>
                  <h3 className="mt-1 text-lg font-semibold text-secondary-darker">{title}</h3>
                  <p className="mt-1 text-sm leading-relaxed text-secondary-muted">{blurb}</p>
                </div>
              </div>
            </motion.div>
            );
          })}
        </div>

        {/* Footnote / trust line */}
        <p
          className="mx-auto mt-10 max-w-3xl text-center text-sm text-primary-muted"
        >
          Fast • Convenient • Safe
        </p>
      </div>
    </section>
  );
}
