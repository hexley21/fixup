import { trimStart, trimEnd } from "./string_util";

export function joinUrlWithRoute(url: string, route: string) {
  return trimEnd("/", url) + "/" + trimStart("/", route);
}