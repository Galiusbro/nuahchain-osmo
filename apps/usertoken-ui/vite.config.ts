import react from "@vitejs/plugin-react";
import path from "path";
import { defineConfig } from "vite";

export default defineConfig({
    plugins: [
        react(),
    ],
    resolve: {
        alias: {
            "@": path.resolve(__dirname, "./src"),
        },
    },
    css: {
        postcss: {
            plugins: [
                require('@tailwindcss/postcss'),
                require('autoprefixer'),
            ],
        },
    },
    server: {
        port: 5173,
        strictPort: false,
        cors: true,
        proxy: {
            '/api': {
                target: 'http://localhost:26657',
                changeOrigin: true,
                rewrite: (path) => path.replace(/^\/api/, '')
            },
            '/rest': {
                target: 'http://localhost:1317',
                changeOrigin: true,
                rewrite: (path) => path.replace(/^\/rest/, '')
            }
        }
    }
});
