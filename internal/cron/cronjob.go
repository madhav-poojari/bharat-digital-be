package cron

import (
	"context"
	"log"
	"time"

	"github.com/madhav-poojari/bharat-digital/internal/cache"
	"github.com/madhav-poojari/bharat-digital/internal/config"
	"github.com/madhav-poojari/bharat-digital/internal/models"
	"github.com/madhav-poojari/bharat-digital/internal/scraper"

	"github.com/robfig/cron/v3"
)

type Croner struct {
	cfg     *config.Config
	cache   *cache.Client
	scraper *scraper.Scraper
	cron    *cron.Cron
}

func New(cfg *config.Config, c *cache.Client, s *scraper.Scraper) *Croner {
	return &Croner{cfg: cfg, cache: c, scraper: s, cron: cron.New()}
}

func (cr *Croner) Start() error {
	_, err := cr.cron.AddFunc(cr.cfg.CronSchedule, func() { cr.RunOnce(context.Background()) })
	if err != nil {
		return err
	}
	cr.cron.Start()
	return nil
}

func (cr *Croner) Stop() context.Context {
	return cr.cron.Stop()
}

// RunOnce executes the job immediately (also used by manual trigger). It iterates FYs and
// respects a 30s sleep between each FY call (per your requirement).
func (cr *Croner) RunOnce(ctx context.Context) {
	log.Println("cron: starting job")
	for i := 0; i < len(cr.cfg.FYList); i++ {
		fy := cr.cfg.FYList[i]
		log.Printf("cron: fetching state=MAHARASHTRA fy=%s\n", fy)
		rows, err := cr.scraper.FetchCSV(ctx, "MAHARASHTRA", fy)

		if err != nil {
			log.Printf("cron: error fetching fy %s: %v\n", fy, err)
			// continue to next FY
		} else {
				monthPairs, yearAgg := scraper.RowsToCache(rows)
			// combine
			allPairs := map[string]models.CacheValue{}
			for k, v := range monthPairs {
				allPairs[k] = v
			}
			for k, v := range yearAgg {
				allPairs[k] = v
			}
			if len(allPairs) > 0 {
				if err := cr.cache.MSetJSON(ctx, allPairs); err != nil {
					log.Printf("cron: redis mset error: %v\n", err)
				} else {
					log.Printf("cron: wrote %d keys for fy=%s\n", len(allPairs), fy)
				}
			} else {
				log.Printf("cron: no rows to store for fy=%s\n", fy)
			}
		}
		// wait 30s between FY calls
		if i < len(cr.cfg.FYList)-1 {
			select {
			case <-time.After(30 * time.Second):
			case <-ctx.Done():
				log.Println("cron: cancelled during sleep")
				return
			}
		}
	}
	log.Println("cron: job finished")
}
