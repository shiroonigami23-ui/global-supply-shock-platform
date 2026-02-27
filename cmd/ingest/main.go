package main

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"github.com/shiroonigami23-ui/global-supply-shock-platform/internal/config"
	"github.com/shiroonigami23-ui/global-supply-shock-platform/internal/contracts"
	"github.com/shiroonigami23-ui/global-supply-shock-platform/internal/httpx"
	"github.com/shiroonigami23-ui/global-supply-shock-platform/internal/mq"
)

func main() {
	cfg := config.Load()

	writer := mq.NewWriter(cfg.KafkaBrokers, cfg.KafkaTopicSignals)
	defer writer.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.SimulatorTick > 0 {
		go runSimulator(ctx, writer, cfg.SimulatorTick)
	}

	router := chi.NewRouter()
	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "service": "ingest"})
	})

	router.Post("/v1/signals", func(w http.ResponseWriter, r *http.Request) {
		var payload contracts.SignalEvent
		if err := httpx.DecodeJSON(r, &payload); err != nil {
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}

		if strings.TrimSpace(payload.Country) == "" || strings.TrimSpace(payload.Commodity) == "" {
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "country and commodity are required"})
			return
		}

		enrichSignal(&payload)
		if err := mq.PublishJSON(r.Context(), writer, payload.Key(), payload); err != nil {
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}

		httpx.WriteJSON(w, http.StatusAccepted, payload)
	})

	router.Post("/v1/simulate", func(w http.ResponseWriter, r *http.Request) {
		type req struct {
			Count int `json:"count"`
		}
		body := req{Count: 10}
		_ = httpx.DecodeJSON(r, &body)

		if body.Count <= 0 {
			body.Count = 10
		}
		if body.Count > 500 {
			body.Count = 500
		}

		sent := 0
		for range body.Count {
			signal := randomSignal()
			if err := mq.PublishJSON(r.Context(), writer, signal.Key(), signal); err != nil {
				log.Printf("simulate publish error: %v", err)
				break
			}
			sent++
		}

		httpx.WriteJSON(w, http.StatusAccepted, map[string]any{"requested": body.Count, "published": sent})
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("ingest listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("ingest server error: %v", err)
	}
}

func runSimulator(ctx context.Context, writer *kafka.Writer, tick time.Duration) {
	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			signal := randomSignal()
			if err := mq.PublishJSON(ctx, writer, signal.Key(), signal); err != nil {
				log.Printf("simulator publish error: %v", err)
			}
		}
	}
}

func enrichSignal(s *contracts.SignalEvent) {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	if s.Timestamp.IsZero() {
		s.Timestamp = time.Now().UTC()
	}
	s.Country = strings.ToUpper(strings.TrimSpace(s.Country))
	s.Region = strings.TrimSpace(s.Region)
	if s.Region == "" {
		s.Region = "global"
	}
	s.Commodity = strings.ToLower(strings.TrimSpace(s.Commodity))
	if s.Severity < 1 {
		s.Severity = 1
	}
	if s.Severity > 10 {
		s.Severity = 10
	}
	if s.Confidence <= 0 {
		s.Confidence = 0.6
	}
	if s.Confidence > 1 {
		s.Confidence = 1
	}
	if s.Source == "" {
		s.Source = contracts.SourceNews
	}
	if s.MetricName == "" {
		s.MetricName = "composite_signal"
	}
}

func randomSignal() contracts.SignalEvent {
	countries := []string{"US", "DE", "IN", "BR", "ZA", "ID", "JP", "NG"}
	regions := []string{"north", "south", "west", "east", "metro", "coastal"}
	commodities := []string{"insulin", "diesel", "wheat", "rice", "antibiotics"}
	sources := []contracts.SignalSource{
		contracts.SourceShippingLane,
		contracts.SourcePortCongestion,
		contracts.SourceWeather,
		contracts.SourcePriceSpike,
		contracts.SourceNews,
	}

	return contracts.SignalEvent{
		ID:          uuid.NewString(),
		Timestamp:   time.Now().UTC(),
		Source:      sources[rand.Intn(len(sources))],
		Country:     countries[rand.Intn(len(countries))],
		Region:      regions[rand.Intn(len(regions))],
		Commodity:   commodities[rand.Intn(len(commodities))],
		MetricName:  "anomaly_index",
		MetricValue: float64(15 + rand.Intn(85)),
		Severity:    2 + rand.Intn(9),
		Confidence:  0.4 + rand.Float64()*0.6,
	}
}
