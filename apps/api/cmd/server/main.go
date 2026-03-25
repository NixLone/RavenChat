package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ravenchat/apps/api/internal/audit"
	"ravenchat/apps/api/internal/auth"
	"ravenchat/apps/api/internal/boards"
	"ravenchat/apps/api/internal/chats"
	"ravenchat/apps/api/internal/config"
	"ravenchat/apps/api/internal/db"
	"ravenchat/apps/api/internal/files"
	"ravenchat/apps/api/internal/groups"
	"ravenchat/apps/api/internal/presence"
	"ravenchat/apps/api/internal/transport"
	"ravenchat/apps/api/internal/users"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	rdb := transport.NewRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)

	filesSvc, err := files.NewService(pool, cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket, cfg.S3UseSSL)
	if err != nil {
		log.Fatal(err)
	}
	if err := filesSvc.EnsureBucket(ctx); err != nil {
		log.Fatal(err)
	}

	usersRepo := users.NewPGRepository(pool)
	auditSvc := audit.NewService(pool)
	wsHub := transport.NewHub()
	chatSvc := chats.NewService(chats.NewRepository(pool), wsHub, auditSvc)

	server := transport.NewServer(
		auth.NewLocalProvider(usersRepo),
		auth.NewJWTManager(cfg.JWTSecret),
		usersRepo,
		chatSvc,
		groups.NewRepository(pool),
		boards.NewRepository(pool),
		filesSvc,
		auditSvc,
		presence.NewService(rdb),
		wsHub,
	)

	httpSrv := &http.Server{Addr: cfg.HTTPAddr, Handler: server.Router(cfg.FrontendURL)}
	go func() {
		log.Printf("api listening on %s", cfg.HTTPAddr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
}
