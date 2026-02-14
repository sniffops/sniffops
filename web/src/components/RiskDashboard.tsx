import { Card, CardContent, CardHeader, CardTitle } from './ui/card'
import type { RiskDistribution, RiskLevel } from '@/lib/types'
import { AlertCircle, AlertTriangle, Info, CheckCircle } from 'lucide-react'

interface RiskDashboardProps {
  distribution: RiskDistribution
  onRiskClick?: (risk: RiskLevel) => void
}

const riskConfig = {
  critical: {
    label: 'Critical',
    color: 'text-red-600 dark:text-red-500',
    bgColor: 'bg-red-600/10 dark:bg-red-500/10',
    borderColor: 'border-red-600/20 dark:border-red-500/20',
    icon: AlertCircle,
  },
  high: {
    label: 'High',
    color: 'text-orange-600 dark:text-orange-500',
    bgColor: 'bg-orange-600/10 dark:bg-orange-500/10',
    borderColor: 'border-orange-600/20 dark:border-orange-500/20',
    icon: AlertTriangle,
  },
  medium: {
    label: 'Medium',
    color: 'text-yellow-600 dark:text-yellow-500',
    bgColor: 'bg-yellow-600/10 dark:bg-yellow-500/10',
    borderColor: 'border-yellow-600/20 dark:border-yellow-500/20',
    icon: Info,
  },
  low: {
    label: 'Low',
    color: 'text-green-600 dark:text-green-500',
    bgColor: 'bg-green-600/10 dark:bg-green-500/10',
    borderColor: 'border-green-600/20 dark:border-green-500/20',
    icon: CheckCircle,
  },
}

export function RiskDashboard({ distribution, onRiskClick }: RiskDashboardProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      {(Object.keys(riskConfig) as RiskLevel[]).map((risk) => {
        const config = riskConfig[risk]
        const Icon = config.icon
        const count = distribution[risk]

        return (
          <Card
            key={risk}
            className={`cursor-pointer transition-all hover:shadow-lg ${config.borderColor} ${config.bgColor}`}
            onClick={() => onRiskClick?.(risk)}
          >
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                {config.label}
              </CardTitle>
              <Icon className={`h-4 w-4 ${config.color}`} />
            </CardHeader>
            <CardContent>
              <div className={`text-2xl font-bold ${config.color}`}>
                {count}
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                operations
              </p>
            </CardContent>
          </Card>
        )
      })}
    </div>
  )
}
