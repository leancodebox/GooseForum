/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
    "./templates/**/*.{html,gohtml}"
  ],
  theme: {
   
  },
  plugins: [require('daisyui')],
  daisyui: {
    themes: ["light", "dark", "retro", "cyberpunk", "valentine", "garden", "forest", "aqua", "luxury", "dracula"],
  },
}