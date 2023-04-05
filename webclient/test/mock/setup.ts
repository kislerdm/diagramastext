import {afterEach, beforeAll, afterAll} from "vitest";

// mock server
import {setupServer} from 'msw/node'
import {rest} from 'msw'

import {ResponseSVG, Config} from "./../../src/ports";

const mockApiURL = "http://api-endpoint";

export const MockSVGResponse: ResponseSVG = {svg: "<svg><g><text>foo</text></g></svg>"},
    config: Config = {
        version: "foobar",
        urlAPI: mockApiURL,
        promptMinLength: 3,
        promptMaxLengthUserBase: 100,
        promptMaxLengthUserRegistered: 300,
    };

const handlers = [
    rest.post(mockApiURL, (req, res, ctx) => {
        return res(ctx.status(200), ctx.json(MockSVGResponse))
    }),
];

const server = setupServer(...handlers)

beforeAll(() => server.listen())
afterAll(() => server.close())
afterEach(() => server.resetHandlers())
