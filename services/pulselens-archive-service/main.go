package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"
	"time"

	appinit "github.com/omniful/pulselens-archive-service/init"
	replayworkers "github.com/omniful/pulselens-archive-service/internal/replay/workers"
	archiveredis "github.com/omniful/pulselens-archive-service/pkg/redis"
	"github.com/omniful/pulselens-archive-service/router"
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
	appinit.Initialize(ctx)
	platformruntime.Start(ctx, archiveredis.Get(), platformruntime.HeartbeatOptions{
		ServiceName: config.GetString("service.name"),
		Mode:        *mode,
		Port:        config.GetString("server.port"),
		Metadata: map[string]string{
			"module": "archive",
		},
		Interval: time.Duration(config.GetInt("runtime.heartbeatIntervalSeconds")) * time.Second,
		TTL:      time.Duration(config.GetInt("runtime.heartbeatTTLSeconds")) * time.Second,
	})

	switch *mode {
	case "worker":
		if err := replayworkers.Run(ctx); err != nil {
			logging.Fatalf("worker exited: %v", err)
		}
	case "http":
		runHTTP(ctx)
	default:
		go func() {
			if err := replayworkers.Run(ctx); err != nil {
				logging.Fatalf("worker exited: %v", err)
			}
		}()
		runHTTP(ctx)
	}
}

func runHTTP(ctx context.Context) {
	server := httpserver.New(config.GetString("server.port"))
	if err := router.Initialize(ctx, server); err != nil {
		logging.Fatalf("failed to initialize archive router: %v", err)
	}
	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()
	logging.Infof("starting archive-service on port %s", config.GetString("server.port"))
	if err := server.Start(); err != nil && err.Error() != "http: Server closed" {
		logging.Fatalf("failed to start archive-service: %v", err)
	}
}
