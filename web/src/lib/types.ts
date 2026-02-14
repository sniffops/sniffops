export type RiskLevel = 'critical' | 'high' | 'medium' | 'low'

export interface Trace {
  id: string
  session_id: string
  timestamp: number
  tool_name: string
  command: string
  namespace: string
  resource_kind: string
  target_resource: string
  risk_level: RiskLevel
  risk_reason: string
  result: 'success' | 'error'
  latency_ms: number
  output?: string
  error_message?: string
  tokens_input?: number
  tokens_output?: number
  cost_estimate?: number
}

export interface TracesResponse {
  traces: Trace[]
  total: number
  limit: number
  offset: number
}

export interface RiskDistribution {
  critical: number
  high: number
  medium: number
  low: number
}

export interface ToolUsage {
  [toolName: string]: number
}

export interface TimelineEntry {
  hour: string
  count: number
}

export interface Stats {
  risk_distribution: RiskDistribution
  tool_usage: ToolUsage
  timeline: TimelineEntry[]
  total_operations: number
  total_cost_estimate: number
}

export interface TraceFilters {
  tool?: string
  namespace?: string
  risk?: RiskLevel
  limit?: number
  offset?: number
  start?: number
  end?: number
}
