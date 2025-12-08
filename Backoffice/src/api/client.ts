const baseUrl =
  import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8090/api";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${baseUrl}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {})
    },
    ...init
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(
      `API error ${res.status}: ${text || res.statusText || "Unknown error"}`
    );
  }

  return (await res.json()) as T;
}

export const apiClient = {
  get: <T>(path: string) => request<T>(path)
};


