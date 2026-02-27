CREATE TABLE IF NOT EXISTS risk_events (
  id UUID PRIMARY KEY,
  event_ts TIMESTAMPTZ NOT NULL,
  country TEXT NOT NULL,
  region TEXT NOT NULL,
  commodity TEXT NOT NULL,
  risk_score DOUBLE PRECISION NOT NULL,
  window_minutes INT NOT NULL,
  contributors JSONB NOT NULL,
  recommended_action TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_risk_events_lookup
  ON risk_events(country, commodity, created_at DESC);

CREATE TABLE IF NOT EXISTS alerts (
  id UUID PRIMARY KEY,
  risk_event_id UUID,
  country TEXT NOT NULL,
  region TEXT NOT NULL,
  commodity TEXT NOT NULL,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  risk_score DOUBLE PRECISION NOT NULL,
  severity TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'open',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  acknowledged_at TIMESTAMPTZ,
  resolved_at TIMESTAMPTZ,
  CONSTRAINT alerts_status_check CHECK (status IN ('open', 'acknowledged', 'resolved'))
);

CREATE INDEX IF NOT EXISTS idx_alerts_status_created
  ON alerts(status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_alerts_geo
  ON alerts(country, commodity, created_at DESC);
