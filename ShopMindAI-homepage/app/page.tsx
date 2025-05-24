"use client"

import { useState, useEffect } from "react"
import Image from "next/image"
import Link from "next/link"
import ThemeToggle from "@/components/theme-toggle"
import HowItWorksSection from "@/components/how-it-works-section"
import FeaturesSection from "@/components/features-section"


export default function Home() {
  const [darkMode, setDarkMode] = useState(false)
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    // Enable smooth scrolling
    document.documentElement.style.scrollBehavior = "smooth"

    // Check for dark mode preference
    const isDarkMode = document.documentElement.classList.contains("dark")
    setDarkMode(isDarkMode)
    setMounted(true)

    // Listen for changes to dark mode
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        if (mutation.attributeName === "class") {
          setDarkMode(document.documentElement.classList.contains("dark"))
        }
      })
    })

    observer.observe(document.documentElement, { attributes: true })

    return () => observer.disconnect()
  }, [])

  // Prevent hydration mismatch
  if (!mounted) return null

  return (
    <main className="min-h-screen flex flex-col bg-gradient-to-b from-[#e6f0ff] to-[#d8e8ff] dark:from-[#1a1a2e] dark:to-[#16213e] bg-transition relative">
      {/* Content Container */}
      <div className="relative z-10 flex flex-col min-h-screen">
        {/* Header */}
        <header className="w-full py-4 px-6 md:px-16 flex items-center justify-between sticky top-0 z-10 backdrop-blur-sm bg-opacity-90 dark:bg-opacity-90 bg-[#e6f0ff] dark:bg-[#1a1a2e] bg-transition">
          <div className="flex items-center">
            <div className="w-14 h-14 relative mr-3">
              <Image src="/logo.svg" alt="ShopMind Logo" fill className="object-contain" />
            </div>
            <h1 className="text-lg font-bold tracking-tight text-transition">
              <span className="text-black dark:text-white font-sans">ShopMind</span>
              <span className="text-black dark:text-white font-sans">AI</span>
            </h1>
          </div>

          <div className="flex items-center space-x-6">
            <nav className="hidden md:flex items-center space-x-8">
              {["How It Works", "Features", "Stores"].map((item, index) => (
                <Link
                  key={index}
                  href={`#${item.toLowerCase().replace(/\s+/g, "-")}`}
                  className="text-sm text-gray-800 dark:text-gray-200 hover:text-[#4d8eff] dark:hover:text-[#6d9eff] transition-colors duration-300 uppercase tracking-wide font-medium text-transition"
                >
                  {item}
                </Link>
              ))}
            </nav>

            <div className="flex items-center space-x-3">
              <ThemeToggle />

              <Link
                href="/auth"
                className="bg-[#4d8eff] text-white px-5 py-2 rounded-full text-sm font-medium hover:bg-[#3a6cd9] transition-colors duration-300 uppercase tracking-wide"
              >
                Log in
              </Link>
            </div>
          </div>
        </header>

        {/* Hero Section */}
        <section className="flex-grow flex flex-col items-center justify-center px-4 pt-10 pb-12">
          <div className="max-w-4xl mx-auto text-center">
            <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold mb-4 text-center tracking-tight leading-tight text-transition">
              <span className="text-black dark:text-white font-sans">ShopMind</span>
              <span className="text-black dark:text-white font-sans">AI</span>
            </h1>

            <h2 className="text-xl md:text-2xl font-semibold mb-3 text-gray-800 dark:text-gray-200 text-center leading-relaxed text-transition">
              Your Personalized AI Shopping Assistant
            </h2>

            <p className="text-base md:text-lg text-gray-700 dark:text-gray-300 mb-6 text-center max-w-2xl mx-auto leading-relaxed text-transition">
              Skip the endless scrolling. Shop smarter with AI.
            </p>

            <Link
              href="http://localhost:3000"
              target="_blank"
              rel="noopener noreferrer"
              className="bg-[#4d8eff] text-white px-8 py-3 rounded-full text-sm font-medium transition-colors duration-300 hover:bg-[#3a6cd9] inline-block uppercase tracking-wide"
            >
              Start Shopping
            </Link>
          </div>

          {/* Query Bubbles */}
          <div className="w-full max-w-5xl mx-auto mt-12 grid grid-cols-1 md:grid-cols-3 gap-6 px-4">
            {[
              "I need comfortable running sneakers under $100",
              "Find me skincare products for sensitive, dry skin",
              "What's the best laptop for graphic design?",
            ].map((query, index) => (
              <div
                key={index}
                className="bg-white dark:bg-gray-800 p-5 rounded-2xl shadow-sm hover:shadow-md flex items-start transition-shadow duration-300 h-full border border-gray-100 dark:border-gray-700 bg-transition"
              >
                <div className="w-8 h-8 bg-[#4d8eff] rounded-full flex items-center justify-center flex-shrink-0 mr-3">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path
                      d="M12 22C17.5228 22 22 17.5228 22 12C22 6.47715 17.5228 2 12 2C6.47715 2 2 6.47715 2 12C2 13.5997 2.37562 15.1116 3.04346 16.4525C3.22094 16.8088 3.28001 17.2161 3.17712 17.6006L2.58151 19.8267C2.32295 20.793 3.20701 21.677 4.17335 21.4185L6.39939 20.8229C6.78393 20.72 7.19121 20.7791 7.54753 20.9565C8.88837 21.6244 10.4003 22 12 22Z"
                      stroke="white"
                      strokeWidth="2"
                      strokeLinejoin="round"
                    />
                    <path d="M8 10.5H16" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                    <path d="M8 14H13.5" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                  </svg>
                </div>
                <p className="text-sm text-gray-800 dark:text-gray-200 self-center text-transition">{query}</p>
              </div>
            ))}
          </div>
        </section>

        {/* How It Works Section */}
        <HowItWorksSection />

        {/* Features Section */}
        <FeaturesSection />


        {/* Footer */}
        <footer className="w-full py-5 text-center">
          <div className="max-w-4xl mx-auto">
            <nav className="mb-3">
              <ul className="flex flex-wrap justify-center gap-3 text-gray-500 dark:text-gray-400 text-xs">
                {["Terms of Service", "Privacy Policy", "Cookie Policy", "Accessibility", "Contact", "FAQ"].map(
                  (item, index, arr) => (
                    <li key={index} className="flex items-center">
                      <Link
                        href={`#${item.toLowerCase().replace(/\s+/g, "-")}`}
                        className="hover:text-[#4d8eff] transition-colors duration-300 uppercase tracking-wide text-transition"
                      >
                        {item}
                      </Link>
                      {index < arr.length - 1 && <span className="mx-3">•</span>}
                    </li>
                  ),
                )}
              </ul>
            </nav>
            <p className="text-gray-500 dark:text-gray-400 text-xs text-transition">
              © 2025 ShopMindAI. All rights reserved.
            </p>
          </div>
        </footer>
      </div>
    </main>
  )
}
