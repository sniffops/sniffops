import { useState, useEffect } from 'react'
import { RiskDashboard } from './components/RiskDashboard'
import { FilterBar } from './components/FilterBar'
import { TraceTimeline } from './components/TraceTimeline'
import { TraceDetail } from './components/TraceDetail'
import { Stats } from './components/Stats'
import type { Trace, TraceFilters, RiskLevel } from './lib/types'
import {
  mockStats,
  mockNamespaces,
  mockTools,
  getMockTracesResponse,
} from './lib/mock-data'
import { Search } from 'lucide-react'

// Using mock data for now - will switch to API later
const USE_MOCK = true

function App() {
  const [filters, setFilters] = useState<TraceFilters>({ limit: 50, offset: 0 })
  const [traces, setTraces] = useState<Trace[]>([])
  const [selectedTrace, setSelectedTrace] = useState<Trace | null>(null)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)

  // Load data
  useEffect(() => {
    const loadData = async () => {
      setLoading(true)
      try {
        if (USE_MOCK) {
          // Use mock data
          const response = getMockTracesResponse(filters)
          setTraces(response.traces)
          setTotal(response.total)
        } else {
          // TODO: Use real API
          // const response = await fetchTraces(filters)
          // setTraces(response.traces)
          // setTotal(response.total)
        }
      } catch (error) {
        console.error('Failed to load traces:', error)
      } finally {
        setLoading(false)
      }
    }

    loadData()
  }, [filters])

  const handleRiskClick = (risk: RiskLevel) => {
    setFilters({ ...filters, risk, offset: 0 })
  }

  const handleLoadMore = () => {
    setFilters({
      ...filters,
      offset: (filters.offset || 0) + (filters.limit || 50),
    })
  }

  const hasMore = total > (filters.offset || 0) + traces.length

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-card">
        <div className="container mx-auto px-4 py-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Search className="h-8 w-8 text-primary" />
              <div>
                <h1 className="text-2xl font-bold">SniffOps Dashboard</h1>
                <p className="text-sm text-muted-foreground">
                  MCP Tool Security Monitoring
                </p>
              </div>
            </div>
            <div className="text-sm text-muted-foreground">
              {mockStats.total_operations} operations tracked
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-6 space-y-6">
        {/* Risk Distribution */}
        <section>
          <h2 className="text-lg font-semibold mb-4">Risk Distribution</h2>
          <RiskDashboard
            distribution={mockStats.risk_distribution}
            onRiskClick={handleRiskClick}
          />
        </section>

        {/* Stats */}
        <section>
          <h2 className="text-lg font-semibold mb-4">Statistics</h2>
          <Stats stats={mockStats} />
        </section>

        {/* Filters */}
        <section>
          <h2 className="text-lg font-semibold mb-4">Filters</h2>
          <FilterBar
            filters={filters}
            onFiltersChange={setFilters}
            tools={mockTools}
            namespaces={mockNamespaces}
          />
        </section>

        {/* Trace Timeline */}
        <section>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold">
              Trace Timeline
              {loading && (
                <span className="ml-2 text-sm text-muted-foreground">
                  Loading...
                </span>
              )}
            </h2>
            <div className="text-sm text-muted-foreground">
              Showing {traces.length} of {total} traces
            </div>
          </div>
          <TraceTimeline
            traces={traces}
            onTraceClick={setSelectedTrace}
            onLoadMore={handleLoadMore}
            hasMore={hasMore}
          />
        </section>
      </main>

      {/* Trace Detail Modal */}
      <TraceDetail
        trace={selectedTrace}
        open={selectedTrace !== null}
        onClose={() => setSelectedTrace(null)}
      />
    </div>
  )
}

export default App
