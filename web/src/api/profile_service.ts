import * as apiClient from "./api_client";
import { joinUrlWithRoute } from "../util/url_util";

const baseUrl = joinUrlWithRoute(apiClient.BASE_URL, "/profile");

class profile {
  id!: string
  first_name!: string
  last_name!: string
  role!: string
  user_status!: boolean
  created_at!: Date
  picture_url!: string
  phone_number?: string
  email?: string
}

export function getProfile(id: string) {
  const url = joinUrlWithRoute(baseUrl, id);
  return (apiClient.get(url, {credentials: "include"}) as Promise<profile>)
}
