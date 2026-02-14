import { Select } from './ui/select'
import type { RiskLevel, TraceFilters } from '@/lib/types'

interface FilterBarProps {
  filters: TraceFilters
  onFiltersChange: (filters: TraceFilters) => void
  tools: string[]
  namespaces: string[]
}

export function FilterBar({ filters, onFiltersChange, tools, namespaces }: FilterBarProps) {
  const handleChange = (key: keyof TraceFilters, value: string) => {
    onFiltersChange({
      ...filters,
      [key]: value === '' ? undefined : value,
      offset: 0, // Reset pagination when filters change
    })
  }

  return (
    <div className="flex flex-col sm:flex-row gap-4 p-4 bg-card border rounded-lg">
      <div className="flex-1">
        <label className="text-sm font-medium mb-1 block">Tool</label>
        <Select
          value={filters.tool || ''}
          onChange={(e) => handleChange('tool', e.target.value)}
        >
          <option value="">All Tools</option>
          {tools.map((tool) => (
            <option key={tool} value={tool}>
              {tool}
            </option>
          ))}
        </Select>
      </div>

      <div className="flex-1">
        <label className="text-sm font-medium mb-1 block">Namespace</label>
        <Select
          value={filters.namespace || ''}
          onChange={(e) => handleChange('namespace', e.target.value)}
        >
          <option value="">All Namespaces</option>
          {namespaces.map((ns) => (
            <option key={ns} value={ns}>
              {ns}
            </option>
          ))}
        </Select>
      </div>

      <div className="flex-1">
        <label className="text-sm font-medium mb-1 block">Risk Level</label>
        <Select
          value={filters.risk || ''}
          onChange={(e) => handleChange('risk', e.target.value as RiskLevel)}
        >
          <option value="">All Levels</option>
          <option value="critical">Critical</option>
          <option value="high">High</option>
          <option value="medium">Medium</option>
          <option value="low">Low</option>
        </Select>
      </div>

      <div className="flex-1">
        <label className="text-sm font-medium mb-1 block">Time Range</label>
        <Select
          value=""
          onChange={(e) => {
            const now = Date.now()
            const ranges: Record<string, { start: number; end: number }> = {
              '1h': { start: now - 3600000, end: now },
              '24h': { start: now - 86400000, end: now },
              '7d': { start: now - 604800000, end: now },
              '30d': { start: now - 2592000000, end: now },
            }
            const range = ranges[e.target.value]
            if (range) {
              onFiltersChange({
                ...filters,
                start: range.start,
                end: range.end,
                offset: 0,
              })
            } else {
              const { start, end, ...rest } = filters
              onFiltersChange(rest)
            }
          }}
        >
          <option value="">All Time</option>
          <option value="1h">Last Hour</option>
          <option value="24h">Last 24 Hours</option>
          <option value="7d">Last 7 Days</option>
          <option value="30d">Last 30 Days</option>
        </Select>
      </div>
    </div>
  )
}
