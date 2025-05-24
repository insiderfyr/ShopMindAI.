"use client"

import { useEffect, useRef } from "react"
import { Sparkles, Zap, Globe, Shield, Clock, Tag } from "lucide-react"

export default function FeaturesSection() {
  const sectionRef = useRef<HTMLElement>(null)

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add("opacity-100", "translate-y-0")
            entry.target.classList.remove("opacity-0", "translate-y-10")
          }
        })
      },
      { threshold: 0.1 },
    )

    const features = document.querySelectorAll(".feature-card")
    features.forEach((feature) => {
      observer.observe(feature)
    })

    return () => {
      features.forEach((feature) => {
        observer.unobserve(feature)
      })
    }
  }, [])

  const features = [
    {
      icon: <Sparkles className="w-8 h-8 text-[#4d8eff]" />,
      title: "AI-Powered Matches",
      description: "Find exactly what you need with intelligent product matching technology.",
    },
    {
      icon: <Zap className="w-8 h-8 text-[#4d8eff]" />,
      title: "Lightning Fast",
      description: "Get personalized recommendations in seconds, not hours of scrolling.",
    },
    {
      icon: <Globe className="w-8 h-8 text-[#4d8eff]" />,
      title: "Cross-Store Search",
      description: "Compare products across multiple retailers with a single query.",
    },
    {
      icon: <Shield className="w-8 h-8 text-[#4d8eff]" />,
      title: "Privacy First",
      description: "Shop confidently with our secure, privacy-focused platform.",
    },
    {
      icon: <Clock className="w-8 h-8 text-[#4d8eff]" />,
      title: "Time Saving",
      description: "Cut shopping time by 70% with our streamlined search experience.",
    },
    {
      icon: <Tag className="w-8 h-8 text-[#4d8eff]" />,
      title: "Best Deals",
      description: "Automatically find the best prices and active discounts across the web.",
    },
  ]

  return (
    <section id="features" ref={sectionRef} className="py-20 px-6 md:px-16 bg-gray-50 dark:bg-gray-900/50">
      <div className="max-w-6xl mx-auto">
        <h2 className="text-3xl md:text-4xl font-bold text-center mb-4 text-gray-800 dark:text-white text-transition">
          Why ShopMindAI?
        </h2>
        <p className="text-center text-gray-600 dark:text-gray-300 mb-16 max-w-2xl mx-auto text-transition">
          Our AI-powered shopping assistant makes finding the perfect products effortless.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {features.map((feature, index) => (
            <div
              key={index}
              className="feature-card opacity-0 translate-y-10 transition-all duration-700 ease-out bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm hover:shadow-md hover:border-blue-200 dark:hover:border-blue-700 border border-gray-100 dark:border-gray-700"
              style={{ transitionDelay: `${index * 100}ms` }}
            >
              <div className="mb-4 p-3 bg-blue-50 dark:bg-blue-900/30 rounded-lg inline-block">{feature.icon}</div>
              <h3 className="text-xl font-bold mb-2 text-gray-800 dark:text-white text-transition">{feature.title}</h3>
              <p className="text-gray-600 dark:text-gray-300 text-transition">{feature.description}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
