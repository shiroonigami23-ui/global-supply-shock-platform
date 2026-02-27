package contracts

import "time"

type SignalSource string

const (
	SourceShippingLane   SignalSource = "shipping_lane"
	SourcePortCongestion SignalSource = "port_congestion"
	SourceWeather        SignalSource = "weather"
	SourcePriceSpike     SignalSource = "price_spike"
	SourceNews           SignalSource = "news"
)

type SignalEvent struct {
	ID          string            `json:"id"`
	Timestamp   time.Time         `json:"timestamp"`
	Source      SignalSource      `json:"source"`
	Country     string            `json:"country"`
	Region      string            `json:"region"`
	Commodity   string            `json:"commodity"`
	MetricName  string            `json:"metric_name"`
	MetricValue float64           `json:"metric_value"`
	Severity    int               `json:"severity"`
	Confidence  float64           `json:"confidence"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type RiskContributor struct {
	Source      SignalSource `json:"source"`
	MetricName  string       `json:"metric_name"`
	MetricValue float64      `json:"metric_value"`
	Score       float64      `json:"score"`
}

type RiskEvent struct {
	ID                string            `json:"id"`
	Timestamp         time.Time         `json:"timestamp"`
	Country           string            `json:"country"`
	Region            string            `json:"region"`
	Commodity         string            `json:"commodity"`
	RiskScore         float64           `json:"risk_score"`
	WindowMinutes     int               `json:"window_minutes"`
	Contributors      []RiskContributor `json:"contributors"`
	RecommendedAction string            `json:"recommended_action"`
}

type AlertRecord struct {
	ID          string    `json:"id"`
	RiskEventID string    `json:"risk_event_id"`
	Country     string    `json:"country"`
	Region      string    `json:"region"`
	Commodity   string    `json:"commodity"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	RiskScore   float64   `json:"risk_score"`
	Severity    string    `json:"severity"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s SignalEvent) Key() string {
	return s.Country + "|" + s.Region + "|" + s.Commodity
}
