import { http } from './http'
import type {
  ListResponse,
  LoginResponse,
  SingleResponse,
} from './types'

interface RawParams {
  [key: string]: string | number | boolean | undefined
}

/** Clean params: drop empty strings/undefined so URLs stay readable. */
export function cleanParams(raw: RawParams): Record<string, string | number | boolean> {
  const out: Record<string, string | number | boolean> = {}
  for (const [key, value] of Object.entries(raw)) {
    if (value === undefined || value === '') continue
    out[key] = value
  }
  return out
}

export async function getSingle<T>(url: string, config?: { params?: RawParams }): Promise<T> {
  const res = (await http.get(url, config)) as unknown as SingleResponse<T>
  return res.data
}

export async function getList<T>(url: string, config?: { params?: RawParams }): Promise<ListResponse<T>> {
  return (await http.get(url, config)) as unknown as ListResponse<T>
}

export async function postSingle<T>(url: string, body: unknown): Promise<T> {
  const res = (await http.post(url, body)) as unknown as SingleResponse<T>
  return res.data
}

export async function patchSingle<T>(url: string, body: unknown): Promise<T> {
  const res = (await http.patch(url, body)) as unknown as SingleResponse<T>
  return res.data
}
