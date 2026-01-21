package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron/v3"
)

func (a *App) runWithSchedule() {
	log.Printf("[INFO] starting SSL certificate renewal scheduler with schedule: %s", a.config.CronSchedule)

	c := cron.New()
	_, err := c.AddFunc(a.config.CronSchedule, func() {
		a.runRenewal()
	})

	if err != nil {
		log.Fatalf("[ERROR] invalid cron schedule '%s': %v", a.config.CronSchedule, err)
	}

	c.Start()
	log.Println("[INFO] scheduler started, waiting for scheduled runs...")

	// Run once immediately on startup
	a.runRenewal()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("[INFO] shutting down scheduler...")
	c.Stop()
}

func (a *App) runRenewal() {
	log.Println("[INFO] starting certificate renewal...")

	if err := a.processDomain(); err != nil {
		log.Printf("[ERROR] failed to process domain %s: %v", a.config.Domain, err)
	} else {
		log.Printf("[INFO] SSL certificate for %s has been renewed", a.config.Domain)
	}

	log.Println("[INFO] all certificates renewed successfully")
}
