import {defineConfig, mergeConfig} from 'vite'
import {defineConfig as defineConfigTesting} from 'vitest/config'
import {Schema, ValidateEnv} from "@julr/vite-plugin-validate-env";

export default mergeConfig(
    defineConfig({
        base: "./",
        esbuild: {
            jsxFactory: "h",
            jsxFragment: "Fragment",
        },
        server: {
            port: 9001,
            host: true,
        },
        build: {
            minify: "terser",
        },
        plugins: [
            ValidateEnv({
                VITE_URL_API: (key, value) => {
                    value = Schema.string({
                        format: "url",
                        protocol: true,
                        tld: false,
                    })(key, value)


                    const allowedHosts = [
                            "api.diagramastext.dev",
                            "api-stage.diagramastext.dev",
                            "localhost",
                        ],
                        split = value.split("://")[1];

                    if (!allowedHosts.filter((v) => split.startsWith(v)).length) {
                        throw new Error(`${key} is set to not supported value`)
                    }

                    return value
                },
            }),
        ]
    }),
    defineConfigTesting({
        test: {
            include: ["./test/**/*.{ts,js}"],
        },
    }),
)
