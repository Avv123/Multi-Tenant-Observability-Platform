package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/Avv123/pulselens-platform/config"
	"github.com/Avv123/pulselens-platform/httpserver"
	"github.com/Avv123/pulselens-platform/idgen"
	"github.com/Avv123/pulselens-platform/logging"
	platformruntime "github.com/Avv123/pulselens-platform/runtime"
	appinit "github.com/Avv123/pulselens-tenant-service/init"
	tenantredis "github.com/Avv123/pulselens-tenant-service/pkg/redis"
	"github.com/Avv123/pulselens-tenant-service/router"
)

func main() {
	if err := config.MustLoadFromEnv(); err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	idgen.Configure(idgen.NodeIDFromServiceName(config.GetString("service.name")))
	if err := appinit.Preflight(ctx); err != nil {
		logging.Fatalf("startup preflight failed: %v", err)
	}
	appinit.Initialize(ctx)
	platformruntime.Start(ctx, tenantredis.Get(), platformruntime.HeartbeatOptions{
		ServiceName: config.GetString("service.name"),
		Mode:        "http",
		Port:        config.GetString("server.port"),
		Metadata: map[string]string{
			"module": "tenant",
		},
		Interval: time.Duration(config.GetInt("runtime.heartbeatIntervalSeconds")) * time.Second,
		TTL:      time.Duration(config.GetInt("runtime.heartbeatTTLSeconds")) * time.Second,
	})
	runHTTP(ctx)
}

func runHTTP(ctx context.Context) {
	server := httpserver.New(config.GetString("server.port"))
	if err := router.Initialize(ctx, server); err != nil {
		logging.Fatalf("failed to initialize router: %v", err)
	}

	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()

	logging.Infof("starting tenant-service on port %s", config.GetString("server.port"))
	if err := server.Start(); err != nil && err.Error() != "http: Server closed" {
		logging.Fatalf("failed to start tenant-service: %v", err)
		panic(err)
	}
}
