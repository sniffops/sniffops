import { Card, CardContent, CardHeader, CardTitle } from './ui/card'
import type { Stats as StatsType } from '@/lib/types'
import { Activity, DollarSign, BarChart3 } from 'lucide-react'

interface StatsProps {
  stats: StatsType
}

export function Stats({ stats }: StatsProps) {
  const topTools = Object.entries(stats.tool_usage)
    .sort(([, a], [, b]) => b - a)
    .slice(0, 5)

  const maxUsage = Math.max(...Object.values(stats.tool_usage))

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
      {/* Total Operations */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">
            Total Operations
          </CardTitle>
          <Activity className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{stats.total_operations}</div>
          <p className="text-xs text-muted-foreground mt-1">
            across all tools
          </p>
        </CardContent>
      </Card>

      {/* Total Cost */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">
            Total Cost
          </CardTitle>
          <DollarSign className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">
            ${stats.total_cost_estimate.toFixed(4)}
          </div>
          <p className="text-xs text-muted-foreground mt-1">
            estimated LLM cost
          </p>
        </CardContent>
      </Card>

      {/* Tool Usage */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">
            Most Used Tools
          </CardTitle>
          <BarChart3 className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {topTools.map(([tool, count]) => (
              <div key={tool} className="flex items-center gap-2">
                <div className="text-xs font-mono flex-shrink-0 w-24">
                  {tool.replace('sniff_', '')}
                </div>
                <div className="flex-1 bg-muted rounded-full h-2 overflow-hidden">
                  <div
                    className="bg-primary h-full transition-all"
                    style={{ width: `${(count / maxUsage) * 100}%` }}
                  />
                </div>
                <div className="text-xs font-medium w-8 text-right">
                  {count}
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
