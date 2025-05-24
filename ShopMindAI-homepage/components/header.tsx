import Image from "next/image"
import Link from "next/link"

export default function Header() {
  return (
    <header className="w-full py-4 px-4 md:px-8 flex items-center justify-between">
      <div className="flex items-center">
        <Image src="/logo.svg" alt="ShopMind Logo" width={32} height={32} className="mr-2" />
        <h1 className="text-xl md:text-2xl font-bold">
          <span className="text-black">ShopMind</span>
          <span className="text-[#4d8eff]">AI</span>
        </h1>
      </div>

      <nav className="hidden md:flex items-center space-x-6">
        <Link href="#how-it-works" className="text-gray-800 hover:text-[#4d8eff] transition-colors">
          How It Works
        </Link>
        <Link href="#features" className="text-gray-800 hover:text-[#4d8eff] transition-colors">
          Features
        </Link>
        <Link href="#stores" className="text-gray-800 hover:text-[#4d8eff] transition-colors">
          Stores
        </Link>
        <Link
          href="#get-started"
          className="bg-[#4d8eff] text-white px-4 py-2 rounded-md font-medium hover:bg-[#3a6cd9] transition-colors"
        >
          Log in
        </Link>
      </nav>

      <button className="md:hidden text-gray-800">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="lucide lucide-menu"
        >
          <line x1="4" x2="20" y1="12" y2="12" />
          <line x1="4" x2="20" y1="6" y2="6" />
          <line x1="4" x2="20" y1="18" y2="18" />
        </svg>
      </button>
    </header>
  )
}
