import { ShieldAlert, AlertTriangle, AlertCircle, Info } from 'lucide-react'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { type Trace, type RiskLevel } from '@/lib/types'
import { format } from 'date-fns'

type TraceDetailSheetProps = {
  trace: Trace | null
  open: boolean
  onClose: () => void
}

const riskConfig = {
  critical: { label: 'Critical', variant: 'destructive' as const, icon: ShieldAlert },
  high: { label: 'High', variant: 'default' as const, icon: AlertTriangle },
  medium: { label: 'Medium', variant: 'secondary' as const, icon: AlertCircle },
  low: { label: 'Low', variant: 'outline' as const, icon: Info },
}

// Flatten nested JSON object to dot notation
function flattenObject(obj: any, prefix = ''): Record<string, any> {
  const flattened: Record<string, any> = {}
  
  for (const [key, value] of Object.entries(obj)) {
    const newKey = prefix ? `${prefix}.${key}` : key
    
    if (value !== null && typeof value === 'object' && !Array.isArray(value)) {
      Object.assign(flattened, flattenObject(value, newKey))
    } else if (Array.isArray(value)) {
      flattened[newKey] = JSON.stringify(value)
    } else {
      flattened[newKey] = value
    }
  }
  
  return flattened
}

export function TraceDetailSheet({ trace, open, onClose }: TraceDetailSheetProps) {
  if (!trace) return null

  const config = riskConfig[trace.risk_level as RiskLevel]
  const Icon = config.icon

  // Helper to format output as JSON if possible
  const formatOutput = (output: string): { isJson: boolean; content: string; parsed?: any } => {
    try {
      const parsed = JSON.parse(output)
      return { isJson: true, content: JSON.stringify(parsed, null, 2), parsed }
    } catch {
      return { isJson: false, content: output }
    }
  }

  const formattedOutput = trace.output ? formatOutput(trace.output) : null
  const flattenedData = formattedOutput?.isJson && formattedOutput.parsed 
    ? flattenObject(formattedOutput.parsed) 
    : null

  return (
    <Sheet open={open} onOpenChange={onClose}>
      <SheetContent className="sm:max-w-3xl">
        <SheetHeader>
          <SheetTitle className="flex items-center gap-2">
            <Icon className="h-5 w-5" />
            Trace Details
          </SheetTitle>
          <SheetDescription>
            {format(new Date(trace.timestamp), 'yyyy-MM-dd HH:mm:ss')}
          </SheetDescription>
        </SheetHeader>

        <ScrollArea className="h-[calc(100vh-8rem)] mt-6 pr-6">
          <div className="space-y-6">
            {/* Risk & Status */}
            <div>
              <h3 className="text-sm font-medium mb-3">Risk & Status</h3>
              <div className="flex gap-2">
                <Badge variant={config.variant}>
                  {config.label} Risk
                </Badge>
                <Badge variant={trace.result === 'success' ? 'default' : 'destructive'}>
                  {trace.result}
                </Badge>
              </div>
              {trace.risk_reason && (
                <p className="mt-2 text-sm text-muted-foreground">{trace.risk_reason}</p>
              )}
            </div>

            <Separator />

            {/* Tool Information */}
            <div>
              <h3 className="text-sm font-medium mb-3">Tool Information</h3>
              <dl className="space-y-2 text-sm">
                <div className="grid grid-cols-[120px_1fr] gap-2">
                  <dt className="text-muted-foreground">Tool</dt>
                  <dd className="font-medium">{trace.tool_name}</dd>
                </div>
                <div className="grid grid-cols-[120px_1fr] gap-2">
                  <dt className="text-muted-foreground">Namespace</dt>
                  <dd><Badge variant="outline">{trace.namespace}</Badge></dd>
                </div>
                <div className="grid grid-cols-[120px_1fr] gap-2">
                  <dt className="text-muted-foreground">Resource Kind</dt>
                  <dd className="font-medium">{trace.resource_kind}</dd>
                </div>
                <div className="grid grid-cols-[120px_1fr] gap-2">
                  <dt className="text-muted-foreground">Target Resource</dt>
                  <dd className="font-mono text-xs break-all">{trace.target_resource}</dd>
                </div>
              </dl>
            </div>

            <Separator />

            {/* Command */}
            <div>
              <h3 className="text-sm font-medium mb-3">Command</h3>
              <div className="rounded-md bg-muted p-3">
                <code className="text-xs font-mono break-all">{trace.command}</code>
              </div>
            </div>

            {/* Output */}
            {formattedOutput && (
              <>
                <Separator />
                <div>
                  <h3 className="text-sm font-medium mb-3">Output</h3>
                  {formattedOutput.isJson ? (
                    <Tabs defaultValue="json" className="w-full">
                      <TabsList>
                        <TabsTrigger value="json">JSON</TabsTrigger>
                        <TabsTrigger value="table">Table</TabsTrigger>
                      </TabsList>
                      <TabsContent value="json" className="mt-2">
                        <div className="rounded-md bg-muted p-3 max-h-[400px] overflow-auto">
                          <pre className="text-xs font-mono whitespace-pre-wrap break-all">{formattedOutput.content}</pre>
                        </div>
                      </TabsContent>
                      <TabsContent value="table" className="mt-2">
                        <div className="rounded-md border max-h-[400px] overflow-auto">
                          <Table>
                            <TableHeader>
                              <TableRow>
                                <TableHead className="w-[180px] min-w-[120px]">Key</TableHead>
                                <TableHead>Value</TableHead>
                              </TableRow>
                            </TableHeader>
                            <TableBody>
                              {flattenedData && Object.entries(flattenedData).map(([key, value]) => (
                                <TableRow key={key}>
                                  <TableCell className="font-mono text-xs">{key}</TableCell>
                                  <TableCell className="font-mono text-xs break-all">
                                    {String(value)}
                                  </TableCell>
                                </TableRow>
                              ))}
                            </TableBody>
                          </Table>
                        </div>
                      </TabsContent>
                    </Tabs>
                  ) : (
                    <div className="rounded-md bg-muted p-3 max-h-[400px] overflow-auto">
                      <pre className="text-xs font-mono whitespace-pre-wrap break-all">{formattedOutput.content}</pre>
                    </div>
                  )}
                </div>
              </>
            )}

            {/* Error Message */}
            {trace.error_message && (
              <>
                <Separator />
                <div>
                  <h3 className="text-sm font-medium mb-3 text-destructive">Error</h3>
                  <div className="rounded-md bg-destructive/10 p-3">
                    <pre className="text-xs font-mono whitespace-pre-wrap text-destructive">
                      {trace.error_message}
                    </pre>
                  </div>
                </div>
              </>
            )}

            <Separator />

            {/* Metrics */}
            <div>
              <h3 className="text-sm font-medium mb-3">Metrics</h3>
              <dl className="space-y-2 text-sm">
                <div className="grid grid-cols-[120px_1fr] gap-2">
                  <dt className="text-muted-foreground">Latency</dt>
                  <dd className="font-medium">
                    {trace.latency_ms && trace.latency_ms > 0 ? `${trace.latency_ms}ms` : 'N/A'}
                  </dd>
                </div>
                <div className="grid grid-cols-[120px_1fr] gap-2">
                  <dt className="text-muted-foreground">Session ID</dt>
                  <dd className="font-mono text-xs break-all">
                    {trace.session_id || 'N/A'}
                  </dd>
                </div>
                {trace.tokens_input !== undefined && (
                  <div className="grid grid-cols-[120px_1fr] gap-2">
                    <dt className="text-muted-foreground">Tokens (In/Out)</dt>
                    <dd className="font-medium">{trace.tokens_input} / {trace.tokens_output}</dd>
                  </div>
                )}
                {trace.cost_estimate !== undefined && (
                  <div className="grid grid-cols-[120px_1fr] gap-2">
                    <dt className="text-muted-foreground">Cost Estimate</dt>
                    <dd className="font-medium">${trace.cost_estimate.toFixed(6)}</dd>
                  </div>
                )}
              </dl>
            </div>
          </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  )
}
