import * as apiClient from "./api_client";
import { joinUrlWithRoute } from "../util/url_util";

const baseUrl = joinUrlWithRoute(apiClient.BASE_URL, "/auth");

export function getUsers() {
  const url = joinUrlWithRoute(baseUrl, "users");
  return apiClient.get(url, { credentials: "include" });
}
