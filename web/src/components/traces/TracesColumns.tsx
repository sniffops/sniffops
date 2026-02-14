import { type ColumnDef } from '@tanstack/react-table'
import { ArrowUpDown, ShieldAlert, AlertTriangle, AlertCircle, Info, Check, X } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { type Trace, type RiskLevel } from '@/lib/types'
import { formatDistanceToNow } from 'date-fns'

const riskConfig = {
  critical: { label: 'Critical', variant: 'destructive' as const, icon: ShieldAlert },
  high: { label: 'High', variant: 'default' as const, icon: AlertTriangle },
  medium: { label: 'Medium', variant: 'secondary' as const, icon: AlertCircle },
  low: { label: 'Low', variant: 'outline' as const, icon: Info },
}

export const tracesColumns = (_setSelectedTrace: (trace: Trace) => void): ColumnDef<Trace>[] => [
  {
    accessorKey: 'timestamp',
    header: ({ column }) => {
      return (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          className="h-8 px-2"
        >
          Time
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      )
    },
    cell: ({ row }) => {
      const timestamp = row.getValue('timestamp') as number
      return (
        <div className="text-sm">
          {formatDistanceToNow(new Date(timestamp * 1000), { addSuffix: true })}
        </div>
      )
    },
  },
  {
    accessorKey: 'risk_level',
    header: 'Risk',
    cell: ({ row }) => {
      const risk = row.getValue('risk_level') as RiskLevel
      const config = riskConfig[risk]
      const Icon = config.icon
      return (
        <Badge variant={config.variant} className="gap-1">
          <Icon className="h-3 w-3" />
          {config.label}
        </Badge>
      )
    },
    filterFn: (row, id, value) => {
      return value.includes(row.getValue(id))
    },
  },
  {
    accessorKey: 'tool_name',
    header: ({ column }) => {
      return (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          className="h-8 px-2"
        >
          Tool
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      )
    },
    cell: ({ row }) => {
      return <div className="font-medium">{row.getValue('tool_name')}</div>
    },
  },
  {
    accessorKey: 'namespace',
    header: 'Namespace',
    cell: ({ row }) => {
      return <Badge variant="outline">{row.getValue('namespace')}</Badge>
    },
  },
  {
    accessorKey: 'target_resource',
    header: 'Resource',
    cell: ({ row }) => {
      const resource = row.getValue('target_resource') as string
      return <div className="max-w-[200px] truncate text-sm">{resource}</div>
    },
  },
  {
    accessorKey: 'command',
    header: 'Command',
    cell: ({ row }) => {
      const command = row.getValue('command') as string
      return (
        <div className="max-w-[300px] truncate text-sm font-mono text-muted-foreground">
          {command}
        </div>
      )
    },
  },
  {
    accessorKey: 'result',
    header: 'Status',
    cell: ({ row }) => {
      const result = row.getValue('result') as string
      return result === 'success' ? (
        <Badge variant="default" className="gap-1">
          <Check className="h-3 w-3" />
          Success
        </Badge>
      ) : (
        <Badge variant="destructive" className="gap-1">
          <X className="h-3 w-3" />
          Error
        </Badge>
      )
    },
  },
  {
    accessorKey: 'latency_ms',
    header: ({ column }) => {
      return (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          className="h-8 px-2"
        >
          Latency
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      )
    },
    cell: ({ row }) => {
      const latency = row.getValue('latency_ms') as number
      return <div className="text-sm">{latency}ms</div>
    },
  },
]
