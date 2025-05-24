export default function QueryBubbles() {
  const examples = [
    "I need comfortable running sneakers under $100",
    "Find me skincare products for sensitive, dry skin",
    "What's the best laptop for graphic design?",
  ]

  return (
    <div className="w-full max-w-4xl mx-auto mb-16">
      <div className="flex flex-col md:flex-row justify-center gap-4 px-4">
        {examples.map((example, index) => (
          <div key={index} className="bg-white p-4 rounded-2xl shadow-md flex-1 min-w-0 border border-[#d8e8ff]">
            <p className="text-gray-800">{example}</p>
          </div>
        ))}
      </div>
    </div>
  )
}
