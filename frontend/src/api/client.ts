import type { ApiResponse } from "@/types/api"

function getApiBase(): string {
  const base = import.meta.env.VITE_API_BASE
  if (!base) {
    throw new Error("VITE_API_BASE is required")
  }
  return base
}

const API_BASE = getApiBase()

async function request<T>(
  path: string,
  options?: RequestInit
): Promise<ApiResponse<T>> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { "Content-Type": "application/json", ...options?.headers },
    ...options,
  })
  return res.json() as Promise<ApiResponse<T>>
}

export async function apiGet<T>(path: string): Promise<ApiResponse<T>> {
  return request<T>(path)
}

export async function apiPost<T>(
  path: string,
  body: unknown
): Promise<ApiResponse<T>> {
  return request<T>(path, {
    method: "POST",
    body: JSON.stringify(body),
  })
}

export async function apiPut<T>(
  path: string,
  body: unknown
): Promise<ApiResponse<T>> {
  return request<T>(path, {
    method: "PUT",
    body: JSON.stringify(body),
  })
}

export async function apiDelete<T>(path: string): Promise<ApiResponse<T>> {
  return request<T>(path, { method: "DELETE" })
}

export async function apiUpload<T>(
  path: string,
  file: File
): Promise<ApiResponse<T>> {
  const formData = new FormData()
  formData.append("file", file)

  const res = await fetch(`${API_BASE}${path}`, {
    method: "POST",
    body: formData,
  })
  return res.json() as Promise<ApiResponse<T>>
}

export async function apiHealth(): Promise<boolean> {
  try {
    const res = await fetch("/health", {
      cache: "no-store",
    })
    if (!res.ok) return false
    const body = await res.json()
    return body?.status === "ok"
  } catch {
    return false
  }
}
