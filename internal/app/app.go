package app

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nik-mLb/avito_task/config"
	"github.com/nik-mLb/avito_task/internal/repository"
	authrepo "github.com/nik-mLb/avito_task/internal/repository/auth"
	pickuprepo "github.com/nik-mLb/avito_task/internal/repository/pickup_point"
	receptionrepo "github.com/nik-mLb/avito_task/internal/repository/reception"
	productrepo "github.com/nik-mLb/avito_task/internal/repository/product"
	autht "github.com/nik-mLb/avito_task/internal/transport/auth"
	pickupt "github.com/nik-mLb/avito_task/internal/transport/pickup_point"
	receptiont "github.com/nik-mLb/avito_task/internal/transport/reception"
	productt "github.com/nik-mLb/avito_task/internal/transport/product"
	"github.com/nik-mLb/avito_task/internal/transport/jwt"
	"github.com/nik-mLb/avito_task/internal/transport/middleware"
	authuc "github.com/nik-mLb/avito_task/internal/usecase/auth"
	pickupuc "github.com/nik-mLb/avito_task/internal/usecase/pickup_point"
	receptionuc "github.com/nik-mLb/avito_task/internal/usecase/reception"
	productuc "github.com/nik-mLb/avito_task/internal/usecase/product"
	"github.com/sirupsen/logrus"
)

// App объединяет все компоненты приложения
type App struct {
	conf   *config.Config
	logger *logrus.Logger
	db     *sql.DB
	router *mux.Router
}

// NewApp инициализирует приложение
func NewApp(conf *config.Config) (*App, error) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Подключение к БД
	dbConnStr, err := repository.GetConnectionString(conf.DBConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %v", err)
	}

	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	config.ConfigureDB(db, conf.DBConfig)

	authRepo := authrepo.New(db)
	tokenator := jwt.NewTokenator(conf.JWTConfig)
	authUC := authuc.New(authRepo, tokenator)
	authHandler := autht.New(authUC)

	pickupRepo := pickuprepo.NewPickupPointRepository(db)
	pickupUC := pickupuc.NewPickupPointUsecase(pickupRepo)
	pickupHandler := pickupt.NewPickupPointHandler(pickupUC)

	receptionRepo := receptionrepo.NewReceptionRepository(db)
	receptionUC := receptionuc.NewReceptionUsecase(receptionRepo)
	receptionHandler := receptiont.NewReceptionHandler(receptionUC)

	productRepo := productrepo.NewProductRepository(db)
	productuc := productuc.NewProductUsecase(productRepo)
	productHandler := productt.NewProductHandler(productuc)

	// Настройка маршрутизатора
	router := mux.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return middleware.LogRequest(logger, next)
	})

	router.HandleFunc("/dummyLogin", authHandler.DummyLogin).Methods("POST")
	router.HandleFunc("/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/register", authHandler.Register).Methods("POST")

	admin := router.PathPrefix("/pvz").Subrouter()
	admin.Use(middleware.AuthMiddleware(tokenator))
	admin.Use(middleware.RoleMiddleware("admin"))
	admin.HandleFunc("", pickupHandler.CreatePickupPoint).Methods("POST")

	worker := router.PathPrefix("").Subrouter()
	{
		worker.HandleFunc("/receptions", receptionHandler.CreateReception).Methods("POST")
		worker.HandleFunc("/products", productHandler.AddProduct).Methods("POST")
		worker.HandleFunc("/pvz/{pvzId}/delete_last_product", productHandler.DeleteLastProduct).Methods("POST")
		worker.HandleFunc("/pvz/{pvzId}/close_last_reception", receptionHandler.CloseReception).Methods("POST")
	}
	worker.Use(middleware.AuthMiddleware(tokenator))
	worker.Use(middleware.RoleMiddleware("worker"))

	// Добавляем новый endpoint
	reader := router.PathPrefix("/pvz").Subrouter()
	reader.Use(middleware.AuthMiddleware(tokenator))
	reader.Use(middleware.RoleMiddleware("admin", "worker"))
	reader.HandleFunc("", pickupHandler.GetPickupPointsWithReceptions).Methods("GET")

	return &App{
		conf:   conf,
		logger: logger,
		db:     db,
		router: router,
	}, nil
}

// Run запускает HTTP-сервер
func (a *App) Run() {
	server := &http.Server{
		Addr:    ":" + a.conf.ServerConfig.Port,
		Handler: a.router,
	}

	a.logger.Infof("Starting server on port %s", a.conf.ServerConfig.Port)
	if err := server.ListenAndServe(); err != nil {
		a.logger.Fatalf("Server failed: %v", err)
	}
}

func (a *App) GetRouter() *mux.Router {
    return a.router
}