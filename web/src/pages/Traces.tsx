import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { TracesTable } from '@/components/traces/TracesTable'
import { fetchTraces, fetchNamespaces, fetchTools } from '@/lib/api'
import { type Trace } from '@/lib/types'

export function Traces() {
  const [searchParams] = useSearchParams()
  const [traces, setTraces] = useState<Trace[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [namespaces, setNamespaces] = useState<string[]>([])
  const [tools, setTools] = useState<string[]>([])

  useEffect(() => {
    const loadMeta = async () => {
      try {
        const [ns, t] = await Promise.all([
          fetchNamespaces(),
          fetchTools(),
        ])
        setNamespaces(ns)
        setTools(t)
      } catch (error) {
        console.error('Failed to load metadata:', error)
      }
    }
    loadMeta()
  }, [])

  useEffect(() => {
    const loadTraces = async () => {
      setLoading(true)
      try {
        const filters = {
          tool: searchParams.get('tool') || undefined,
          namespace: searchParams.get('namespace') || undefined,
          risk: searchParams.get('risk') as any || undefined,
          limit: parseInt(searchParams.get('limit') || '50'),
          offset: parseInt(searchParams.get('offset') || '0'),
        }
        
        const response = await fetchTraces(filters)
        setTraces(response.traces || [])
        setTotal(response.total || 0)
      } catch (error) {
        console.error('Failed to load traces:', error)
      } finally {
        setLoading(false)
      }
    }
    loadTraces()
  }, [searchParams])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Traces</h1>
        <p className="text-muted-foreground">
          View and analyze all security traces
        </p>
      </div>

      <TracesTable
        data={traces}
        total={total}
        loading={loading}
        namespaces={namespaces}
        tools={tools}
      />
    </div>
  )
}
