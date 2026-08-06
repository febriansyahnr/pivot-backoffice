package main

//go:generate go tool templ generate

//go:generate npx @tailwindcss/cli -i ./views/assets/css/input.css -o ./views/assets/css/tailwind.css --minify
