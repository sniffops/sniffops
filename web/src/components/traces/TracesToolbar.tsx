import { useSearchParams } from 'react-router-dom'
import { X, Search } from 'lucide-react'
import { type Table } from '@tanstack/react-table'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { type Trace } from '@/lib/types'

type TracesToolbarProps = {
  table: Table<Trace>
  namespaces: string[]
  tools: string[]
  loading: boolean
}

export function TracesToolbar({ namespaces, tools, loading }: TracesToolbarProps) {
  const [searchParams, setSearchParams] = useSearchParams()

  const updateParam = (key: string, value: string | null) => {
    const newParams = new URLSearchParams(searchParams)
    if (value) {
      newParams.set(key, value)
    } else {
      newParams.delete(key)
    }
    newParams.set('offset', '0') // Reset to first page
    setSearchParams(newParams)
  }

  const clearFilters = () => {
    setSearchParams({})
  }

  const hasFilters = searchParams.has('tool') || searchParams.has('namespace') || searchParams.has('risk') || searchParams.has('search')

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-2">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search command or resource..."
            value={searchParams.get('search') || ''}
            onChange={(e) => updateParam('search', e.target.value || null)}
            className="pl-8"
            disabled={loading}
          />
        </div>

        <Select
          value={searchParams.get('tool') || 'all'}
          onValueChange={(value) => updateParam('tool', value === 'all' ? null : value)}
          disabled={loading}
        >
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="Filter by tool" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Tools</SelectItem>
            {tools.map((tool) => (
              <SelectItem key={tool} value={tool}>
                {tool}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select
          value={searchParams.get('namespace') || 'all'}
          onValueChange={(value) => updateParam('namespace', value === 'all' ? null : value)}
          disabled={loading}
        >
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="Filter by namespace" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Namespaces</SelectItem>
            {namespaces.map((ns) => (
              <SelectItem key={ns} value={ns}>
                {ns}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select
          value={searchParams.get('risk') || 'all'}
          onValueChange={(value) => updateParam('risk', value === 'all' ? null : value)}
          disabled={loading}
        >
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="Filter by risk" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Risk Levels</SelectItem>
            <SelectItem value="critical">Critical</SelectItem>
            <SelectItem value="high">High</SelectItem>
            <SelectItem value="medium">Medium</SelectItem>
            <SelectItem value="low">Low</SelectItem>
          </SelectContent>
        </Select>

        {hasFilters && (
          <Button
            variant="ghost"
            onClick={clearFilters}
            className="h-10 px-4"
            disabled={loading}
          >
            <X className="mr-2 h-4 w-4" />
            Clear
          </Button>
        )}
      </div>
    </div>
  )
}
