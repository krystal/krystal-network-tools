import getApiUrl, { ServerLocation } from "./get-api-url";

export type RequestOptions = RequestInit & {
  location?: ServerLocation;
};

export const DEFAULT_REQUEST_OPTIONS: RequestOptions = {
  headers: {
    "Content-Type": "application/json",
  },
};

const request = async <T>(
  endpoint: string,
  requestOptions: RequestOptions = {}
) => {
  const { location, ...options } = requestOptions;

  const url = getApiUrl(endpoint, location);

  const res = await fetch(url, { ...DEFAULT_REQUEST_OPTIONS, ...options });
  if (!res.ok) throw new Error(res.statusText);
  const data = await res.json();
  return data as T;
};

export default request;
