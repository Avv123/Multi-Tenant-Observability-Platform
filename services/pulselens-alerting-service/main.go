package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"
	"time"

	appinit "github.com/omniful/pulselens-alerting-service/init"
	alertworkers "github.com/omniful/pulselens-alerting-service/internal/alerts/workers"
	alertredis "github.com/omniful/pulselens-alerting-service/pkg/redis"
	"github.com/omniful/pulselens-alerting-service/router"
	"github.com/omniful/pulselens-platform/config"
	"github.com/omniful/pulselens-platform/httpserver"
	"github.com/omniful/pulselens-platform/idgen"
	"github.com/omniful/pulselens-platform/logging"
	platformruntime "github.com/omniful/pulselens-platform/runtime"
)

func main() {
	if err := config.MustLoadFromEnv(); err != nil {
		panic(err)
	}

	mode := flag.String("mode", "all", "run mode: http|worker|all")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	idgen.Configure(idgen.NodeIDFromServiceName(config.GetString("service.name")))
	if err := appinit.Preflight(ctx); err != nil {
		logging.Fatalf("startup preflight failed: %v", err)
	}
	appinit.Initialize(ctx)
	platformruntime.Start(ctx, alertredis.Get(), platformruntime.HeartbeatOptions{
		ServiceName: config.GetString("service.name"),
		Mode:        *mode,
		Port:        config.GetString("server.port"),
		Metadata: map[string]string{
			"module": "alerting",
		},
		Interval: time.Duration(config.GetInt("runtime.heartbeatIntervalSeconds")) * time.Second,
		TTL:      time.Duration(config.GetInt("runtime.heartbeatTTLSeconds")) * time.Second,
	})

	switch *mode {
	case "worker":
		if err := alertworkers.Run(ctx); err != nil {
			logging.Fatalf("worker exited: %v", err)
		}
	case "http":
		runHTTP(ctx)
	default:
		go func() {
			if err := alertworkers.Run(ctx); err != nil {
				logging.Fatalf("worker exited: %v", err)
			}
		}()
		runHTTP(ctx)
	}
}

func runHTTP(ctx context.Context) {
	server := httpserver.New(config.GetString("server.port"))
	if err := router.Initialize(ctx, server); err != nil {
		logging.Fatalf("failed to initialize alerting router: %v", err)
	}
	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()
	logging.Infof("starting alerting-service on port %s", config.GetString("server.port"))
	if err := server.Start(); err != nil && err.Error() != "http: Server closed" {
		logging.Fatalf("failed to start alerting-service: %v", err)
	}
}
