package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/shiroonigami23-ui/global-supply-shock-platform/internal/config"
	"github.com/shiroonigami23-ui/global-supply-shock-platform/internal/contracts"
	"github.com/shiroonigami23-ui/global-supply-shock-platform/internal/mq"
	"github.com/shiroonigami23-ui/global-supply-shock-platform/internal/storage"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbPool, err := storage.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("alert-service database error: %v", err)
	}
	defer dbPool.Close()

	repo := storage.NewRepository(dbPool)

	reader := mq.NewReader(cfg.KafkaBrokers, cfg.KafkaTopicRisk, cfg.ConsumerGroupPrefix+"-alert-service")
	defer reader.Close()

	log.Printf("alert-service consuming %s threshold=%.2f", cfg.KafkaTopicRisk, cfg.AlertThreshold)

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Println("alert-service shutting down")
				return
			}
			log.Printf("alert-service read error: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		riskEvent, err := mq.ParseMessageJSON[contracts.RiskEvent](msg)
		if err != nil {
			log.Printf("alert-service decode risk event error: %v", err)
			continue
		}

		if riskEvent.RiskScore < cfg.AlertThreshold {
			continue
		}

		exists, err := repo.HasOpenAlertInCooldown(ctx, riskEvent.Country, riskEvent.Region, riskEvent.Commodity, cfg.AlertCooldown)
		if err != nil {
			log.Printf("alert-service cooldown check error: %v", err)
			continue
		}
		if exists {
			continue
		}

		alert := contracts.AlertRecord{
			ID:          uuid.NewString(),
			RiskEventID: riskEvent.ID,
			Country:     riskEvent.Country,
			Region:      riskEvent.Region,
			Commodity:   riskEvent.Commodity,
			Title:       fmt.Sprintf("High disruption risk for %s", riskEvent.Commodity),
			Description: fmt.Sprintf("%s/%s scored %.2f. %s", riskEvent.Country, riskEvent.Region, riskEvent.RiskScore, riskEvent.RecommendedAction),
			RiskScore:   riskEvent.RiskScore,
			Severity:    severityFromScore(riskEvent.RiskScore),
			Status:      "open",
		}

		if err := repo.InsertAlert(ctx, alert); err != nil {
			log.Printf("alert-service insert alert error: %v", err)
			continue
		}

		log.Printf("alert created id=%s country=%s commodity=%s score=%.2f", alert.ID, alert.Country, alert.Commodity, alert.RiskScore)
	}
}

func severityFromScore(score float64) string {
	switch {
	case score >= 90:
		return "critical"
	case score >= 75:
		return "high"
	case score >= 60:
		return "medium"
	default:
		return "low"
	}
}
