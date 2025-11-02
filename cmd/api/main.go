package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/DataDog/datadog-go/v5/statsd"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
	gorillatrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"
	redistrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/redis/go-redis.v9"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"

	"github.com/kanehiroyuu/datadog-tour/internal/infrastructure/mysql"
	infraredis "github.com/kanehiroyuu/datadog-tour/internal/infrastructure/redis"
	"github.com/kanehiroyuu/datadog-tour/internal/infrastructure/tracing"
	"github.com/kanehiroyuu/datadog-tour/internal/presentation/handler"
	"github.com/kanehiroyuu/datadog-tour/internal/usecase"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
}

func main() {
	// Start Datadog tracer (APM - 分散トレーシング)
	// tracer.Start()により、dd-trace-goライブラリがDatadog Agent（デフォルトでlocalhost:8126）に接続
	// span.Finish()が呼ばれた時に自動的にtrace-idとspan情報をDatadog Agentに送信
	// 用途: リクエストの流れを追跡（Handler → UseCase → Repository）
	tracer.Start(
		tracer.WithEnv(os.Getenv("DD_ENV")),
		tracer.WithService(os.Getenv("DD_SERVICE")),
		tracer.WithServiceVersion(os.Getenv("DD_VERSION")),
		tracer.WithLogStartup(true),
	)
	defer tracer.Stop()

	// Start Datadog profiler (継続的プロファイリング)
	// 用途: コードレベルのパフォーマンス分析（CPU使用率、メモリ割り当て）
	// CPUProfile: どの関数がCPU時間を消費しているか
	// HeapProfile: メモリ割り当てパターン、メモリリークの検出
	// 定期的にプロファイルデータをDatadog Agentに送信
	//
	// Datadogでの確認方法:
	// 1. APM → Profiling → サービスを選択
	// 2. CPU使用率を見る:
	//    - Profile Type: "CPU" を選択
	//    - Flame Graphで横幅 = CPU時間の割合、縦 = 呼び出しスタック
	//    - 関数をクリック → Self CPU (関数自体の時間), Total CPU (子関数含む)
	// 3. メモリ使用量を見る:
	//    - Profile Type: "Allocated Memory" を選択
	//    - Flame Graphで Self Allocated (関数自体の割り当て), Total Allocated (子関数含む)
	// 4. 例: userRepo.Create()が遅い → db.ExecContext()が60%のCPUを消費していることが判明
	err := profiler.Start(
		profiler.WithService(os.Getenv("DD_SERVICE")),
		profiler.WithEnv(os.Getenv("DD_ENV")),
		profiler.WithVersion(os.Getenv("DD_VERSION")),
		profiler.WithProfileTypes(
			profiler.CPUProfile,
			profiler.HeapProfile,
		),
	)
	if err != nil {
		logger.WithError(err).Warn("Failed to start profiler")
	}
	defer profiler.Stop()

	// Initialize DogStatsD client
	statsdClient, err := statsd.New(fmt.Sprintf("%s:%s",
		os.Getenv("DD_AGENT_HOST"),
		"8125"),
		statsd.WithTags([]string{
			"env:" + os.Getenv("DD_ENV"),
			"service:" + os.Getenv("DD_SERVICE"),
		}),
	)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize StatsD client")
	}
	defer statsdClient.Close()

	// Initialize MySQL with tracing
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"),
	)
	db, err := sqltrace.Open("mysql", dsn, sqltrace.WithServiceName("mysql"))
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to MySQL")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.WithError(err).Fatal("Failed to ping MySQL")
	}
	logger.Info("Successfully connected to MySQL")

	// Initialize Redis with tracing
	redisClient := redistrace.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s",
			os.Getenv("REDIS_HOST"),
			os.Getenv("REDIS_PORT"),
		),
	}, redistrace.WithServiceName("redis"))

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.WithError(err).Fatal("Failed to connect to Redis")
	}
	logger.Info("Successfully connected to Redis")

	// Setup repositories (without tracing)
	userRepoBase := mysql.NewUserRepository(db, logger)
	cacheRepoBase := infraredis.NewCacheRepository(redisClient)

	// Wrap repositories with tracing decorators
	userRepo := tracing.NewUserRepositoryTracer(userRepoBase)
	cacheRepo := tracing.NewCacheRepositoryTracer(cacheRepoBase, cacheRepoBase.GetTTL())

	// Setup use cases
	userUseCase := usecase.NewUserUseCase(userRepo, cacheRepo, logger)

	// Setup handlers
	healthHandler := handler.NewHealthHandler(logger)
	userHandler := handler.NewUserHandler(userUseCase, logger)

	// Setup router with tracing
	router := gorillatrace.NewRouter(gorillatrace.WithServiceName(os.Getenv("DD_SERVICE")))

	// Health endpoints
	router.HandleFunc("/", healthHandler.HealthCheck).Methods("GET")
	router.HandleFunc("/health", healthHandler.HealthCheck).Methods("GET")

	// User endpoints
	router.HandleFunc("/api/users", userHandler.CreateUser).Methods("POST")
	router.HandleFunc("/api/users", userHandler.GetAllUsers).Methods("GET")
	router.HandleFunc("/api/users/{id}", userHandler.GetUser).Methods("GET")

	// Setup CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
		},
		AllowCredentials: true,
		MaxAge:           300,
	})

	// Start server
	port := "8080"
	logger.WithField("port", port).Info("Starting server")
	if err := http.ListenAndServe(":"+port, corsHandler.Handler(router)); err != nil {
		logger.WithError(err).Fatal("Server failed to start")
	}
}
