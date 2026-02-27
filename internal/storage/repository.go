package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shiroonigami23-ui/global-supply-shock-platform/internal/contracts"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) InsertRiskEvent(ctx context.Context, event contracts.RiskEvent) error {
	contributors, err := json.Marshal(event.Contributors)
	if err != nil {
		return fmt.Errorf("marshal contributors: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
        INSERT INTO risk_events
            (id, event_ts, country, region, commodity, risk_score, window_minutes, contributors, recommended_action)
        VALUES
            ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9)
        ON CONFLICT (id) DO NOTHING
    `, event.ID, event.Timestamp, event.Country, event.Region, event.Commodity, event.RiskScore, event.WindowMinutes, string(contributors), event.RecommendedAction)
	if err != nil {
		return fmt.Errorf("insert risk event: %w", err)
	}

	return nil
}

func (r *Repository) ListRiskEvents(ctx context.Context, country, commodity string, limit int) ([]contracts.RiskEvent, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	rows, err := r.pool.Query(ctx, `
        SELECT id, event_ts, country, region, commodity, risk_score, window_minutes, contributors, recommended_action
        FROM risk_events
        WHERE ($1 = '' OR country = $1)
          AND ($2 = '' OR commodity = $2)
        ORDER BY event_ts DESC
        LIMIT $3
    `, country, commodity, limit)
	if err != nil {
		return nil, fmt.Errorf("query risk events: %w", err)
	}
	defer rows.Close()

	results := make([]contracts.RiskEvent, 0, limit)
	for rows.Next() {
		var event contracts.RiskEvent
		var contributorsRaw []byte
		if err := rows.Scan(
			&event.ID,
			&event.Timestamp,
			&event.Country,
			&event.Region,
			&event.Commodity,
			&event.RiskScore,
			&event.WindowMinutes,
			&contributorsRaw,
			&event.RecommendedAction,
		); err != nil {
			return nil, fmt.Errorf("scan risk event: %w", err)
		}

		_ = json.Unmarshal(contributorsRaw, &event.Contributors)
		results = append(results, event)
	}

	return results, nil
}

func (r *Repository) HasOpenAlertInCooldown(ctx context.Context, country, region, commodity string, cooldown time.Duration) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT 1
            FROM alerts
            WHERE status IN ('open', 'acknowledged')
              AND country = $1
              AND region = $2
              AND commodity = $3
              AND created_at >= NOW() - $4::interval
        )
    `, country, region, commodity, fmt.Sprintf("%f seconds", cooldown.Seconds())).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check cooldown alert: %w", err)
	}
	return exists, nil
}

func (r *Repository) InsertAlert(ctx context.Context, alert contracts.AlertRecord) error {
	if alert.ID == "" {
		alert.ID = uuid.NewString()
	}

	_, err := r.pool.Exec(ctx, `
        INSERT INTO alerts
            (id, risk_event_id, country, region, commodity, title, description, risk_score, severity, status)
        VALUES
            ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `, alert.ID, nullableUUID(alert.RiskEventID), alert.Country, alert.Region, alert.Commodity, alert.Title, alert.Description, alert.RiskScore, alert.Severity, alert.Status)
	if err != nil {
		return fmt.Errorf("insert alert: %w", err)
	}

	return nil
}

func (r *Repository) ListAlerts(ctx context.Context, status string, limit int) ([]contracts.AlertRecord, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	rows, err := r.pool.Query(ctx, `
        SELECT id, COALESCE(risk_event_id::text,''), country, region, commodity, title, description, risk_score, severity, status, created_at, updated_at
        FROM alerts
        WHERE ($1 = '' OR status = $1)
        ORDER BY created_at DESC
        LIMIT $2
    `, status, limit)
	if err != nil {
		return nil, fmt.Errorf("query alerts: %w", err)
	}
	defer rows.Close()

	alerts := make([]contracts.AlertRecord, 0, limit)
	for rows.Next() {
		var alert contracts.AlertRecord
		if err := rows.Scan(
			&alert.ID,
			&alert.RiskEventID,
			&alert.Country,
			&alert.Region,
			&alert.Commodity,
			&alert.Title,
			&alert.Description,
			&alert.RiskScore,
			&alert.Severity,
			&alert.Status,
			&alert.CreatedAt,
			&alert.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (r *Repository) UpdateAlertStatus(ctx context.Context, id, status string) error {
	cmd, err := r.pool.Exec(ctx, `
        UPDATE alerts
        SET status = $2,
            updated_at = NOW(),
            acknowledged_at = CASE WHEN $2 = 'acknowledged' THEN NOW() ELSE acknowledged_at END,
            resolved_at = CASE WHEN $2 = 'resolved' THEN NOW() ELSE resolved_at END
        WHERE id = $1
    `, id, status)
	if err != nil {
		return fmt.Errorf("update alert status: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

type DashboardSummary struct {
	OpenAlerts      int     `json:"open_alerts"`
	Acknowledged    int     `json:"acknowledged_alerts"`
	Resolved24h     int     `json:"resolved_last_24h"`
	AvgRiskScore24h float64 `json:"avg_risk_score_24h"`
}

type DashboardSeriesPoint struct {
	BucketStart      time.Time `json:"bucket_start"`
	AvgRiskScore     float64   `json:"avg_risk_score"`
	RiskEvents       int       `json:"risk_events"`
	OpenAlertsOpened int       `json:"open_alerts_opened"`
}

type Hotspot struct {
	Country         string    `json:"country"`
	Region          string    `json:"region"`
	Commodity       string    `json:"commodity"`
	AvgRiskScore    float64   `json:"avg_risk_score"`
	LatestRiskScore float64   `json:"latest_risk_score"`
	ActiveAlerts    int       `json:"active_alerts"`
	LastEventAt     time.Time `json:"last_event_at"`
}

func (r *Repository) DashboardSummary(ctx context.Context) (DashboardSummary, error) {
	var summary DashboardSummary
	err := r.pool.QueryRow(ctx, `
        SELECT
            COUNT(*) FILTER (WHERE status = 'open') AS open_alerts,
            COUNT(*) FILTER (WHERE status = 'acknowledged') AS acknowledged_alerts,
            COUNT(*) FILTER (WHERE status = 'resolved' AND resolved_at >= NOW() - INTERVAL '24 hours') AS resolved_last_24h,
            COALESCE((SELECT AVG(risk_score) FROM risk_events WHERE event_ts >= NOW() - INTERVAL '24 hours'), 0)
        FROM alerts
    `).Scan(&summary.OpenAlerts, &summary.Acknowledged, &summary.Resolved24h, &summary.AvgRiskScore24h)
	if err != nil {
		return DashboardSummary{}, fmt.Errorf("dashboard summary: %w", err)
	}
	return summary, nil
}

func (r *Repository) DashboardTimeSeries(ctx context.Context, hours int) ([]DashboardSeriesPoint, error) {
	if hours <= 0 || hours > 168 {
		hours = 24
	}
	interval := fmt.Sprintf("%d hours", hours)

	rows, err := r.pool.Query(ctx, `
        WITH buckets AS (
            SELECT generate_series(
                date_trunc('hour', NOW() - $1::interval),
                date_trunc('hour', NOW()),
                interval '1 hour'
            ) AS bucket_start
        )
        SELECT
            b.bucket_start,
            COALESCE((
                SELECT AVG(re.risk_score)
                FROM risk_events re
                WHERE re.event_ts >= b.bucket_start
                  AND re.event_ts < b.bucket_start + interval '1 hour'
            ), 0) AS avg_risk_score,
            COALESCE((
                SELECT COUNT(*)
                FROM risk_events re
                WHERE re.event_ts >= b.bucket_start
                  AND re.event_ts < b.bucket_start + interval '1 hour'
            ), 0) AS risk_events,
            COALESCE((
                SELECT COUNT(*)
                FROM alerts a
                WHERE a.created_at >= b.bucket_start
                  AND a.created_at < b.bucket_start + interval '1 hour'
                  AND a.status = 'open'
            ), 0) AS open_alerts_opened
        FROM buckets b
        ORDER BY b.bucket_start ASC
    `, interval)
	if err != nil {
		return nil, fmt.Errorf("dashboard timeseries query: %w", err)
	}
	defer rows.Close()

	points := make([]DashboardSeriesPoint, 0, 48)
	for rows.Next() {
		var point DashboardSeriesPoint
		if err := rows.Scan(&point.BucketStart, &point.AvgRiskScore, &point.RiskEvents, &point.OpenAlertsOpened); err != nil {
			return nil, fmt.Errorf("dashboard timeseries scan: %w", err)
		}
		points = append(points, point)
	}

	return points, nil
}

func (r *Repository) Hotspots(ctx context.Context, hours, limit int) ([]Hotspot, error) {
	if hours <= 0 || hours > 168 {
		hours = 24
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	interval := fmt.Sprintf("%d hours", hours)
	rows, err := r.pool.Query(ctx, `
        SELECT
            re.country,
            re.region,
            re.commodity,
            ROUND(AVG(re.risk_score)::numeric, 2)::float8 AS avg_risk_score,
            MAX(re.risk_score) AS latest_risk_score,
            MAX(re.event_ts) AS last_event_at,
            COALESCE((
                SELECT COUNT(*)
                FROM alerts a
                WHERE a.country = re.country
                  AND a.region = re.region
                  AND a.commodity = re.commodity
                  AND a.status IN ('open', 'acknowledged')
            ), 0) AS active_alerts
        FROM risk_events re
        WHERE re.event_ts >= NOW() - $1::interval
        GROUP BY re.country, re.region, re.commodity
        ORDER BY avg_risk_score DESC, active_alerts DESC, last_event_at DESC
        LIMIT $2
    `, interval, limit)
	if err != nil {
		return nil, fmt.Errorf("hotspots query: %w", err)
	}
	defer rows.Close()

	hotspots := make([]Hotspot, 0, limit)
	for rows.Next() {
		var hotspot Hotspot
		if err := rows.Scan(
			&hotspot.Country,
			&hotspot.Region,
			&hotspot.Commodity,
			&hotspot.AvgRiskScore,
			&hotspot.LatestRiskScore,
			&hotspot.LastEventAt,
			&hotspot.ActiveAlerts,
		); err != nil {
			return nil, fmt.Errorf("hotspots scan: %w", err)
		}
		hotspots = append(hotspots, hotspot)
	}

	return hotspots, nil
}

func nullableUUID(v string) any {
	if v == "" {
		return nil
	}
	return v
}
