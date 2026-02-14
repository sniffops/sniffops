package trace

import (
	"database/sql"
	"fmt"
	"time"
)

// Stats represents aggregated statistics from traces
type Stats struct {
	RiskDistribution map[string]int       `json:"risk_distribution"`
	ToolUsage        map[string]int       `json:"tool_usage"`
	Timeline         []TimelinePoint      `json:"timeline"`
	TotalOperations  int                  `json:"total_operations"`
	TotalCost        float64              `json:"total_cost_estimate"`
}

// TimelinePoint represents a point in the timeline chart
type TimelinePoint struct {
	Hour  string `json:"hour"`
	Count int    `json:"count"`
}

// GetStats retrieves aggregated statistics for the specified period
// period can be: "1h", "24h", "7d", "30d", or empty for all-time
func GetStats(db *sql.DB, period string) (*Stats, error) {
	stats := &Stats{
		RiskDistribution: make(map[string]int),
		ToolUsage:        make(map[string]int),
		Timeline:         []TimelinePoint{},
	}

	// Calculate time range based on period
	var startTime *int64
	if period != "" {
		start := calculateStartTime(period)
		if start != nil {
			ts := start.UnixMilli()
			startTime = &ts
		}
	}

	// Build WHERE clause for time filtering
	whereClause := ""
	var args []interface{}
	if startTime != nil {
		whereClause = "WHERE timestamp >= ?"
		args = append(args, *startTime)
	}

	// 1. Risk distribution
	riskQuery := fmt.Sprintf("SELECT risk_level, COUNT(*) FROM traces %s GROUP BY risk_level", whereClause)
	rows, err := db.Query(riskQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query risk distribution: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var level string
		var count int
		if err := rows.Scan(&level, &count); err != nil {
			return nil, err
		}
		stats.RiskDistribution[level] = count
		stats.TotalOperations += count
	}
	rows.Close()

	// 2. Tool usage
	toolQuery := fmt.Sprintf("SELECT tool_name, COUNT(*) FROM traces %s GROUP BY tool_name ORDER BY COUNT(*) DESC", whereClause)
	rows, err = db.Query(toolQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tool usage: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tool string
		var count int
		if err := rows.Scan(&tool, &count); err != nil {
			return nil, err
		}
		stats.ToolUsage[tool] = count
	}
	rows.Close()

	// 3. Timeline (hourly buckets)
	// SQLite doesn't have built-in date_trunc, so we'll use strftime
	timelineQuery := fmt.Sprintf(`
		SELECT 
			strftime('%%Y-%%m-%%dT%%H:00:00Z', datetime(timestamp/1000, 'unixepoch')) as hour,
			COUNT(*) as count
		FROM traces
		%s
		GROUP BY hour
		ORDER BY hour DESC
		LIMIT 24
	`, whereClause)

	rows, err = db.Query(timelineQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query timeline: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var point TimelinePoint
		if err := rows.Scan(&point.Hour, &point.Count); err != nil {
			return nil, err
		}
		stats.Timeline = append(stats.Timeline, point)
	}
	rows.Close()

	// 4. Total cost
	costQuery := fmt.Sprintf("SELECT COALESCE(SUM(cost_estimate), 0) FROM traces %s", whereClause)
	err = db.QueryRow(costQuery, args...).Scan(&stats.TotalCost)
	if err != nil {
		return nil, fmt.Errorf("failed to query total cost: %w", err)
	}

	return stats, nil
}

// GetDistinctValues retrieves distinct values for a column (for filter autocomplete)
func GetDistinctValues(db *sql.DB, column string) ([]string, error) {
	// Whitelist allowed columns to prevent SQL injection
	allowedColumns := map[string]bool{
		"namespace": true,
		"tool_name": true,
	}

	if !allowedColumns[column] {
		return nil, fmt.Errorf("invalid column name: %s", column)
	}

	query := fmt.Sprintf("SELECT DISTINCT %s FROM traces WHERE %s IS NOT NULL AND %s != '' ORDER BY %s", 
		column, column, column, column)

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query distinct values: %w", err)
	}
	defer rows.Close()

	var values []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return values, nil
}

// calculateStartTime converts period string to start time
func calculateStartTime(period string) *time.Time {
	now := time.Now()
	var duration time.Duration

	switch period {
	case "1h":
		duration = time.Hour
	case "24h":
		duration = 24 * time.Hour
	case "7d":
		duration = 7 * 24 * time.Hour
	case "30d":
		duration = 30 * 24 * time.Hour
	default:
		return nil // All-time
	}

	start := now.Add(-duration)
	return &start
}
