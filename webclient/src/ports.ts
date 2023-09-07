import {FingerprintScanner, TokensStore} from "./ciam";
import {HTTPClient} from "./httpclient";

export declare type Config = {
    version: string
    urlAPI: string
    promptMinLength: number
    promptMaxLengthUserBase: number
    promptMaxLengthUserRegistered: number
    httpClientSVGRendering: HTTPClient
    cookieStore?: TokensStore
    fingerprintScanner?: FingerprintScanner
    httpClientCIAM?: HTTPClient
}

export type ResponseSVG = {
    svg: string;
}

export type ResponseError = {
    error: string;
}

export function IsResponseError(obj: object): obj is ResponseError {
    return "error" in obj;
}

export function IsResponseSVG(obj: object): obj is ResponseSVG {
    return "svg" in obj;
}
