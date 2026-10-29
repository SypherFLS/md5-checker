package bootstrap

import (
	"fmt"
	"log/slog"
	"os"

	"ksp/internal/api/handlers"
	"ksp/internal/api/router"
	"ksp/internal/config"
	"ksp/internal/logger"
	"ksp/internal/repository/postgres"
	"ksp/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type App struct {
	Router *gin.Engine
	DB     *gorm.DB
	Logger *slog.Logger
}

func NewApp(cfg config.Config) (*App, error) {
	appLogger := logger.New()

	_ = godotenv.Load()

	db, err := postgres.InitDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}

	repo := postgres.NewRepo(db)
	svc := services.NewService(repo)
	handler := handlers.NewHandler(svc, appLogger)
	r := router.NewRouter(handler, cfg, appLogger)

	appLogger.Info(
		"application initialized",
		"port", cfg.Server.Port,
	)

	return &App{
		Router: r,
		DB:     db,
		Logger: appLogger,
	}, nil
}

func (a *App) Close() error {
	sqlDB, err := a.DB.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}

	return nil
}

func ConfigGet() (config.Config, error) {
	var configEnv string

	switch os.Getenv("ENV") {
	case "local":
		configEnv = "CONFIG_PATH_BACKEND"
	case "dev":
		configEnv = "CONFIG_PATH_DOCKER"
	default:
		return config.Config{}, fmt.Errorf(
			"unsupported ENV: %q",
			os.Getenv("ENV"),
		)
	}

	path := os.Getenv(configEnv)
	if path == "" {
		return config.Config{}, fmt.Errorf(
			"environment variable %s is empty",
			configEnv,
		)
	}

	cfg, err := config.InitConfig(path)
	if err != nil {
		return config.Config{}, fmt.Errorf("init config: %w", err)
	}

	return cfg, nil
}
