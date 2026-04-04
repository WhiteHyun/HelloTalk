package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/whitehyun/HelloTalk/apps/server/internal/auth"
	"github.com/whitehyun/HelloTalk/apps/server/internal/config"
	"github.com/whitehyun/HelloTalk/apps/server/internal/database"
	"github.com/whitehyun/HelloTalk/apps/server/internal/feed"
	"github.com/whitehyun/HelloTalk/apps/server/internal/middleware"
	"github.com/whitehyun/HelloTalk/apps/server/internal/response"
	"github.com/whitehyun/HelloTalk/apps/server/internal/user"
)

func main() {
	// 1. Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	// 2. Database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	defer db.Close()

	// 3. Migrations
	if err := database.Migrate(db, "./migrations"); err != nil {
		log.Fatal("Failed to run migrations: ", err)
	}

	// 4. Dependencies
	authStore := auth.NewStore(db)
	authService := auth.NewService(authStore, cfg)
	authHandler := auth.NewHandler(authService)

	userStore := user.NewStore(db)
	userHandler := user.NewHandler(userStore)

	feedStore := feed.NewStore(db)
	feedHandler := feed.NewHandler(feedStore)

	// 5. Router
	mux := http.NewServeMux()

	authMW := middleware.Auth(cfg.JWTSecret)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	authHandler.RegisterRoutes(mux, authMW)
	userHandler.RegisterRoutes(mux, authMW)
	feedHandler.RegisterRoutes(mux, authMW)

	// 6. Middleware chain
	handler := middleware.Logging(
		middleware.CORS(
			middleware.Recovery(mux),
		),
	)

	// 7. Server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 8. Graceful shutdown
	go func() {
		log.Printf("Server starting on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed: ", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}
	log.Println("Server stopped")
}
