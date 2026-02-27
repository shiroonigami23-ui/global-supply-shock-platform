package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/shiroonigami23-ui/global-supply-shock-platform/internal/config"
	"github.com/shiroonigami23-ui/global-supply-shock-platform/internal/contracts"
	"github.com/shiroonigami23-ui/global-supply-shock-platform/internal/mq"
	"github.com/shiroonigami23-ui/global-supply-shock-platform/internal/risk"
	"github.com/shiroonigami23-ui/global-supply-shock-platform/internal/storage"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbPool, err := storage.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("risk-engine database error: %v", err)
	}
	defer dbPool.Close()

	repo := storage.NewRepository(dbPool)

	reader := mq.NewReader(cfg.KafkaBrokers, cfg.KafkaTopicSignals, cfg.ConsumerGroupPrefix+"-risk-engine")
	defer reader.Close()

	writer := mq.NewWriter(cfg.KafkaBrokers, cfg.KafkaTopicRisk)
	defer writer.Close()

	engine := risk.NewEngine(30 * time.Minute)

	log.Printf("risk-engine consuming %s and producing %s", cfg.KafkaTopicSignals, cfg.KafkaTopicRisk)
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Println("risk-engine shutting down")
				return
			}
			log.Printf("risk-engine read error: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		signal, err := mq.ParseMessageJSON[contracts.SignalEvent](msg)
		if err != nil {
			log.Printf("risk-engine decode signal error: %v", err)
			continue
		}
		if signal.Timestamp.IsZero() {
			signal.Timestamp = time.Now().UTC()
		}

		scored := engine.Process(signal)

		if err := repo.InsertRiskEvent(ctx, scored); err != nil {
			log.Printf("risk-engine store risk event error: %v", err)
		}

		if err := mq.PublishJSON(ctx, writer, scored.Country+"|"+scored.Commodity, scored); err != nil {
			var temporary kafka.Error
			if errors.As(err, &temporary) {
				log.Printf("risk-engine kafka temporary error: %v", temporary)
			} else {
				log.Printf("risk-engine publish error: %v", err)
			}
			continue
		}

		log.Printf("risk-event %s %s/%s score=%.2f", scored.ID, scored.Country, scored.Commodity, scored.RiskScore)
	}
}
