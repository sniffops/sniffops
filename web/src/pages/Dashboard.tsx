import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { AlertTriangle, ShieldAlert, AlertCircle, Info, Activity, Wrench } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { fetchStats, fetchTraces } from '@/lib/api'
import { type Stats, type Trace, type RiskLevel } from '@/lib/types'
import { formatDistanceToNow } from 'date-fns'

const riskConfig = {
  critical: { label: 'Critical', icon: ShieldAlert, color: 'text-red-500', bg: 'bg-red-500/10' },
  high: { label: 'High', icon: AlertTriangle, color: 'text-orange-500', bg: 'bg-orange-500/10' },
  medium: { label: 'Medium', icon: AlertCircle, color: 'text-yellow-500', bg: 'bg-yellow-500/10' },
  low: { label: 'Low', icon: Info, color: 'text-blue-500', bg: 'bg-blue-500/10' },
}

export function Dashboard() {
  const navigate = useNavigate()
  const [stats, setStats] = useState<Stats | null>(null)
  const [recentTraces, setRecentTraces] = useState<Trace[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const loadData = async () => {
      try {
        const [statsData, tracesData] = await Promise.all([
          fetchStats(),
          fetchTraces({ limit: 5 })
        ])
        setStats(statsData)
        setRecentTraces(tracesData.traces || [])
      } catch (error) {
        console.error('Failed to load dashboard data:', error)
      } finally {
        setLoading(false)
      }
    }
    loadData()
  }, [])

  if (loading) {
    return <div className="flex items-center justify-center h-full">Loading...</div>
  }

  const distribution = stats?.risk_distribution || { critical: 0, high: 0, medium: 0, low: 0 }
  const toolUsage = stats?.tool_usage || {}
  const topTools = Object.entries(toolUsage)
    .sort(([, a], [, b]) => b - a)
    .slice(0, 5)

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground">
          MCP Tool Security Monitoring Overview
        </p>
      </div>

      {/* Risk Distribution Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {(Object.entries(distribution) as [RiskLevel, number][]).map(([level, count]) => {
          const config = riskConfig[level]
          const Icon = config.icon
          return (
            <Card
              key={level}
              className="cursor-pointer transition-colors hover:bg-accent"
              onClick={() => navigate(`/traces?risk=${level}`)}
            >
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium capitalize">
                  {config.label} Risk
                </CardTitle>
                <Icon className={`h-4 w-4 ${config.color}`} />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{count}</div>
                <p className="text-xs text-muted-foreground">
                  operations detected
                </p>
              </CardContent>
            </Card>
          )
        })}
      </div>

      {/* Statistics */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              Total Operations
            </CardTitle>
            <CardDescription>All tracked operations</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{stats?.total_operations || 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Wrench className="h-5 w-5" />
              Most Used Tools
            </CardTitle>
            <CardDescription>Top 5 tools by usage</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            {topTools.map(([tool, count]) => (
              <div key={tool} className="flex items-center justify-between">
                <span className="text-sm font-medium">{tool}</span>
                <Badge variant="secondary">{count}</Badge>
              </div>
            ))}
            {topTools.length === 0 && (
              <p className="text-sm text-muted-foreground">No data available</p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Recent Traces */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Traces</CardTitle>
          <CardDescription>Latest 5 operations</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {recentTraces.map((trace) => {
              const config = riskConfig[trace.risk_level]
              const Icon = config.icon
              return (
                <div
                  key={trace.id}
                  className="flex items-start gap-4 rounded-lg border p-4 transition-colors hover:bg-accent cursor-pointer"
                  onClick={() => navigate('/traces')}
                >
                  <div className={`rounded-lg p-2 ${config.bg}`}>
                    <Icon className={`h-4 w-4 ${config.color}`} />
                  </div>
                  <div className="flex-1 space-y-1">
                    <div className="flex items-center gap-2">
                      <p className="font-medium">{trace.tool_name}</p>
                      <Badge variant="outline" className="text-xs">
                        {trace.namespace}
                      </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground line-clamp-1">
                      {trace.command}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {formatDistanceToNow(new Date(trace.timestamp * 1000), { addSuffix: true })}
                    </p>
                  </div>
                  <Badge variant={trace.result === 'success' ? 'default' : 'destructive'}>
                    {trace.result}
                  </Badge>
                </div>
              )
            })}
            {recentTraces.length === 0 && (
              <p className="text-sm text-center text-muted-foreground py-8">
                No recent traces
              </p>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
