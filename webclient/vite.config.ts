import {defineConfig, mergeConfig} from 'vite'
import {defineConfig as defineConfigTesting} from 'vitest/config'

export default mergeConfig(
    defineConfig({
        base: "./",
        esbuild: {
            jsxFactory: "h",
            jsxFragment: "Fragment",
        },
        build: {
            minify: "terser",
        },
    }),
    defineConfigTesting({
        test: {
            include: ["./test/**/*.{ts,js}"],
        },
    }),
)