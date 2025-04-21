package integration_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/nik-mLb/avito_task/config"
	"github.com/nik-mLb/avito_task/internal/app"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type IntegrationTestSuite struct {
	suite.Suite
	db         *sql.DB
	migrator   *migrate.Migrate
	app        *app.App
	httpServer *httptest.Server
	token      string
	client     *http.Client
}

func (s *IntegrationTestSuite) SetupSuite() {
	ctx := context.Background()

	// Запуск PostgreSQL в Docker
	req := testcontainers.ContainerRequest{
        Image:        "postgres:15-alpine",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_USER":         "test",
            "POSTGRES_PASSWORD":     "test",
            "POSTGRES_DB":           "test",
            "POSTGRES_HOST_AUTH_METHOD": "md5",
        },
        WaitingFor: wait.ForAll(
            wait.ForLog("database system is ready to accept connections"),
            wait.ForSQL(nat.Port("5432"), "postgres", func(host string, port nat.Port) string {
                return fmt.Sprintf("postgres://test:test@%s:%s/test?sslmode=disable", host, port.Port())
            }).WithStartupTimeout(45*time.Second),
        ),
    }

    postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
	s.Require().NoError(err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
    s.Require().NoError(err)
    fmt.Println("PostgreSQL port:", port)

	time.Sleep(3 * time.Second)

	connStr := fmt.Sprintf(
        "postgres://test:test@localhost:%s/test?sslmode=disable&connect_timeout=5",
        port.Port(),
    )
    fmt.Println("Connection string:", connStr)

	var db *sql.DB
    for i := 0; i < 5; i++ {
        db, err = sql.Open("postgres", connStr)
        if err == nil {
            err = db.Ping()
            if err == nil {
                break
            }
        }
        time.Sleep(2 * time.Second)
    }
    s.Require().NoError(err, "Failed to connect after 5 attempts")
    s.db = db

	// Получаем абсолютный путь к миграциям
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	migrationsPath := filepath.Join(dir, "../../db/migrations")

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		connStr,
	)
	s.Require().NoError(err)
	s.migrator = m

	err = m.Up()
	s.Require().NoError(err)

	// Тестовый конфиг
	testConfig := &config.Config{
		DBConfig: &config.DBConfig{
			User:            "test",
			Password:        "test",
			DB:              "test",
			Port:            port.Int(),
			Host:            "localhost",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: time.Minute,
		},
		ServerConfig: &config.ServerConfig{
			Port: "0",
		},
		JWTConfig: &config.JWTConfig{
			Signature:     "test_secret",
			TokenLifeSpan: time.Hour,
		},
		MigrationsConfig: &config.MigrationsConfig{
			Path: fmt.Sprintf("file://%s", migrationsPath),
		},
	}

	application, err := app.NewApp(testConfig)
	s.Require().NoError(err)
	s.app = application

	s.httpServer = httptest.NewServer(s.app.GetRouter())
	s.client = &http.Client{}
	s.getAuthToken()
}

func (s *IntegrationTestSuite) getAuthToken() {
	// Регистрация администратора
	registerData := map[string]interface{}{
		"email":    "admin@test.com",
		"password": "testpass",
		"role":     "admin",
	}
	jsonData, _ := json.Marshal(registerData)

	resp, err := http.Post(s.httpServer.URL+"/register", "application/json", bytes.NewBuffer(jsonData))
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, resp.StatusCode)

	// Логин для получения токена
	loginData := map[string]interface{}{
		"email":    "admin@test.com",
		"password": "testpass",
	}
	jsonData, _ = json.Marshal(loginData)

	resp, err = http.Post(s.httpServer.URL+"/login", "application/json", bytes.NewBuffer(jsonData))
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	// Получаем куки из ответа
	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "token" {
			s.token = cookie.Value
			break
		}
	}
	s.Require().NotEmpty(s.token, "Token cookie not found")
}

func (s *IntegrationTestSuite) newAuthenticatedRequest(method, url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: s.token,
	})

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (s *IntegrationTestSuite) TestFullFlow() {
    // 1. Создаем нового worker пользователя
    workerData := map[string]interface{}{
        "email":    "worker@test.com",
        "password": "workerpass",
        "role":     "worker",
    }
    jsonWorker, _ := json.Marshal(workerData)
    resp, err := http.Post(s.httpServer.URL+"/register", "application/json", bytes.NewBuffer(jsonWorker))
    s.Require().NoError(err)
    s.Require().Equal(http.StatusCreated, resp.StatusCode)

    // Получаем токен worker
    var workerToken string
    cookies := resp.Cookies()
    for _, cookie := range cookies {
        if cookie.Name == "token" {
            workerToken = cookie.Value
            break
        }
    }
    s.Require().NotEmpty(workerToken, "Worker token not found")

    // 3. Создаем ПВЗ с администратором (используем изначальный токен)
	pvzData := map[string]interface{}{"city": "Москва"}
    jsonData, _ := json.Marshal(pvzData)
    req, _ := s.newAuthenticatedRequest("POST", s.httpServer.URL+"/pvz", jsonData)
    resp, _ = s.client.Do(req)
    s.Require().Equal(http.StatusCreated, resp.StatusCode)

    type PVZResponse struct {
        ID   string `json:"id"`
        City string `json:"city"`
    }
    var pvzResult PVZResponse
    err = json.NewDecoder(resp.Body).Decode(&pvzResult)
    s.Require().NoError(err)
    s.Require().NotEmpty(pvzResult.ID)
    fmt.Println("Created PVZ ID:", pvzResult.ID) // Добавляем логирование

    // 4. Все последующие операции выполняем с worker
    s.token = workerToken

    // Создаем приемку
    receptionData := map[string]interface{}{
        "pvzId": pvzResult.ID, // Убедимся, что используем правильное поле
    }
    jsonData, _ = json.Marshal(receptionData)
	fmt.Println("Request body:", string(jsonData)) // Логируем тело запроса

    req, err = s.newAuthenticatedRequest("POST", s.httpServer.URL+"/receptions", jsonData)
    s.Require().NoError(err)
    
    resp, _ = s.client.Do(req)
    s.Require().Equal(http.StatusCreated, resp.StatusCode)

    var receptionResult struct{ ID string `json:"id"` }
    json.NewDecoder(resp.Body).Decode(&receptionResult)

    // Добавляем товары
    for i := 0; i < 50; i++ {
        productData := map[string]interface{}{
            "pvzId": pvzResult.ID,
            "type": "электроника",
        }
        jsonData, _ = json.Marshal(productData)
        req, _ = s.newAuthenticatedRequest("POST", s.httpServer.URL+"/products", jsonData)
        resp, _ = s.client.Do(req)
        s.Require().Equal(http.StatusCreated, resp.StatusCode)
    }

    // Закрываем приемку
    req, _ = s.newAuthenticatedRequest("POST", 
        s.httpServer.URL+fmt.Sprintf("/pvz/%s/close_last_reception", pvzResult.ID), nil)
    resp, _ = s.client.Do(req)
    s.Require().Equal(http.StatusOK, resp.StatusCode)

    // Проверяем статус
    var closedReception struct{ Status string `json:"status"` }
    json.NewDecoder(resp.Body).Decode(&closedReception)
    s.Require().Equal("close", closedReception.Status)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.migrator != nil {
		s.migrator.Down()
	}
	if s.db != nil {
		s.db.Close()
	}
	if s.httpServer != nil {
		s.httpServer.Close()
	}
}

func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	suite.Run(t, new(IntegrationTestSuite))
}