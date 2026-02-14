import type { Trace, Stats, TracesResponse } from './types'

// Mock data for development
export const mockTraces: Trace[] = [
  {
    id: 'trace-001',
    session_id: 'session-abc',
    timestamp: Date.now() - 1000 * 60 * 5,
    tool_name: 'sniff_apply',
    command: 'kubectl apply -f deployment.yaml -n production',
    namespace: 'production',
    resource_kind: 'deployment',
    target_resource: 'nginx-deployment',
    risk_level: 'medium',
    risk_reason: 'Production namespace modification',
    result: 'success',
    latency_ms: 245,
    tokens_input: 120,
    tokens_output: 50,
    cost_estimate: 0.0012,
  },
  {
    id: 'trace-002',
    session_id: 'session-abc',
    timestamp: Date.now() - 1000 * 60 * 15,
    tool_name: 'sniff_delete',
    command: 'kubectl delete pod critical-pod -n kube-system',
    namespace: 'kube-system',
    resource_kind: 'pod',
    target_resource: 'critical-pod',
    risk_level: 'critical',
    risk_reason: 'Critical namespace deletion',
    result: 'success',
    latency_ms: 120,
    tokens_input: 95,
    tokens_output: 30,
    cost_estimate: 0.0008,
  },
  {
    id: 'trace-003',
    session_id: 'session-xyz',
    timestamp: Date.now() - 1000 * 60 * 30,
    tool_name: 'sniff_get',
    command: 'kubectl get pods -n development',
    namespace: 'development',
    resource_kind: 'pod',
    target_resource: 'app-pods',
    risk_level: 'low',
    risk_reason: 'Read-only operation in dev namespace',
    result: 'success',
    latency_ms: 89,
    tokens_input: 75,
    tokens_output: 200,
    cost_estimate: 0.0015,
  },
  {
    id: 'trace-004',
    session_id: 'session-xyz',
    timestamp: Date.now() - 1000 * 60 * 45,
    tool_name: 'sniff_logs',
    command: 'kubectl logs nginx-7d9c8f -n production --tail=100',
    namespace: 'production',
    resource_kind: 'pod',
    target_resource: 'nginx-7d9c8f',
    risk_level: 'low',
    risk_reason: 'Read-only log access',
    result: 'success',
    latency_ms: 340,
    tokens_input: 85,
    tokens_output: 450,
    cost_estimate: 0.0028,
  },
  {
    id: 'trace-005',
    session_id: 'session-def',
    timestamp: Date.now() - 1000 * 60 * 60,
    tool_name: 'sniff_apply',
    command: 'kubectl apply -f service.yaml -n staging',
    namespace: 'staging',
    resource_kind: 'service',
    target_resource: 'api-service',
    risk_level: 'medium',
    risk_reason: 'Service configuration change',
    result: 'error',
    latency_ms: 150,
    error_message: 'Invalid service specification',
    tokens_input: 110,
    tokens_output: 40,
    cost_estimate: 0.001,
  },
  {
    id: 'trace-006',
    session_id: 'session-ghi',
    timestamp: Date.now() - 1000 * 60 * 90,
    tool_name: 'sniff_get',
    command: 'kubectl get nodes',
    namespace: 'default',
    resource_kind: 'node',
    target_resource: 'cluster-nodes',
    risk_level: 'low',
    risk_reason: 'Cluster-level read operation',
    result: 'success',
    latency_ms: 67,
    tokens_input: 60,
    tokens_output: 180,
    cost_estimate: 0.0013,
  },
  {
    id: 'trace-007',
    session_id: 'session-ghi',
    timestamp: Date.now() - 1000 * 60 * 120,
    tool_name: 'sniff_apply',
    command: 'kubectl apply -f configmap.yaml -n production',
    namespace: 'production',
    resource_kind: 'configmap',
    target_resource: 'app-config',
    risk_level: 'high',
    risk_reason: 'Production configuration change',
    result: 'success',
    latency_ms: 198,
    tokens_input: 130,
    tokens_output: 45,
    cost_estimate: 0.0011,
  },
]

export const mockStats: Stats = {
  risk_distribution: {
    critical: 3,
    high: 12,
    medium: 45,
    low: 96,
  },
  tool_usage: {
    sniff_get: 78,
    sniff_logs: 34,
    sniff_apply: 32,
    sniff_delete: 3,
    sniff_exec: 9,
  },
  timeline: Array.from({ length: 24 }, (_, i) => ({
    hour: new Date(Date.now() - (23 - i) * 60 * 60 * 1000).toISOString(),
    count: Math.floor(Math.random() * 20) + 2,
  })),
  total_operations: 156,
  total_cost_estimate: 0.0234,
}

export const mockNamespaces = [
  'default',
  'kube-system',
  'production',
  'staging',
  'development',
]

export const mockTools = [
  'sniff_get',
  'sniff_logs',
  'sniff_apply',
  'sniff_delete',
  'sniff_exec',
]

// Mock API implementation for development
export function getMockTracesResponse(filters: any = {}): TracesResponse {
  let filtered = [...mockTraces]
  
  if (filters.tool) {
    filtered = filtered.filter(t => t.tool_name === filters.tool)
  }
  
  if (filters.namespace) {
    filtered = filtered.filter(t => t.namespace === filters.namespace)
  }
  
  if (filters.risk) {
    filtered = filtered.filter(t => t.risk_level === filters.risk)
  }
  
  const limit = filters.limit || 50
  const offset = filters.offset || 0
  
  return {
    traces: filtered.slice(offset, offset + limit),
    total: filtered.length,
    limit,
    offset,
  }
}
