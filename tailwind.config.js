/** @type {import('tailwindcss').Config} */
export default {
  content: ["./templates/**/*.templ"],
  theme: {
    fontFamily: {
      sans: ['Signika Negative'],
    },
    extend: {
      fontFamily: {
        "asdf": ["Signika Negative"],
      },
    },
  },
  plugins: [],
}
