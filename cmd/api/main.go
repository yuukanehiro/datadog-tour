package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/DataDog/datadog-go/v5/statsd"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
	redistrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/redis/go-redis.v9"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

var logger *slog.Logger

func init() {
	// Create JSON handler for structured logging
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger = slog.New(handler)
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
		logger.Warn("Failed to start profiler", "error", err)
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
		logger.Error("Failed to initialize StatsD client", "error", err)
		os.Exit(1)
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
		logger.Error("Failed to connect to MySQL", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping MySQL", "error", err)
		os.Exit(1)
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
		logger.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	logger.Info("Successfully connected to Redis")

	// Setup repositories and router
	repoLocator := SetupRepositories(db, redisClient, logger)
	e := SetupRouter(logger, repoLocator)

	// Start Echo server
	port := "8080"
	logger.Info("Starting server", "port", port)
	if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
