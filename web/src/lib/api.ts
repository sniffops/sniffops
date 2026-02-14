import type { TracesResponse, Trace, Stats, TraceFilters } from './types'

const API_BASE = '/api'

export async function fetchTraces(filters: TraceFilters = {}): Promise<TracesResponse> {
  const params = new URLSearchParams()
  
  Object.entries(filters).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      params.append(key, String(value))
    }
  })
  
  const url = `${API_BASE}/traces${params.toString() ? `?${params}` : ''}`
  const response = await fetch(url)
  
  if (!response.ok) {
    throw new Error(`Failed to fetch traces: ${response.statusText}`)
  }
  
  return response.json()
}

export async function fetchTraceById(id: string): Promise<Trace> {
  const response = await fetch(`${API_BASE}/traces/${id}`)
  
  if (!response.ok) {
    throw new Error(`Failed to fetch trace: ${response.statusText}`)
  }
  
  return response.json()
}

export async function fetchStats(period: string = '24h'): Promise<Stats> {
  const response = await fetch(`${API_BASE}/stats?period=${period}`)
  
  if (!response.ok) {
    throw new Error(`Failed to fetch stats: ${response.statusText}`)
  }
  
  return response.json()
}

export async function fetchNamespaces(): Promise<string[]> {
  const response = await fetch(`${API_BASE}/namespaces`)
  
  if (!response.ok) {
    throw new Error(`Failed to fetch namespaces: ${response.statusText}`)
  }
  
  return response.json()
}

export async function fetchTools(): Promise<string[]> {
  const response = await fetch(`${API_BASE}/tools`)
  
  if (!response.ok) {
    throw new Error(`Failed to fetch tools: ${response.statusText}`)
  }
  
  return response.json()
}
