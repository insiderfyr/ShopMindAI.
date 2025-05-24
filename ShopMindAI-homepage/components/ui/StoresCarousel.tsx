"use client";

import React from "react";

// Example SVG logos (replace/add as needed)
const logos = [
  // Cerc
  (
    <svg viewBox="0 0 40 40" fill="currentColor" className="w-12 h-12 text-black dark:text-black" xmlns="http://www.w3.org/2000/svg">
      <circle cx="20" cy="20" r="16" />
    </svg>
  ),
  // Pătrat
  (
    <svg viewBox="0 0 40 40" fill="currentColor" className="w-12 h-12 text-black dark:text-black" xmlns="http://www.w3.org/2000/svg">
      <rect x="8" y="8" width="24" height="24" rx="4" />
    </svg>
  ),
  // Triunghi
  (
    <svg viewBox="0 0 40 40" fill="currentColor" className="w-12 h-12 text-black dark:text-black" xmlns="http://www.w3.org/2000/svg">
      <polygon points="20,6 36,34 4,34" />
    </svg>
  ),
  // Stea
  (
    <svg viewBox="0 0 40 40" fill="currentColor" className="w-12 h-12 text-black dark:text-black" xmlns="http://www.w3.org/2000/svg">
      <polygon points="20,6 24,16 35,16 26,23 29,34 20,27 11,34 14,23 5,16 16,16" />
    </svg>
  ),
  // Inimă
  (
    <svg viewBox="0 0 40 40" fill="currentColor" className="w-12 h-12 text-black dark:text-black" xmlns="http://www.w3.org/2000/svg">
      <path d="M20 34s-12-8.35-12-16.5S13.5 4 20 11.5 32 4 32 17.5 20 34 20 34z" />
    </svg>
  ),
];

function duplicate(arr: React.ReactNode[], times: number) {
  return Array(times)
    .fill(null)
    .flatMap(() => arr);
}

const ROW_REPEAT = 2; // Duplicate logos for seamless scroll

export const StoresCarousel: React.FC = () => {
  return (
    <div className="w-full py-8">
      <div className="flex flex-col gap-4">
        {[false, true].map((reverse, idx) => (
          <div
            key={idx}
            className={`overflow-hidden w-full`}
          >
            <div
              className={`flex gap-6 min-w-max items-center animate-${reverse ? "scroll-reverse" : "scroll"} hover:[animation-play-state:paused]`}
            >
              {duplicate(logos, ROW_REPEAT).map((logo, i) => (
                <div
                  key={i}
                  className="bg-white rounded-xl shadow-md flex items-center justify-center p-4 min-w-[96px] min-h-[96px]"
                >
                  {logo}
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default StoresCarousel; 