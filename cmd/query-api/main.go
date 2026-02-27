package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"

	"github.com/shiroonigami23-ui/global-supply-shock-platform/internal/config"
	"github.com/shiroonigami23-ui/global-supply-shock-platform/internal/httpx"
	"github.com/shiroonigami23-ui/global-supply-shock-platform/internal/storage"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbPool, err := storage.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("query-api database error: %v", err)
	}
	defer dbPool.Close()

	if err := storage.RunMigrations(ctx, dbPool); err != nil {
		log.Fatalf("query-api migration error: %v", err)
	}

	repo := storage.NewRepository(dbPool)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(15 * time.Second))
	router.Use(corsMiddleware)

	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "service": "query-api"})
	})

	router.Get("/v1/risks", func(w http.ResponseWriter, r *http.Request) {
		country := r.URL.Query().Get("country")
		commodity := r.URL.Query().Get("commodity")
		limit := parseLimit(r.URL.Query().Get("limit"), 100)

		events, err := repo.ListRiskEvents(r.Context(), country, commodity, limit)
		if err != nil {
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": events})
	})

	router.Get("/v1/alerts", func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		limit := parseLimit(r.URL.Query().Get("limit"), 100)

		alerts, err := repo.ListAlerts(r.Context(), status, limit)
		if err != nil {
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": alerts})
	})

	router.Patch("/v1/alerts/{id}/ack", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := repo.UpdateAlertStatus(r.Context(), id, "acknowledged"); err != nil {
			handleStatusUpdateError(w, err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"id": id, "status": "acknowledged"})
	})

	router.Patch("/v1/alerts/{id}/resolve", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := repo.UpdateAlertStatus(r.Context(), id, "resolved"); err != nil {
			handleStatusUpdateError(w, err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"id": id, "status": "resolved"})
	})

	router.Get("/v1/dashboard/summary", func(w http.ResponseWriter, r *http.Request) {
		summary, err := repo.DashboardSummary(r.Context())
		if err != nil {
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		httpx.WriteJSON(w, http.StatusOK, summary)
	})

	router.Get("/v1/dashboard/timeseries", func(w http.ResponseWriter, r *http.Request) {
		hours := parseBoundedInt(r.URL.Query().Get("hours"), 24, 1, 168)
		points, err := repo.DashboardTimeSeries(r.Context(), hours)
		if err != nil {
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{
			"hours": hours,
			"items": points,
		})
	})

	router.Get("/v1/dashboard/hotspots", func(w http.ResponseWriter, r *http.Request) {
		hours := parseBoundedInt(r.URL.Query().Get("hours"), 24, 1, 168)
		limit := parseBoundedInt(r.URL.Query().Get("limit"), 20, 1, 100)

		hotspots, err := repo.Hotspots(r.Context(), hours, limit)
		if err != nil {
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{
			"hours": hours,
			"items": hotspots,
		})
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

	log.Printf("query-api listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("query-api server error: %v", err)
	}
}

func parseLimit(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	if n <= 0 {
		return fallback
	}
	return n
}

func parseBoundedInt(raw string, fallback, min, max int) int {
	n := parseLimit(raw, fallback)
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

func handleStatusUpdateError(w http.ResponseWriter, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"error": "alert not found"})
		return
	}
	httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
