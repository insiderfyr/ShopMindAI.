import type React from "react"
import "./globals.css"
import { Inter } from "next/font/google"

const inter = Inter({ subsets: ["latin"] })

// Metadata can't be used in client components, so we define it separately
const metadata = {
  title: "ShopMindAI - Your Personalized AI Shopping Assistant",
  description: "Skip the endless scrolling. Shop smarter with AI.",
    generator: 'v0.dev'
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <head>
        <title>{metadata.title}</title>
        <meta name="description" content={metadata.description} />
        <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1" />
      </head>
      <body className={inter.className}>{children}</body>
    </html>
  )
}
