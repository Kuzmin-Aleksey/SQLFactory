package app

import (
	"SQLFactory/internal/app/logger"
	"SQLFactory/internal/config"
	"SQLFactory/internal/domain/service/auth"
	"SQLFactory/internal/domain/service/executor"
	"SQLFactory/internal/domain/service/sqlrunner"
	"SQLFactory/internal/infrastructure/llm/gemini"
	"SQLFactory/internal/infrastructure/persistence/mysql"
	"SQLFactory/internal/infrastructure/persistence/redis"
	"SQLFactory/internal/server/httpserver"
	"SQLFactory/pkg/contextx"
	"SQLFactory/pkg/middlwarex"
	"context"
	"errors"
	"github.com/gorilla/mux"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
)

func Run(cfg *config.Config) {
	l, err := logger.GetLogger(&cfg.Log)
	if err != nil {
		log.Fatal("failed create logger: ", err)
	}

	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	db, err := mysql.Connect(cfg.MySQl)
	if err != nil {
		log.Fatal("mysql: ", err)
	}
	defer db.Close()

	tokensCache, err := redis.NewTokensCache(cfg.Redis)
	if err != nil {
		log.Fatal("redis: ", err)
	}
	confirmEmailCache, err := redis.NewConfirmEmailCache(cfg.Redis)
	if err != nil {
		log.Fatal("redis: ", err)
	}

	llm, err := gemini.NewClient(context.Background(), "")
	if err != nil {
		log.Fatal("gemini: ", err)
	}

	usersRepo := mysql.NewUsersRepo(db)
	templatesRepo := mysql.NewTemplatesRepo(db)
	historyRepo := mysql.NewHistoryRepo(db)
	dictRepo := mysql.NewDictRepo(db)

	sqlRunnerService := sqlrunner.NewService(cfg.SQLRunner)
	executorService := executor.NewService(historyRepo, templatesRepo, dictRepo, sqlRunnerService, llm)
	authService := auth.NewService(usersRepo, tokensCache, confirmEmailCache, cfg.Auth)

	httpServer := newHttpServer(l, authService, templatesRepo, historyRepo, dictRepo, executorService, cfg.HttpServer)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-shutdown

	if err := httpServer.Shutdown(context.Background()); err != nil {
		log.Println("shutdown http server failed:", err)
	}
}

func newHttpServer(l *slog.Logger,
	authService *auth.Service,
	templatesService httpserver.TemplatesService,
	historyService httpserver.HistoryService,
	dictService httpserver.DictService,
	executorService *executor.Service,
	cfg config.HttpServerConfig,
) *http.Server {
	restAuthServer := httpserver.NewAuthServer(authService)
	restTemplatesServer := httpserver.NewTemplatesServer(templatesService)
	restHistoryServer := httpserver.NewHistoryServer(historyService)
	restDictServer := httpserver.NewDictServer(dictService)
	restExecutorServer := httpserver.NewExecutorServer(executorService)

	restServer := httpserver.NewServer(restAuthServer, restTemplatesServer, restHistoryServer, restDictServer, restExecutorServer)

	rtr := mux.NewRouter()
	restServer.RegisterRoutes(rtr)

	rtr.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filename := r.URL.Path

		if path.Ext(filename) == "" {
			if filename == "/" {
				filename = "/index"
			}
			filename += ".html"
		}
		http.ServeFile(w, r, path.Join("web", filename))
	})

	var sensitiveFields = []string{
		"password", "authorisation",
	}

	rtr.Use(
		middlwarex.AddTraceId,
		middlwarex.NewLogRequest(&middlwarex.LogOptions{
			MaxContentLen:   cfg.Log.MaxRequestContentLen,
			LoggingContent:  cfg.Log.RequestLoggingContent,
			SensitiveFields: sensitiveFields,
		}),
		middlwarex.NewLogResponse(&middlwarex.LogOptions{
			MaxContentLen:   cfg.Log.MaxResponseContentLen,
			LoggingContent:  cfg.Log.ResponseLoggingContent,
			SensitiveFields: sensitiveFields,
		}),
	)

	return &http.Server{
		Handler:      rtr,
		Addr:         cfg.Addr,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		ErrorLog:     slog.NewLogLogger(l.Handler(), slog.LevelError),
		BaseContext: func(net.Listener) context.Context {
			return contextx.WithLogger(context.Background(), l)
		},
	}
}
