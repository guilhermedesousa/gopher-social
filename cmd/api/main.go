package main

import (
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/guilhermedesousa/social/internal/auth"
	"github.com/guilhermedesousa/social/internal/db"
	"github.com/guilhermedesousa/social/internal/env"
	"github.com/guilhermedesousa/social/internal/mailer"
	"github.com/guilhermedesousa/social/internal/ratelimiter"
	"github.com/guilhermedesousa/social/internal/store"
	"github.com/guilhermedesousa/social/internal/store/cache"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const version = "0.0.1"

//	@title			GopherSocial API
//	@description	API for GopherSocial application, a social network for gophers.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
//
// @securityDefinitions.apiKey	ApiKeyAuth
// @in							header
// @name						Authorization
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := config{
		addr:        env.GetString("ADDR", ":8080"),
		apiURL:      env.GetString("API_URL", "localhost:8080"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:4000"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		redisCfg: redisConfig{
			addr:    env.GetString("REDIS_ADDR", "localhost:6379"),
			pw:      env.GetString("REDIS_PW", ""),
			db:      env.GetInt("REDIS_DB", 0),
			enabled: env.GetBool("REDIS_ENABLED", false),
		},
		env: env.GetString("ENV", "development"),
		mail: mailConfig{
			exp:       time.Hour * 24 * 3, // 3 days
			fromEmail: env.GetString("FROM_EMAIL", ""),
			sendGrid: sendGridConfig{
				apikey: env.GetString("SENDGRID_API_KEY", ""),
			},
		},
		auth: authConfig{
			basic: basicConfig{
				user: env.GetString("AUTH_BASIC_USER", "admin"),
				pass: env.GetString("AUTH_BASIC_PASS", "admin"),
			},
			token: tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", "example"),
				exp:    time.Hour * 24 * 3, // 3 days
				iss:    "gophersocial",
			},
		},
		rateLimiter: ratelimiter.Config{
			RequestPerTimeFrame: env.GetInt("RATELIMITER_REQUESTS_COUNT", 20),
			TimeFrame:           time.Second * 5,
			Enabled:             env.GetBool("RATE_LIMITER_ENABLED", true),
		},
	}

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Database
	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("database connection pool established")

	// Cache
	var rdb *redis.Client
	if cfg.redisCfg.enabled {
		rdb = cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.pw, cfg.redisCfg.db)
		logger.Info("redis connection pool established")
		defer rdb.Close()
	}

	// Rate limiter
	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.rateLimiter.RequestPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	// Mailer
	mailer := mailer.NewSendgrid(cfg.mail.sendGrid.apikey, cfg.mail.fromEmail)

	jwtAuthenticator := auth.NewJWTAuthenticator(
		cfg.auth.token.secret,
		cfg.auth.token.iss,
		cfg.auth.token.iss,
	)

	store := store.NewStorage(db)
	cacheStorage := cache.NewRedisStorage(rdb)

	app := &application{
		config:        cfg,
		store:         store,
		cacheStorage:  cacheStorage,
		logger:        logger,
		mailer:        mailer,
		authenticator: jwtAuthenticator,
		rateLimiter:   rateLimiter,
	}

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
