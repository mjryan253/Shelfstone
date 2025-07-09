import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	plugins: [sveltekit()],
	test: {
		globals: true,
		environment: 'jsdom',
		setupFiles: ['./src/setupTests.ts'], // if you have a setup file
		include: ['src/**/*.{test,spec}.{js,ts}'],
		// reporters: ['default', 'html'], // Optional: For HTML reports
	}
});
