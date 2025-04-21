package tests

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	errs "github.com/nik-mLb/avito_task/internal/models/errs"
	pickup_point "github.com/nik-mLb/avito_task/internal/models/pickup_point"
	product "github.com/nik-mLb/avito_task/internal/models/product"
	reception "github.com/nik-mLb/avito_task/internal/models/reception"
	"github.com/nik-mLb/avito_task/internal/transport/dto"
	pickup "github.com/nik-mLb/avito_task/internal/transport/pickup_point"
	"github.com/nik-mLb/avito_task/internal/usecase/mocks"
)

func normalizeJSONTime(jsonStr string) string {
    // Удаляем наносекунды из временных меток
    re := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})\.\d+([+-]\d{2}:\d{2})`)
    return re.ReplaceAllString(jsonStr, `$1$2`)
}

func compareJSON(t *testing.T, expected, actual string) {
    normExpected := normalizeJSONTime(expected)
    normActual := normalizeJSONTime(actual)
    
    if normExpected != normActual {
        t.Errorf("expected:\n%s\nbut got:\n%s", normExpected, normActual)
    }
}

func TestPickupPointHandler_CreatePickupPoint(t *testing.T) {
    testUUID1 := uuid.New()
    registrationDate := time.Now().Format(time.RFC3339)
    
    tests := []struct {
        name           string
        requestBody    string
        mockReturn     *pickup_point.PickupPoint
        mockError      error
        expectedStatus int
        expectedBody   string
    }{
        {
            name:        "successful creation",
            requestBody: `{"city": "Москва"}`,
            mockReturn: &pickup_point.PickupPoint{
                ID:               testUUID1,
                City:             "Москва",
                RegistrationDate: registrationDate,
            },
            mockError:      nil,
            expectedStatus: http.StatusCreated,
            expectedBody:   `{"id":"` + testUUID1.String() + `","city":"Москва","registrationDate":"` + registrationDate + `"}`,
        },
        {
            name:           "invalid city",
            requestBody:    `{"city": "InvalidCity"}`,
            mockReturn:     nil,
            mockError:      errs.ErrCityNotAllowed,
            expectedStatus: http.StatusBadRequest,
            expectedBody:   `{"message":"City not allowed"}`,
        },
        {
            name:           "internal server error",
            requestBody:    `{"city": "Москва"}`,
            mockReturn:     nil,
            mockError:      errors.New("some error"),
            expectedStatus: http.StatusInternalServerError,
            expectedBody:   `{"message":"Failed to create PickupPoint"}`,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockUsecase := mocks.NewMockPickupPointUsecase(ctrl)
            h := pickup.NewPickupPointHandler(mockUsecase)

            // Создаем HTTP-запрос
            httpReq := httptest.NewRequest("POST", "/pickup-points", strings.NewReader(tt.requestBody))
            
            // Отдельно парсим тело запроса для мока
            var parsedReq dto.PickupPointRequest
            if err := json.Unmarshal([]byte(tt.requestBody), &parsedReq); err == nil {
                mockUsecase.EXPECT().
                    CreatePickupPoint(gomock.Any(), parsedReq.City).
                    Return(tt.mockReturn, tt.mockError).
                    Times(1)
            }

            w := httptest.NewRecorder()

            h.CreatePickupPoint(w, httpReq)

            resp := w.Result()
            if resp.StatusCode != tt.expectedStatus {
                t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
            }

            body := strings.TrimSpace(w.Body.String())
            if body != tt.expectedBody {
                t.Errorf("expected body %s, got %s", tt.expectedBody, body)
            }
        })
    }
}

func TestPickupPointHandler_GetPickupPointsWithReceptions(t *testing.T) {
    now := time.Now()
    startDate := now.Add(-24 * time.Hour)
    endDate := now

    ppUUID1 := uuid.New()
    recUUID1 := uuid.New()
    prodUUID1 := uuid.New()
    ppUUID2 := uuid.New()
    recUUID2 := uuid.New()
    prodUUID2 := uuid.New()
    registrationDate := now.Format(time.RFC3339)
    receptionDate := now.Add(-1 * time.Hour).Format(time.RFC3339)

    tests := []struct {
        name           string
        queryParams    map[string]string
        mockReturn     []dto.PickupPointListResponse
        mockError      error
        expectedStatus int
        expectedBody   string
    }{
        {
            name: "successful request with dates",
            queryParams: map[string]string{
                "startDate": startDate.Format(time.RFC3339),
                "endDate":   endDate.Format(time.RFC3339),
                "page":      "1",
                "limit":     "10",
            },
            mockReturn: []dto.PickupPointListResponse{
                {
                    PickupPoint: pickup_point.PickupPoint{
                        ID:               ppUUID1,
                        City:             "Москва",
                        RegistrationDate: registrationDate,
                    },
                    Receptions: []dto.ReceptionWithProducts{
                        {
                            Reception: reception.Reception{
                                ID:            recUUID1,
                                ReceptionDate: now.Add(-1 * time.Hour),
                                PickupPointID: ppUUID1,
                                Status:        "in_progress",
                            },
                            Products: []product.Product{
                                {
                                    ID:            prodUUID1,
                                    ReceptionDate: now.Add(-1 * time.Hour),
                                    ReceptionID:   recUUID1,
                                    ProductType:   product.Electronics,
                                },
                            },
                        },
                    },
                },
            },
            mockError:      nil,
            expectedStatus: http.StatusOK,
            expectedBody: `[{"pvz":{"id":"` + ppUUID1.String() + `","city":"Москва","registrationDate":"` + registrationDate + `"},"receptions":[{"reception":{"id":"` + recUUID1.String() + `","dateTime":"` + receptionDate + `","pvzId":"` + ppUUID1.String() + `","status":"in_progress"},"products":[{"id":"` + prodUUID1.String() + `","dateTime":"` + receptionDate + `","receptionId":"` + recUUID1.String() + `","type":"электроника"}]}]}]`,
        },
        {
            name: "successful request without dates",
            queryParams: map[string]string{
                "page":  "2",
                "limit": "20",
            },
            mockReturn: []dto.PickupPointListResponse{
                {
                    PickupPoint: pickup_point.PickupPoint{
                        ID:               ppUUID2,
                        City:             "Saint Petersburg",
                        RegistrationDate: registrationDate,
                    },
                    Receptions: []dto.ReceptionWithProducts{
                        {
                            Reception: reception.Reception{
                                ID:            recUUID2,
                                ReceptionDate: now.Add(-1 * time.Hour),
                                PickupPointID: ppUUID2,
                                Status:        "in_progress",
                            },
                            Products: []product.Product{
                                {
                                    ID:            prodUUID2,
                                    ReceptionDate: now.Add(-1 * time.Hour),
                                    ReceptionID:   recUUID2,
                                    ProductType:   product.Clothing,
                                },
                            },
                        },
                    },
                },
            },
            mockError:      nil,
            expectedStatus: http.StatusOK,
            expectedBody: `[{"pvz":{"id":"` + ppUUID2.String() + `","city":"Saint Petersburg","registrationDate":"` + registrationDate + `"},"receptions":[{"reception":{"id":"` + recUUID2.String() + `","dateTime":"` + receptionDate + `","pvzId":"` + ppUUID2.String() + `","status":"in_progress"},"products":[{"id":"` + prodUUID2.String() + `","dateTime":"` + receptionDate + `","receptionId":"` + recUUID2.String() + `","type":"одежда"}]}]}]`,
        },
        {
            name: "invalid date format",
            queryParams: map[string]string{
                "startDate": "invalid-date",
                "endDate":   endDate.Format(time.RFC3339),
            },
            mockReturn:     []dto.PickupPointListResponse{},
            mockError:      nil,
            expectedStatus: http.StatusOK,
            expectedBody:   `[]`,
        },
        {
            name: "internal server error",
            queryParams: map[string]string{
                "page":  "1",
                "limit": "10",
            },
            mockReturn:     nil,
            mockError:      errors.New("some error"),
            expectedStatus: http.StatusInternalServerError,
            expectedBody:   `{"message":"Failed to get PickupPoints"}`,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockUsecase := mocks.NewMockPickupPointUsecase(ctrl)
            h := pickup.NewPickupPointHandler(mockUsecase)

            // Parse expected parameters from query
            var expectedStartDate, expectedEndDate *time.Time
            if startStr, ok := tt.queryParams["startDate"]; ok {
                if t, err := time.Parse(time.RFC3339, startStr); err == nil {
                    expectedStartDate = &t
                }
            }
            if endStr, ok := tt.queryParams["endDate"]; ok {
                if t, err := time.Parse(time.RFC3339, endStr); err == nil {
                    expectedEndDate = &t
                }
            }

            expectedPage := 1
            if pageStr, ok := tt.queryParams["page"]; ok {
                if page, err := strconv.Atoi(pageStr); err == nil {
                    expectedPage = page
                }
            }

            expectedLimit := 10
            if limitStr, ok := tt.queryParams["limit"]; ok {
                if limit, err := strconv.Atoi(limitStr); err == nil {
                    expectedLimit = limit
                }
            }

            // Setup mock expectation if we expect the usecase to be called
            if tt.mockError != nil || tt.mockReturn != nil {
                mockUsecase.EXPECT().
                    GetPickupPointsWithReceptions(gomock.Any(), expectedStartDate, expectedEndDate, expectedPage, expectedLimit).
                    Return(tt.mockReturn, tt.mockError).
                    Times(1)
            }

            // Build request with query parameters
            req := httptest.NewRequest("GET", "/pickup-points", nil)
            q := req.URL.Query()
            for k, v := range tt.queryParams {
                q.Add(k, v)
            }
            req.URL.RawQuery = q.Encode()

            w := httptest.NewRecorder()

            h.GetPickupPointsWithReceptions(w, req)

            resp := w.Result()
            if resp.StatusCode != tt.expectedStatus {
                t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
            }

            body := strings.TrimSpace(w.Body.String())
            compareJSON(t, tt.expectedBody, body)
        })
    }
}