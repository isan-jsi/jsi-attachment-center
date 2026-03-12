import type { ApiEnvelope, ApiError } from "@/types/api";

export class ApiClientError extends Error {
  constructor(
    public status: number,
    public statusText: string,
    public body: ApiError | null
  ) {
    super(body?.error ?? `${status} ${statusText}`);
    this.name = "ApiClientError";
  }
}

type TokenGetter = () => string | null;
let getToken: TokenGetter = () => null;

export function setTokenGetter(fn: TokenGetter) {
  getToken = fn;
}

async function request<T>(
  url: string,
  init?: RequestInit
): Promise<ApiEnvelope<T>> {
  const token = getToken();
  const headers = new Headers(init?.headers);

  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  if (!headers.has("Content-Type") && !(init?.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }

  const res = await fetch(url, { ...init, headers });

  if (!res.ok) {
    let body: ApiError | null = null;
    try {
      body = await res.json();
    } catch {
      // ignore parse errors
    }
    throw new ApiClientError(res.status, res.statusText, body);
  }

  if (res.status === 204) {
    return { data: undefined as unknown as T };
  }

  return res.json();
}

export const api = {
  get<T>(url: string): Promise<ApiEnvelope<T>> {
    return request<T>(url);
  },

  post<T>(url: string, body?: unknown): Promise<ApiEnvelope<T>> {
    return request<T>(url, {
      method: "POST",
      body: body instanceof FormData ? body : JSON.stringify(body),
    });
  },

  put<T>(url: string, body?: unknown): Promise<ApiEnvelope<T>> {
    return request<T>(url, {
      method: "PUT",
      body: JSON.stringify(body),
    });
  },

  patch<T>(url: string, body?: unknown): Promise<ApiEnvelope<T>> {
    return request<T>(url, {
      method: "PATCH",
      body: JSON.stringify(body),
    });
  },

  delete<T>(url: string): Promise<ApiEnvelope<T>> {
    return request<T>(url, { method: "DELETE" });
  },
};

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function buildQueryString(params: Record<string, any>): string {
  const searchParams = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== null && value !== "") {
      searchParams.set(key, String(value));
    }
  }
  const qs = searchParams.toString();
  return qs ? `?${qs}` : "";
}
