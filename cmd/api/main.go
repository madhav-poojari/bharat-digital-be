package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/madhav-poojari/bharat-digital/internal/cache"
	"github.com/madhav-poojari/bharat-digital/internal/config"
	"github.com/madhav-poojari/bharat-digital/internal/cron"
	"github.com/madhav-poojari/bharat-digital/internal/handlers"
	"github.com/madhav-poojari/bharat-digital/internal/scraper"

	"github.com/gorilla/mux"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Origin", "https://bharat-digital-fe.brschess.com")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func main() {
	cfg := config.Load()

	rdb := cache.New(cfg.RedisAddr, cfg.RedisPass)
	defer rdb.Close()

	sc := scraper.New(cfg.DataGovAPIKey, cfg.HTTPTimeout)
	cr := cron.New(cfg, rdb, sc)
	// register cron route after creating cr

	app := &handlers.App{
		Cache:  rdb,
		Croner: cr,
		CFG:    cfg,
	}

	r := mux.NewRouter()
	app.RegisterRoutes(r)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: enableCORS(r),
	}

	go func() {
		if err := cr.Start(); err != nil {
			log.Printf("cron start error: %v", err)
		}
		// also run once at startup to populate
		cr.RunOnce(context.Background())
	}()

	go func() {
		log.Printf("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server err: %v", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cr.Stop()
	_ = srv.Shutdown(ctx)
}
