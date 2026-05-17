package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"log/slog"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/glebarez/sqlite"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/zetkey/waka3x/utils"
	_ "gorm.io/driver/mysql"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	conf "github.com/zetkey/waka3x/config"
	"github.com/zetkey/waka3x/middlewares"
	"github.com/zetkey/waka3x/migrations"
	"github.com/zetkey/waka3x/repositories"
	"github.com/zetkey/waka3x/routes"
	"github.com/zetkey/waka3x/routes/api"
	shieldsV1Routes "github.com/zetkey/waka3x/routes/compat/shields/v1"
	wtV1Routes "github.com/zetkey/waka3x/routes/compat/wakatime/v1"
	"github.com/zetkey/waka3x/services"
	"github.com/zetkey/waka3x/services/mail"
	"github.com/zetkey/waka3x/static/docs"
	fsutils "github.com/zetkey/waka3x/utils/fs"
)

// Embed version.txt
//
//go:embed version.txt
var version string

// Embed static files
//
//go:embed static
var staticFiles embed.FS

//go:embed frontend/dist
var frontendFiles embed.FS

var (
	db     *gorm.DB
	config *conf.Config
)

var (
	aliasRepository           repositories.IAliasRepository
	heartbeatRepository       repositories.IHeartbeatRepository
	userRepository            repositories.IUserRepository
	languageMappingRepository repositories.ILanguageMappingRepository
	projectLabelRepository    repositories.IProjectLabelRepository
	summaryRepository         repositories.ISummaryRepository
	leaderboardRepository     *repositories.LeaderboardRepository
	keyValueRepository        repositories.IKeyValueRepository
	diagnosticsRepository     repositories.IDiagnosticsRepository
	metricsRepository         *repositories.MetricsRepository
	durationRepository        *repositories.DurationRepository
	apiKeyRepository          repositories.IApiKeyRepository
	webAuthnRepository        repositories.IWebAuthnRepository
)

var (
	aliasService           services.IAliasService
	heartbeatService       services.IHeartbeatService
	userService            services.IUserService
	languageMappingService services.ILanguageMappingService
	projectLabelService    services.IProjectLabelService
	durationService        services.IDurationService
	summaryService         services.ISummaryService
	leaderboardService     services.ILeaderboardService
	aggregationService     services.IAggregationService
	mailService            services.IMailService
	keyValueService        services.IKeyValueService
	reportService          services.IReportService
	activityService        services.IActivityService
	diagnosticsService     services.IDiagnosticsService
	housekeepingService    services.IHousekeepingService
	miscService            services.IMiscService
	apiKeyService          services.IApiKeyService
	webAuthnService        services.IWebAuthnService
)

// TODO: Refactor entire project to be structured after business domains

// @title Waka3x API
// @version 1.0
// @description REST API to interact with Waka3x.
// @description
// @description ## Authentication
// @description Set header `Authorization` to your API Key encoded as Base64 and prefixed with `Basic`
// @description **Example:** `Basic ODY2NDhkNzQtMTljNS00NTJiLWJhMDEtZmIzZWM3MGQ0YzJmCg==`

// @contact.name Ferdinand Mütsch
// @contact.url https://github.com/muety
// @contact.email ferdinand@muetsch.io

// @license.name GPL-3.0
// @license.url https://github.com/zetkey/waka3x/blob/master/LICENSE

// @securitydefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	var versionFlag = flag.Bool("version", false, "print version")
	var configFlag = flag.String("config", conf.DefaultConfigPath, "config file location")
	flag.Parse()

	if *versionFlag {
		print(version)
		os.Exit(0)
	}
	config = conf.Load(*configFlag, version)

	// Configure Swagger docs
	docs.SwaggerInfo.BasePath = config.Server.BasePath + "/api"

	slog.Info("Waka3x", "version", config.Version)

	// Set up GORM
	gormLogger := logger.New(
		log.New(os.Stdout, "", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Minute,
			Colorful:      false,
			LogLevel:      logger.Silent,
		},
	)

	// Connect to database
	var err error
	slog.Info("starting with database", "dialect", config.Db.Dialect)
	db, err = gorm.Open(
		config.Db.GetDialector(),
		&gorm.Config{Logger: gormLogger, TranslateError: true},
		conf.GetWakapiDBOpts(&config.Db),
	)
	if err != nil {
		conf.Log().Fatal("could not connect to database", "error", err)
	}

	if config.IsDev() {
		db = db.Debug()
	}
	sqlDb, err := db.DB()
	if err != nil {
		conf.Log().Fatal("could not connect to database", "error", err)
	}
	sqlDb.SetMaxIdleConns(int(config.Db.MaxConn))
	sqlDb.SetMaxOpenConns(int(config.Db.MaxConn))
	defer sqlDb.Close()

	// Migrate database schema
	if !config.SkipMigrations {
		migrations.Run(db, config)
	}

	// Repositories
	aliasRepository = repositories.NewAliasRepository(db)
	heartbeatRepository = repositories.NewHeartbeatRepository(db)
	userRepository = repositories.NewUserRepository(db)
	languageMappingRepository = repositories.NewLanguageMappingRepository(db)
	projectLabelRepository = repositories.NewProjectLabelRepository(db)
	summaryRepository = repositories.NewSummaryRepository(db)
	leaderboardRepository = repositories.NewLeaderboardRepository(db)
	keyValueRepository = repositories.NewKeyValueRepository(db)
	diagnosticsRepository = repositories.NewDiagnosticsRepository(db)
	metricsRepository = repositories.NewMetricsRepository(db)
	durationRepository = repositories.NewDurationRepository(db)
	apiKeyRepository = repositories.NewApiKeyRepository(db)
	webAuthnRepository = repositories.NewWebAuthnRepository(db)

	// Services
	mailService = mail.NewMailService()
	aliasService = services.NewAliasService(aliasRepository)
	keyValueService = services.NewKeyValueService(keyValueRepository)
	apiKeyService = services.NewApiKeyService(apiKeyRepository)
	userService = services.NewUserService(keyValueService, mailService, apiKeyService, userRepository)
	languageMappingService = services.NewLanguageMappingService(languageMappingRepository)
	projectLabelService = services.NewProjectLabelService(projectLabelRepository)
	heartbeatService = services.NewHeartbeatService(heartbeatRepository, languageMappingService)
	durationService = services.NewDurationService(durationRepository, heartbeatService, userService, languageMappingService)
	summaryService = services.NewSummaryService(summaryRepository, heartbeatService, durationService, aliasService, projectLabelService)
	aggregationService = services.NewAggregationService(userService, summaryService, heartbeatService, durationService)
	reportService = services.NewReportService(summaryService, userService, mailService)
	activityService = services.NewActivityService(summaryService)
	diagnosticsService = services.NewDiagnosticsService(diagnosticsRepository)
	housekeepingService = services.NewHousekeepingService(userService, heartbeatService, summaryService, aliasRepository) // can pass any repo here
	miscService = services.NewMiscService(userService, heartbeatService, summaryService, keyValueService, mailService)
	webAuthnService = services.NewWebAuthnService(webAuthnRepository)

	if config.App.LeaderboardEnabled {
		leaderboardService = services.NewLeaderboardService(leaderboardRepository, summaryService, userService)
	}

	// Schedule background tasks
	go conf.StartJobs()
	go aggregationService.Schedule()
	go reportService.Schedule()
	go housekeepingService.Schedule()
	go miscService.Schedule()

	if config.App.LeaderboardEnabled {
		go leaderboardService.Schedule()
	}

	// API Handlers
	rootApiHandler := api.NewApiRootHandler()
	healthApiHandler := api.NewHealthApiHandler(db)
	heartbeatApiHandler := api.NewHeartbeatApiHandler(userService, heartbeatService, languageMappingService)
	summaryApiHandler := api.NewSummaryApiHandler(userService, summaryService, heartbeatService, durationService, aliasService)
	metricsHandler := api.NewMetricsHandler(userService, summaryService, heartbeatService, leaderboardService, keyValueService, metricsRepository)
	diagnosticsHandler := api.NewDiagnosticsApiHandler(userService, diagnosticsService)
	avatarHandler := api.NewAvatarHandler()
	activityHandler := api.NewActivityApiHandler(userService, activityService)
	badgeHandler := api.NewBadgeHandler(userService, summaryService)
	captchaHandler := api.NewCaptchaHandler()
	authApiHandler := api.NewAuthApiHandler(userService, mailService, keyValueService, webAuthnService)
	projectsApiHandler := api.NewProjectsApiHandler(userService, heartbeatService)
	leaderboardApiHandler := api.NewLeaderboardApiHandler(userService, leaderboardService)
	metaApiHandler := api.NewMetaApiHandler(userService, keyValueService)
	settingsApiHandler := api.NewSettingsApiHandler(
		userService,
		heartbeatService,
		durationService,
		summaryService,
		aliasService,
		aggregationService,
		languageMappingService,
		projectLabelService,
		keyValueService,
		mailService,
		apiKeyService,
		webAuthnService,
	)

	// Compat Handlers
	wakatimeV1StatusBarHandler := wtV1Routes.NewStatusBarHandler(userService, summaryService)
	wakatimeV1AllHandler := wtV1Routes.NewAllTimeHandler(userService, summaryService)
	wakatimeV1SummariesHandler := wtV1Routes.NewSummariesHandler(userService, summaryService)
	wakatimeV1StatsHandler := wtV1Routes.NewStatsHandler(userService, summaryService)
	wakatimeV1UsersHandler := wtV1Routes.NewUsersHandler(userService, heartbeatService)
	wakatimeV1ProjectsHandler := wtV1Routes.NewProjectsHandler(userService, heartbeatService)
	wakatimeV1HeartbeatsHandler := wtV1Routes.NewHeartbeatHandler(userService, heartbeatService)
	wakatimeV1LeadersHandler := wtV1Routes.NewLeadersHandler(userService, leaderboardService)
	wakatimeV1UserAgentsHandler := wtV1Routes.NewUserAgentsHandler(userService, heartbeatService)
	shieldV1BadgeHandler := shieldsV1Routes.NewBadgeHandler(summaryService, userService)

	subscriptionHandler := routes.NewSubscriptionHandler(userService, mailService, keyValueService)

	// Setup Routing
	router := chi.NewRouter()
	router.Use(
		middleware.CleanPath,
		middleware.StripSlashes,
		middleware.Recoverer,
		middleware.GetHead,
		middlewares.NewSharedDataMiddleware(),
		middlewares.NewLoggingMiddleware(slog.Info, []string{
			"/assets",
			"/favicon",
			"/service-worker.js",
			"/api/health",
			"/api/avatar",
		}),
	)
	if config.Sentry.Dsn != "" {
		router.Use(middlewares.NewSentryMiddleware())
	}

	// Setup Sub Routers
	rootRouter := chi.NewRouter()
	rootRouter.Use(middlewares.NewSecurityMiddleware())

	apiRouter := chi.NewRouter()

	// Hook sub routers
	router.Mount("/", rootRouter)
	router.Mount("/api", apiRouter)

	// API route registrations
	rootRouter.Get("/oidc/{provider}/login", authApiHandler.GetOidcLogin)
	rootRouter.Get("/oidc/{provider}/callback", authApiHandler.GetOidcCallback)
	subscriptionHandler.RegisterRoutes(rootRouter)

	rootApiHandler.RegisterRoutes(apiRouter)
	summaryApiHandler.RegisterRoutes(apiRouter)
	healthApiHandler.RegisterRoutes(apiRouter)
	heartbeatApiHandler.RegisterRoutes(apiRouter)
	metricsHandler.RegisterRoutes(apiRouter)
	diagnosticsHandler.RegisterRoutes(apiRouter)
	avatarHandler.RegisterRoutes(apiRouter)
	activityHandler.RegisterRoutes(apiRouter)
	badgeHandler.RegisterRoutes(apiRouter)
	wakatimeV1StatusBarHandler.RegisterRoutes(apiRouter)
	wakatimeV1AllHandler.RegisterRoutes(apiRouter)
	wakatimeV1SummariesHandler.RegisterRoutes(apiRouter)
	wakatimeV1StatsHandler.RegisterRoutes(apiRouter)
	wakatimeV1UsersHandler.RegisterRoutes(apiRouter)
	wakatimeV1ProjectsHandler.RegisterRoutes(apiRouter)
	wakatimeV1HeartbeatsHandler.RegisterRoutes(apiRouter)
	wakatimeV1LeadersHandler.RegisterRoutes(apiRouter)
	wakatimeV1UserAgentsHandler.RegisterRoutes(apiRouter)
	shieldV1BadgeHandler.RegisterRoutes(apiRouter)
	captchaHandler.RegisterRoutes(apiRouter)
	authApiHandler.RegisterRoutes(apiRouter)
	projectsApiHandler.RegisterRoutes(apiRouter)
	leaderboardApiHandler.RegisterRoutes(apiRouter)
	metaApiHandler.RegisterRoutes(apiRouter)
	settingsApiHandler.RegisterRoutes(apiRouter)

	// Static Routes
	// https://github.com/golang/go/issues/43431
	embeddedStatic, _ := fs.Sub(staticFiles, "static")
	static := conf.ChooseFS("static", embeddedStatic)
	staticFileServer := http.FileServer(http.FS(fsutils.NeuteredFileSystem{FS: static}))

	router.Get("/contribute.json", staticFileServer.ServeHTTP)
	router.Get("/swagger-ui", http.RedirectHandler("swagger-ui/", http.StatusMovedPermanently).ServeHTTP) // https://github.com/swaggo/http-swagger/issues/44
	router.Get("/swagger-ui/*", httpSwagger.WrapHandler)

	// SPA Static
	embeddedFrontend, _ := fs.Sub(frontendFiles, "frontend/dist")
	frontendFs := http.FS(embeddedFrontend)
	frontendFileServer := http.FileServer(frontendFs)

	// Catch-all for SPA
	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// If requesting a file that exists in frontend/dist, serve it
		if f, err := frontendFs.Open(strings.TrimPrefix(r.URL.Path, "/")); err == nil {
			f.Close()
			frontendFileServer.ServeHTTP(w, r)
			return
		}
		// Otherwise serve index.html (SPA fallback)
		if f, err := frontendFs.Open("index.html"); err == nil {
			defer f.Close()
			http.ServeContent(w, r, "index.html", time.Now(), f)
			return
		}
		http.NotFound(w, r)
	})

	if config.EnablePprof {
		slog.Info("profiling enabled, exposing pprof data", "url", "http://127.0.0.1:6060/debug/pprof")
		go func() {
			_ = http.ListenAndServe("127.0.0.1:6060", nil)
		}()
	}

	// Listen HTTP
	listen(router)
}

func listen(handler http.Handler) {
	var s4, s6, sSocket *http.Server

	// IPv4
	if config.Server.ListenIpV4 != "-" && config.Server.ListenIpV4 != "" {
		bindString4 := config.Server.ListenIpV4 + ":" + strconv.Itoa(config.Server.Port)
		s4 = &http.Server{
			Handler:      handler,
			Addr:         bindString4,
			ReadTimeout:  time.Duration(config.Server.TimeoutSec) * time.Second,
			WriteTimeout: time.Duration(config.Server.TimeoutSec) * time.Second,
		}
	}

	// IPv6
	if config.Server.ListenIpV6 != "-" && config.Server.ListenIpV6 != "" {
		bindString6 := "[" + config.Server.ListenIpV6 + "]:" + strconv.Itoa(config.Server.Port)
		s6 = &http.Server{
			Handler:      handler,
			Addr:         bindString6,
			ReadTimeout:  time.Duration(config.Server.TimeoutSec) * time.Second,
			WriteTimeout: time.Duration(config.Server.TimeoutSec) * time.Second,
		}
	}

	// UNIX domain socket
	if config.Server.ListenSocket != "-" && config.Server.ListenSocket != "" {
		// Remove if exists
		if _, err := os.Stat(config.Server.ListenSocket); err == nil {
			slog.Info("👉 Removing unix socket", "listenSocket", config.Server.ListenSocket)
			if err := os.Remove(config.Server.ListenSocket); err != nil {
				conf.Log().Fatal(err.Error())
			}
		}
		sSocket = &http.Server{
			Handler:      handler,
			ReadTimeout:  time.Duration(config.Server.TimeoutSec) * time.Second,
			WriteTimeout: time.Duration(config.Server.TimeoutSec) * time.Second,
		}
	}

	if config.UseTLS() {
		if s4 != nil && !utils.IPv4HandledByDualStackHttp(s4, s6) { // https://github.com/zetkey/waka3x/issues/860
			slog.Info("👉 Listening for HTTPS... ✅", "address", s4.Addr)
			go func() {
				if err := s4.ListenAndServeTLS(config.Server.TlsCertPath, config.Server.TlsKeyPath); err != nil {
					err := err.Error()
					if s6 != nil {
						err += " - possibly a dual-stack problem (https://github.com/zetkey/waka3x/issues/860)?"
					}
					conf.Log().Fatal(err)
				}
			}()
		}
		if s6 != nil {
			slog.Info("👉 Listening for HTTPS... ✅", "address", s6.Addr)
			go func() {
				if err := s6.ListenAndServeTLS(config.Server.TlsCertPath, config.Server.TlsKeyPath); err != nil {
					conf.Log().Fatal(err.Error())
				}
			}()
		}
		if sSocket != nil {
			slog.Info("👉 Listening for HTTPS... ✅", "address", config.Server.ListenSocket)
			go func() {
				unixListener, err := net.Listen("unix", config.Server.ListenSocket)
				if err != nil {
					conf.Log().Fatal(err.Error())
				}
				if err := os.Chmod(config.Server.ListenSocket, os.FileMode(config.Server.ListenSocketMode)); err != nil {
					slog.Warn("failed to set user permissions for unix socket", "error", err)
				}
				if err := sSocket.ServeTLS(unixListener, config.Server.TlsCertPath, config.Server.TlsKeyPath); err != nil {
					conf.Log().Fatal(err.Error())
				}
			}()
		}
	} else {
		if s4 != nil && !utils.IPv4HandledByDualStackHttp(s4, s6) { // https://github.com/zetkey/waka3x/issues/860
			slog.Info("👉 Listening for HTTP... ✅", "address", s4.Addr)
			go func() {
				if err := s4.ListenAndServe(); err != nil {
					err := err.Error()
					if s6 != nil {
						err += " - possibly a dual-stack problem (https://github.com/zetkey/waka3x/issues/860)?"
					}
					conf.Log().Fatal(err)
				}
			}()
		}
		if s6 != nil {
			slog.Info("👉 Listening for HTTP... ✅", "address", s6.Addr)
			go func() {
				if err := s6.ListenAndServe(); err != nil {
					conf.Log().Fatal(err.Error())
				}
			}()
		}
		if sSocket != nil {
			slog.Info("👉 Listening for HTTP... ✅", "address", config.Server.ListenSocket)
			go func() {
				unixListener, err := net.Listen("unix", config.Server.ListenSocket)
				if err != nil {
					conf.Log().Fatal(err.Error())
				}
				if err := os.Chmod(config.Server.ListenSocket, os.FileMode(config.Server.ListenSocketMode)); err != nil {
					slog.Warn("failed to set user permissions for unix socket", "error", err)
				}
				if err := sSocket.Serve(unixListener); err != nil {
					conf.Log().Fatal(err.Error())
				}
			}()
		}
	}

	<-make(chan interface{}, 1)
}
