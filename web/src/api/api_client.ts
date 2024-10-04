export const BASE_URL: string = "http://localhost:8080/v1/";

export function post(url: string, body: any, header?: RequestInit) {
  return fetchWrapper(url, "POST", body, header);
}

export function put(url: string, body: any, header?: RequestInit) {
  return fetchWrapper(url, "PUT", body, header);
}

export function get(url: string, header?: RequestInit) {
  return fetchWrapper(url, "GET", header);
}

export function httpDelete(url: string, header?: RequestInit) {
  return fetchWrapper(url, "DELETE", header);
}

function fetchWrapper(url: string, method: string, body?: any, header?: RequestInit) {
  return fetch(url, {
    method,
    headers: {
      'Content-Type': 'application/json',
    },
    body: body,
    ...header,
  })
    .then(handleResponse)
    .catch(handleError)
}

async function handleResponse(response: Response) {
  if (response.status >= 400) {
    const contentType = response.headers.get("content-type");
    if (contentType && contentType.indexOf("application/json") !== -1) {
      return Promise.reject({ message: (await response.json()).message } as ErrorResponse);
    }

    return Promise.reject({ message: await response.text() || "An error occurred" } as ErrorResponse)
  }

  const contentType = response.headers.get("content-type");
  if (contentType && contentType.indexOf("application/json") !== -1) {
    return await response.json()
  }

  return response
}

export interface ErrorResponse {
  message: string
}

async function handleError(error: ErrorResponse) {
  console.error(JSON.stringify(error))
  return error
}