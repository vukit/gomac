package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env"
	"github.com/vukit/gomac/internal/gophermart/config"
	"github.com/vukit/gomac/internal/gophermart/logger"
	"github.com/vukit/gomac/internal/gophermart/repositories"
	"github.com/vukit/gomac/internal/gophermart/router"
	"github.com/vukit/gomac/internal/gophermart/services"
	"github.com/vukit/gomac/internal/gophermart/utils"
	"golang.org/x/sync/errgroup"
)

func main() {
	mLogger := logger.NewLogger(os.Stderr)

	mConfig := &config.Config{}
	flag.StringVar(&mConfig.RunAddress, "a", "localhost:8080", "run address")
	flag.StringVar(&mConfig.DataBaseURI, "d", "postgres://postgres:postgres@localhost:5432/gophermart?sslmode=disable", "database uri")
	flag.StringVar(&mConfig.AccrualSystemAddress, "r", "http://localhost:7070", "accrual system address")
	flag.Parse()

	err := env.Parse(mConfig)
	if err != nil {
		mLogger.Panic(err.Error())
	}

	utils.MigrationUp("file://internal/gophermart/migrations/", mConfig.DataBaseURI, mLogger)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	mRepo, err := repositories.NewRepositoryPostgreSQL(mConfig.DataBaseURI)
	if err != nil {
		mLogger.Panic(err.Error())
	}
	defer mRepo.Close()

	mRouter, err := router.NewRouter(ctx, mRepo, mLogger)
	if err != nil {
		mLogger.Panic(err.Error())
	}

	mServer := &http.Server{Addr: mConfig.RunAddress, Handler: mRouter}

	errGroup, errGroupCtx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		return mServer.ListenAndServe()
	})

	errGroup.Go(func() error {
		<-errGroupCtx.Done()

		return mServer.Shutdown(context.Background())
	})

	errGroup.Go(func() error {
		ticker := time.NewTicker(time.Second)
		loyaltyService := services.LoyaltyService{
			Address: mConfig.AccrualSystemAddress,
			Repo:    mRepo,
			Logger:  mLogger,
		}
		tasks, err := mRepo.FindTasks(ctx, "NEW", "REGISTERED", "PROCESSING")
		if err != nil {
			mLogger.Warning(err.Error())
		}
		for {
			select {
			case <-ticker.C:
				for _, task := range tasks {
					go loyaltyService.EarnPoints(ctx, task)
				}
				tasks, err = mRepo.FindTasks(ctx, "NEW")
				if err != nil {
					mLogger.Warning(err.Error())
				}
			case <-errGroupCtx.Done():
				ticker.Stop()

				return nil
			}
		}
	})

	if err := errGroup.Wait(); err != nil {
		mLogger.Info(err.Error())
	}

	cancel()
}
