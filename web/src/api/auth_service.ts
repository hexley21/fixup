import * as apiClient from "./api_client";
import { joinUrlWithRoute } from "../util/url_util";

const baseUrl = joinUrlWithRoute(apiClient.BASE_URL, "/auth");

export function registerCustomer(body: any) {
    return apiClient.post(joinUrlWithRoute(baseUrl, "register/customer"), body);
}

export function registerProvider(body: any) {
    return apiClient.post(joinUrlWithRoute(baseUrl, "register/provider"), body);
}

export function loginUser(body: any) {
    return apiClient.post(joinUrlWithRoute(baseUrl, "login"), body)
} ``