"use client"

import type React from "react"

import { useState, useEffect } from "react"
import Image from "next/image"
import Link from "next/link"

export default function AuthPage() {
  const [mounted, setMounted] = useState(false)
  const [formData, setFormData] = useState({
    emailOrPhone: "",
    password: "",
  })

  useEffect(() => {
    setMounted(true)
  }, [])

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }))
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    // Handle form submission here
    console.log("Form submitted:", formData)
  }

  if (!mounted) return null

  return (
    <main className="min-h-screen bg-gray-50 flex items-center justify-center px-6 md:px-8 py-8">
      <div className="w-full max-w-sm">
        <div className="bg-white rounded-2xl shadow-2xl px-8 py-8 w-full">
          {/* Logo and Brand */}
          <div className="flex justify-center items-center gap-2 mb-8 animate-in fade-in slide-in-from-bottom-4 duration-700">
            <div className="w-10 h-10 relative">
              <Image src="/logo.svg" alt="ShopMind Logo" fill className="object-contain" />
            </div>
            <h2 className="text-xl font-bold text-gray-800 font-sans">ShopMindAI</h2>
          </div>

          {/* Title and Subtitle */}
          <div className="text-center mb-6">
            <h1 className="text-lg font-medium text-gray-800 mb-1">Log in to your account</h1>
            <p className="text-sm text-gray-500">Enter your credentials to continue</p>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit} className="space-y-5">
            {/* Email/Phone */}
            <div>
              <label htmlFor="emailOrPhone" className="block text-sm font-medium text-gray-700 mb-1">
                Email or phone number
              </label>
              <input
                type="text"
                id="emailOrPhone"
                name="emailOrPhone"
                value={formData.emailOrPhone}
                onChange={handleInputChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-[#4d8eff] focus:border-transparent transition-all duration-200 bg-gray-50 focus:bg-white text-sm"
                placeholder="Enter your email or phone"
                required
              />
            </div>

            {/* Password */}
            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
                Password
              </label>
              <input
                type="password"
                id="password"
                name="password"
                value={formData.password}
                onChange={handleInputChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-[#4d8eff] focus:border-transparent transition-all duration-200 bg-gray-50 focus:bg-white text-sm"
                placeholder="Enter your password"
                required
              />
            </div>

            {/* Login Button */}
            <button
              type="submit"
              className="w-full bg-[#4d8eff] text-white py-2.5 px-4 rounded-lg font-medium hover:bg-[#3a6cd9] transition-colors duration-200 text-sm mt-2"
            >
              Log in
            </button>
          </form>

          {/* Links */}
          <div className="flex justify-between mt-4 text-sm">
            <Link href="/forgot-password" className="text-[#4d8eff] hover:underline">
              Forgot password?
            </Link>
            <Link href="/signup" className="text-[#4d8eff] hover:underline">
              Sign up
            </Link>
          </div>

          {/* Divider */}
          <div className="flex items-center my-6">
            <div className="flex-1 border-t border-gray-200"></div>
            <span className="px-4 text-xs text-gray-500">OR</span>
            <div className="flex-1 border-t border-gray-200"></div>
          </div>

          {/* Google Button */}
          <button
            type="button"
            className="w-full bg-white border border-gray-300 text-gray-700 py-2.5 px-4 rounded-lg text-sm font-medium hover:bg-gray-50 transition-colors duration-200 flex items-center justify-center space-x-2"
          >
            <svg className="w-4 h-4" viewBox="0 0 24 24">
              <path
                fill="#4285F4"
                d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
              />
              <path
                fill="#34A853"
                d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
              />
              <path
                fill="#FBBC05"
                d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
              />
              <path
                fill="#EA4335"
                d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
              />
            </svg>
            <span>Log in with Google</span>
          </button>

          {/* Terms */}
          <p className="text-xs text-gray-500 text-center mt-6 leading-relaxed">
            By continuing, you agree to our{" "}
            <Link href="/terms-of-service" className="text-[#4d8eff] hover:underline">
              Terms of Service
            </Link>{" "}
            and{" "}
            <Link href="/privacy" className="text-[#4d8eff] hover:underline">
              Privacy Policy
            </Link>
            .
          </p>
        </div>
      </div>
    </main>
  )
}
