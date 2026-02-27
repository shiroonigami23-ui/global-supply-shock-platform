package risk

import (
	"math"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/shiroonigami23-ui/global-supply-shock-platform/internal/contracts"
)

type Engine struct {
	mu      sync.Mutex
	window  time.Duration
	history map[string][]contracts.SignalEvent
}

func NewEngine(window time.Duration) *Engine {
	return &Engine{
		window:  window,
		history: make(map[string][]contracts.SignalEvent),
	}
}

func (e *Engine) Process(signal contracts.SignalEvent) contracts.RiskEvent {
	now := time.Now().UTC()
	key := signal.Key()

	e.mu.Lock()
	defer e.mu.Unlock()

	entries := append(e.history[key], signal)
	cutoff := now.Add(-e.window)

	trimmed := entries[:0]
	for _, s := range entries {
		if s.Timestamp.IsZero() {
			s.Timestamp = now
		}
		if s.Timestamp.After(cutoff) {
			trimmed = append(trimmed, s)
		}
	}

	if len(trimmed) > 150 {
		trimmed = trimmed[len(trimmed)-150:]
	}

	e.history[key] = trimmed

	score, contributors := aggregate(trimmed)

	return contracts.RiskEvent{
		ID:                uuid.NewString(),
		Timestamp:         now,
		Country:           signal.Country,
		Region:            signal.Region,
		Commodity:         signal.Commodity,
		RiskScore:         score,
		WindowMinutes:     int(e.window.Minutes()),
		Contributors:      contributors,
		RecommendedAction: recommendation(score),
	}
}

func aggregate(signals []contracts.SignalEvent) (float64, []contracts.RiskContributor) {
	if len(signals) == 0 {
		return 0, nil
	}

	scored := make([]contracts.RiskContributor, 0, len(signals))
	total := 0.0
	for _, s := range signals {
		c := contracts.RiskContributor{
			Source:      s.Source,
			MetricName:  s.MetricName,
			MetricValue: s.MetricValue,
			Score:       signalScore(s),
		}
		total += c.Score
		scored = append(scored, c)
	}

	avg := total / float64(len(scored))
	score := clamp(avg, 0, 100)

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})
	if len(scored) > 5 {
		scored = scored[:5]
	}

	return round2(score), scored
}

func signalScore(s contracts.SignalEvent) float64 {
	severity := clamp(float64(s.Severity), 0, 10) * 5.0
	confidence := clamp(s.Confidence, 0, 1) * 20.0

	metricNorm := s.MetricValue
	if metricNorm > 100 {
		metricNorm = 100
	}
	if metricNorm < 0 {
		metricNorm = 0
	}

	valueComponent := (metricNorm / 100.0) * 40.0

	weighted := (severity + confidence + valueComponent) * sourceWeight(s.Source)
	return clamp(weighted, 0, 100)
}

func sourceWeight(source contracts.SignalSource) float64 {
	switch source {
	case contracts.SourceShippingLane:
		return 1.25
	case contracts.SourcePortCongestion:
		return 1.30
	case contracts.SourceWeather:
		return 1.20
	case contracts.SourcePriceSpike:
		return 1.10
	case contracts.SourceNews:
		return 1.00
	default:
		return 1.00
	}
}

func recommendation(score float64) string {
	switch {
	case score >= 85:
		return "Immediate intervention: pre-position inventory and activate cross-border backup routes."
	case score >= 70:
		return "High risk: increase safety stock and notify regional distributors within 2 hours."
	case score >= 50:
		return "Moderate risk: monitor hourly and prepare route alternatives."
	default:
		return "Low risk: continue monitoring with standard cadence."
	}
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
