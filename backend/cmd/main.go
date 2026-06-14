package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/config"
	"backend/internal/auth"
	"backend/internal/booking"
	"backend/internal/movie"
	"backend/internal/seed"
	"backend/internal/shared/db"
	"backend/internal/shared/email"
	"backend/internal/shared/middleware"
	"backend/internal/shared/mq"
	"backend/internal/shared/scheduler"
	"backend/internal/shared/ws"
	"backend/internal/showtime"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	// ── Load config ───────────────────────────────────────────────────────────
	cfg := config.Load()
	gin.SetMode(cfg.GinMode)

	log.Println("🎬 Cinema Booking System starting...")

	// ── Connect to MongoDB ────────────────────────────────────────────────────
	mongoDB, err := db.NewMongoDB(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatalf("❌ MongoDB connection failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := mongoDB.EnsureIndexes(ctx); err != nil {
		log.Printf("⚠️  Index creation warning: %v", err)
	}

	// ── Seed Database (if empty) ──────────────────────────────────────────────
	if err := seed.SeedDatabase(ctx, mongoDB, cfg); err != nil {
		log.Fatalf("❌ Database seeding failed: %v", err)
	}

	// ── Connect to Redis ──────────────────────────────────────────────────────
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("❌ Redis connection failed: %v", err)
	}
	log.Println("✅ Redis connected")

	// ── Repositories ─────────────────────────────────────────────────────────
	userRepo := auth.NewUserRepo(mongoDB)
	movieRepo := movie.NewRepo(mongoDB)
	showtimeRepo := showtime.NewRepo(mongoDB)
	theaterRepo := showtime.NewTheaterRepo(mongoDB)
	bookingRepo := booking.NewRepo(mongoDB)
	auditLogRepo := booking.NewAuditLogRepo(mongoDB)

	// ── Services ──────────────────────────────────────────────────────────────
	authSvc := auth.NewService(cfg, userRepo)
	lockSvc := booking.NewLockService(rdb)
	movieSvc := movie.NewService(movieRepo)
	showtimeSvc := showtime.NewService(showtimeRepo, theaterRepo, movieRepo)
	emailSvc := email.NewNotificationService(cfg.ResendAPIKey, cfg.ResendFromEmail)

	// ── Message Queue ─────────────────────────────────────────────────────────
	publisher := mq.NewPublisher(rdb)
	bookingSvc := booking.NewService(showtimeRepo, movieRepo, bookingRepo, lockSvc, publisher, auditLogRepo)

	// ── WebSocket Hub ─────────────────────────────────────────────────────────
	hub := ws.NewHub()
	go hub.Run()

	// ── MQ Consumer (async audit logging + notifications) ────────────────────
	consumer := mq.NewConsumer(rdb, emailSvc)
	consumer.Start(ctx)

	// ── Lock Expiry Scheduler ─────────────────────────────────────────────────
	lockScheduler := scheduler.NewLockExpiryScheduler(showtimeRepo, lockSvc, auditLogRepo, hub)
	lockScheduler.Start(ctx)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authHandler := auth.NewHandler(authSvc, cfg)
	movieHandler := movie.NewHandler(movieSvc, showtimeSvc)
	showtimeHandler := showtime.NewHandler(showtimeSvc)
	bookingHandler := booking.NewHandler(bookingSvc, auditLogRepo, hub)
	originPolicy := middleware.NewOriginPolicy(cfg.AppOrigins)
	wsHandler := ws.NewHandler(hub, originPolicy.IsAllowed)

	// ── Router ────────────────────────────────────────────────────────────────
	r := setupRouter(authHandler, movieHandler, showtimeHandler, bookingHandler, wsHandler, authSvc, originPolicy)
	r.GET("/ready", func(c *gin.Context) {
		readyCtx, readyCancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer readyCancel()
		if err := mongoDB.Client.Ping(readyCtx, nil); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		if err := rdb.Ping(readyCtx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// ── HTTP Server with graceful shutdown ────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server listening on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down gracefully...")
	cancel() // Stop background goroutines

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("❌ Server shutdown error: %v", err)
	}

	mongoDB.Disconnect(shutdownCtx)
	rdb.Close()

	log.Println("✅ Server stopped")
}

// setupRouter wires all routes and returns the configured Gin engine.
func setupRouter(
	authHandler *auth.Handler,
	movieHandler *movie.Handler,
	showtimeHandler *showtime.Handler,
	bookingHandler *booking.Handler,
	wsHandler *ws.Handler,
	authSvc *auth.Service,
	originPolicy *middleware.OriginPolicy,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(originPolicy.CORSMiddleware())
	r.Use(originPolicy.TrustedOriginMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// ── WebSocket ─────────────────────────────────────────────────────────────
	r.GET("/ws/showtimes/:id", wsHandler.ServeWS)

	// ── API v1 ────────────────────────────────────────────────────────────────
	api := r.Group("/api")

	// Public auth routes
	authGroup := api.Group("/auth")
	{
		authGroup.GET("/google/login", authHandler.GoogleLogin)
		authGroup.GET("/google/callback", authHandler.GoogleCallback)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/logout", authHandler.Logout)
	}

	// Protected user routes (JWT required)
	// authSvc.ValidateJWT satisfies middleware.TokenValidator (same signature)
	user := api.Group("")
	user.Use(middleware.AuthMiddleware(authSvc.ValidateJWT, auth.SessionCookieName))
	{
		user.GET("/auth/me", authHandler.Me)
		user.POST("/auth/switch-role", authHandler.SwitchRole)

		// Movies
		user.GET("/movies", movieHandler.GetAll)
		user.GET("/movies/suggestions", movieHandler.GetSuggestions)
		user.GET("/movies/:id", movieHandler.GetByID)

		// Showtimes
		user.GET("/showtimes", showtimeHandler.GetByMovieID)
		user.GET("/showtimes/:id", showtimeHandler.GetByID)

		// Booking flow
		user.POST("/bookings/locks", bookingHandler.LockSeat)
		user.DELETE("/bookings/locks", bookingHandler.ReleaseSeat)
		user.POST("/bookings", bookingHandler.ConfirmBooking)
		user.GET("/users/:id/bookings", bookingHandler.MyBookings)
	}

	// Admin routes (JWT + ADMIN role required)
	adminGroup := api.Group("")
	adminGroup.Use(middleware.AuthMiddleware(authSvc.ValidateJWT, auth.SessionCookieName))
	adminGroup.Use(middleware.RequireRole(auth.RoleAdmin))
	{
		adminGroup.GET("/bookings", bookingHandler.GetAll)
		adminGroup.GET("/bookings/stats", bookingHandler.GetStats)
		adminGroup.GET("/audit-logs", bookingHandler.GetAuditLogs)
		adminGroup.GET("/users/suggestions", authHandler.GetUserSuggestions)
		adminGroup.POST("/movies", movieHandler.Create)
	}

	return r
}
