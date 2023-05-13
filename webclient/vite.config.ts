import {defineConfig} from 'vite'
import {Schema, ValidateEnv} from "@julr/vite-plugin-validate-env";
import {parseRegexp} from "@vitest/utils";

const validateEnvFn = ValidateEnv({
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
});

// @ts-ignore
export default defineConfig(({mode}) => {
        const cfg = {
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
            test: {
                globals: true,
                environment: "jsdom",
                setupFiles: ['./test/mock/setup.ts'],
                include: ["./test/**/*.{ts,js}"],
                exclude: ["./test/mock"],
                css: {
                    include: parseRegexp("src\/(.*)css"),
                    modules: {
                        classNameStrategy: "non-scoped",
                    },
                },
                coverage: {
                    all: true,
                    provider: "c8",
                },
            },
            plugins: [],
        };

        if (mode !== "test") {
            // @ts-ignore
            cfg.plugins.push(validateEnvFn)
        }

        return cfg;
    }
)
