import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table'
import { Badge } from './ui/badge'
import { Button } from './ui/button'
import type { Trace, RiskLevel } from '@/lib/types'
import { format } from 'date-fns'
import { CheckCircle, XCircle } from 'lucide-react'

interface TraceTimelineProps {
  traces: Trace[]
  onTraceClick: (trace: Trace) => void
  onLoadMore?: () => void
  hasMore?: boolean
}

const riskColors: Record<RiskLevel, string> = {
  critical: 'bg-red-600 hover:bg-red-600',
  high: 'bg-orange-600 hover:bg-orange-600',
  medium: 'bg-yellow-600 hover:bg-yellow-600',
  low: 'bg-green-600 hover:bg-green-600',
}

const riskEmojis: Record<RiskLevel, string> = {
  critical: '🔴',
  high: '🟠',
  medium: '🟡',
  low: '🟢',
}

export function TraceTimeline({ traces, onTraceClick, onLoadMore, hasMore }: TraceTimelineProps) {
  if (traces.length === 0) {
    return (
      <div className="text-center py-12 text-muted-foreground">
        No traces found. Try adjusting your filters.
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[140px]">Time</TableHead>
              <TableHead className="w-[80px]">Risk</TableHead>
              <TableHead className="w-[120px]">Tool</TableHead>
              <TableHead className="w-[140px]">Namespace</TableHead>
              <TableHead className="w-[160px]">Resource</TableHead>
              <TableHead>Command</TableHead>
              <TableHead className="w-[80px] text-center">Status</TableHead>
              <TableHead className="w-[80px] text-right">Latency</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {traces.map((trace) => (
              <TableRow
                key={trace.id}
                className="cursor-pointer"
                onClick={() => onTraceClick(trace)}
              >
                <TableCell className="font-mono text-xs">
                  {format(new Date(trace.timestamp), 'HH:mm:ss')}
                  <div className="text-muted-foreground">
                    {format(new Date(trace.timestamp), 'yyyy-MM-dd')}
                  </div>
                </TableCell>
                <TableCell>
                  <Badge className={riskColors[trace.risk_level]}>
                    {riskEmojis[trace.risk_level]} {trace.risk_level}
                  </Badge>
                </TableCell>
                <TableCell className="font-mono text-xs">
                  {trace.tool_name.replace('sniff_', '')}
                </TableCell>
                <TableCell className="font-mono text-xs">
                  {trace.namespace}
                </TableCell>
                <TableCell className="font-mono text-xs truncate max-w-[160px]">
                  {trace.target_resource}
                </TableCell>
                <TableCell className="truncate max-w-[300px]" title={trace.command}>
                  <code className="text-xs">{trace.command}</code>
                </TableCell>
                <TableCell className="text-center">
                  {trace.result === 'success' ? (
                    <CheckCircle className="h-4 w-4 text-green-600 inline" />
                  ) : (
                    <XCircle className="h-4 w-4 text-red-600 inline" />
                  )}
                </TableCell>
                <TableCell className="text-right font-mono text-xs">
                  {trace.latency_ms}ms
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {hasMore && (
        <div className="text-center">
          <Button onClick={onLoadMore} variant="outline">
            Load More
          </Button>
        </div>
      )}
    </div>
  )
}
