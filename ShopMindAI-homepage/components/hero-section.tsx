import Link from "next/link"

export default function HeroSection() {
  return (
    <section className="text-center max-w-4xl mx-auto mb-16">
      <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold mb-4 text-gray-900">
        Your Personalized AI Shopping Assistant
      </h1>
      <p className="text-xl md:text-2xl text-gray-700 mb-8">Skip the endless scrolling. Shop smarter with AI.</p>
      <Link
        href="#start-shopping"
        className="bg-[#4d8eff] text-white px-8 py-4 rounded-md text-lg font-medium hover:bg-[#3a6cd9] transition-colors inline-block"
      >
        Start Shopping
      </Link>
    </section>
  )
}
