import { defineConfig } from 'astro/config';
import react from '@astrojs/react';
import { nodePolyfills } from 'vite-plugin-node-polyfills';

// https://astro.build/config
export default defineConfig({
    integrations: [react()],
    server: {
        host: true,
        port: 4321,
    },
    vite: {
        plugins: [
            nodePolyfills({
                protocolImports: true,
            }),
        ],
        define: {
            global: 'globalThis',
        },
        optimizeDeps: {
            include: ['@solana/web3.js', 'buffer'],
        },
    },
});