import Link from "next/link"

export default function Footer() {
  const links = [
    { name: "Terms of Service", href: "#terms" },
    { name: "Privacy Policy", href: "#privacy" },
    { name: "Cookie Policy", href: "#cookies" },
    { name: "Accessibility", href: "#accessibility" },
    { name: "Contact", href: "#contact" },
    { name: "FAQ", href: "#faq" },
  ]

  return (
    <footer className="w-full py-8 px-4 md:px-8 text-center bg-[#f5f9ff]">
      <div className="max-w-4xl mx-auto">
        <nav className="mb-4">
          <ul className="flex flex-wrap justify-center gap-2 md:gap-4 text-gray-500">
            {links.map((link, index) => (
              <li key={index} className="flex items-center">
                <Link href={link.href} className="hover:text-[#4d8eff] transition-colors">
                  {link.name}
                </Link>
                {index < links.length - 1 && <span className="ml-2 md:ml-4">•</span>}
              </li>
            ))}
          </ul>
        </nav>
        <p className="text-gray-500 text-sm">© 2025 ShopMindAI. All rights reserved.</p>
      </div>
    </footer>
  )
}
