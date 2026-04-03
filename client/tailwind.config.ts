import type { Config } from "tailwindcss";

export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        sans: [
          "Inter",
          "-apple-system",
          "BlinkMacSystemFont",
          "Segoe UI",
          "sans-serif",
        ],
        mono: ["JetBrains Mono", "Fira Code", "Menlo", "monospace"],
      },
      colors: {
        // Primary surface scale
        surface: {
          DEFAULT: "#09090b", // zinc-950
          raised: "#18181b", // zinc-900
          overlay: "#27272a", // zinc-800
        },
        border: "#3f3f46", // zinc-700
        muted: "#71717a", // zinc-500
        subtle: "#a1a1aa", // zinc-400
        primary: "#e4e4e7", // zinc-200
        // Accent
        accent: {
          DEFAULT: "#6366f1", // indigo-500
          hover: "#818cf8", // indigo-400
          subtle: "#1e1b4b", // indigo-950
        },
        // Status
        success: "#22c55e",
        warning: "#f59e0b",
        error: "#ef4444",
      },
    },
  },
  plugins: [],
} satisfies Config;
