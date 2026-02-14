import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from './ui/dialog'
import { Badge } from './ui/badge'
import type { Trace, RiskLevel } from '@/lib/types'
import { format } from 'date-fns'
import { CheckCircle, XCircle, Clock, Coins, Database } from 'lucide-react'

interface TraceDetailProps {
  trace: Trace | null
  open: boolean
  onClose: () => void
}

const riskColors: Record<RiskLevel, string> = {
  critical: 'bg-red-600 hover:bg-red-600',
  high: 'bg-orange-600 hover:bg-orange-600',
  medium: 'bg-yellow-600 hover:bg-yellow-600',
  low: 'bg-green-600 hover:bg-green-600',
}

export function TraceDetail({ trace, open, onClose }: TraceDetailProps) {
  if (!trace) return null

  return (
    <Dialog open={open} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto" onClose={onClose}>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            Trace Details
            <Badge className={riskColors[trace.risk_level]}>
              {trace.risk_level}
            </Badge>
          </DialogTitle>
          <DialogDescription>
            {format(new Date(trace.timestamp), 'PPpp')}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Status */}
          <div className="flex items-center gap-2">
            {trace.result === 'success' ? (
              <>
                <CheckCircle className="h-5 w-5 text-green-600" />
                <span className="font-medium text-green-600">Success</span>
              </>
            ) : (
              <>
                <XCircle className="h-5 w-5 text-red-600" />
                <span className="font-medium text-red-600">Failed</span>
              </>
            )}
          </div>

          {/* Basic Info */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <div className="text-sm font-medium text-muted-foreground">Session ID</div>
              <div className="font-mono text-sm">{trace.session_id}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-muted-foreground">Trace ID</div>
              <div className="font-mono text-sm">{trace.id}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-muted-foreground">Tool</div>
              <div className="font-mono text-sm">{trace.tool_name}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-muted-foreground">Resource Kind</div>
              <div className="font-mono text-sm">{trace.resource_kind}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-muted-foreground">Namespace</div>
              <div className="font-mono text-sm">{trace.namespace}</div>
            </div>
            <div>
              <div className="text-sm font-medium text-muted-foreground">Target Resource</div>
              <div className="font-mono text-sm">{trace.target_resource}</div>
            </div>
          </div>

          {/* Command */}
          <div>
            <div className="text-sm font-medium text-muted-foreground mb-1">Command</div>
            <pre className="bg-muted p-3 rounded text-xs overflow-x-auto">
              {trace.command}
            </pre>
          </div>

          {/* Risk Reason */}
          <div>
            <div className="text-sm font-medium text-muted-foreground mb-1">Risk Assessment</div>
            <div className="bg-muted p-3 rounded text-sm">
              {trace.risk_reason}
            </div>
          </div>

          {/* Error Message */}
          {trace.error_message && (
            <div>
              <div className="text-sm font-medium text-red-600 mb-1">Error Message</div>
              <div className="bg-red-600/10 border border-red-600/20 p-3 rounded text-sm text-red-600">
                {trace.error_message}
              </div>
            </div>
          )}

          {/* Output */}
          {trace.output && (
            <div>
              <div className="text-sm font-medium text-muted-foreground mb-1">Output</div>
              <pre className="bg-muted p-3 rounded text-xs overflow-x-auto max-h-64">
                {trace.output}
              </pre>
            </div>
          )}

          {/* Metrics */}
          <div className="grid grid-cols-2 gap-4 pt-4 border-t">
            <div className="flex items-center gap-2">
              <Clock className="h-4 w-4 text-muted-foreground" />
              <div>
                <div className="text-xs text-muted-foreground">Latency</div>
                <div className="font-mono text-sm font-medium">{trace.latency_ms}ms</div>
              </div>
            </div>
            {trace.tokens_input !== undefined && (
              <div className="flex items-center gap-2">
                <Database className="h-4 w-4 text-muted-foreground" />
                <div>
                  <div className="text-xs text-muted-foreground">Tokens</div>
                  <div className="font-mono text-sm font-medium">
                    {trace.tokens_input} / {trace.tokens_output}
                  </div>
                </div>
              </div>
            )}
            {trace.cost_estimate !== undefined && (
              <div className="flex items-center gap-2">
                <Coins className="h-4 w-4 text-muted-foreground" />
                <div>
                  <div className="text-xs text-muted-foreground">Cost Estimate</div>
                  <div className="font-mono text-sm font-medium">
                    ${trace.cost_estimate.toFixed(4)}
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
