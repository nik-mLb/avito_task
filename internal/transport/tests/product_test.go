package tests

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	errs "github.com/nik-mLb/avito_task/internal/models/errs"
	models "github.com/nik-mLb/avito_task/internal/models/product"
	"github.com/nik-mLb/avito_task/internal/transport/dto"
	"github.com/nik-mLb/avito_task/internal/usecase/mocks"
	product "github.com/nik-mLb/avito_task/internal/transport/product"
)

func TestProductHandler_AddProduct(t *testing.T) {
    testTime := time.Now().Truncate(time.Second) // Округляем до секунд
    testProduct := &models.Product{
        ID:            uuid.MustParse("7b9039a7-35e0-4063-94ab-a640d887a07f"),
        ReceptionDate: testTime,
        ReceptionID:   uuid.MustParse("da480424-011d-4fc2-9452-0b7f9bb18fda"),
        ProductType:   models.Electronics,
    }
    testProductJSON := `{"id":"7b9039a7-35e0-4063-94ab-a640d887a07f","dateTime":"` + testTime.Format(time.RFC3339) + `","receptionId":"da480424-011d-4fc2-9452-0b7f9bb18fda","type":"электроника"}`

    tests := []struct {
        name           string
        requestBody    string
        mockReturn     *models.Product
        mockError      error
        expectedStatus int
        expectedBody   string
        shouldMock     bool // Флаг для определения, нужно ли мокировать вызов
    }{
        {
            name:           "successful add product",
            requestBody:    `{"type":"электроника","pvzId":"550e8400-e29b-41d4-a716-446655440000"}`,
            mockReturn:     testProduct,
            mockError:      nil,
            expectedStatus: http.StatusCreated,
            expectedBody:   testProductJSON,
            shouldMock:     true,
        },
        {
            name:           "no active reception",
            requestBody:    `{"type":"электроника","pvzId":"550e8400-e29b-41d4-a716-446655440000"}`,
            mockReturn:     nil,
            mockError:      errs.ErrNoActiveReception,
            expectedStatus: http.StatusBadRequest,
            expectedBody:   `{"message":"No active reception found"}`,
            shouldMock:     true,
        },
        {
            name:           "invalid request body - malformed JSON",
            requestBody:    `{invalid json}`,
            mockReturn:     nil,
            mockError:      nil,
            expectedStatus: http.StatusBadRequest,
            expectedBody:   `{"message":"Invalid request"}`,
            shouldMock:     false,
        },
        {
            name:           "internal server error",
            requestBody:    `{"type":"электроника","pvzId":"550e8400-e29b-41d4-a716-446655440000"}`,
            mockReturn:     nil,
            mockError:      errors.New("some error"),
            expectedStatus: http.StatusInternalServerError,
            expectedBody:   `{"message":"Failed to add product"}`,
            shouldMock:     true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockUsecase := mocks.NewMockProductUsecase(ctrl)
            h := product.NewProductHandler(mockUsecase)

            if tt.shouldMock {
                var req dto.ProductRequest
                if err := json.Unmarshal([]byte(tt.requestBody), &req); err == nil {
                    mockUsecase.EXPECT().
                        AddProduct(gomock.Any(), req.PickupPointID, req.Type).
                        Return(tt.mockReturn, tt.mockError).
                        Times(1)
                }
            }

            req := httptest.NewRequest("POST", "/products", strings.NewReader(tt.requestBody))
            w := httptest.NewRecorder()

            h.AddProduct(w, req)

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

func TestProductHandler_DeleteLastProduct(t *testing.T) {
	tests := []struct {
		name           string
		pvzID          string
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful delete",
			pvzID:          "550e8400-e29b-41d4-a716-446655440000",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "no products to delete",
			pvzID:          "550e8400-e29b-41d4-a716-446655440000",
			mockError:      errs.ErrNoProductsToDelete,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"message":"No active reception found or no products to delete"}`,
		},
		{
			name:           "internal server error",
			pvzID:          "550e8400-e29b-41d4-a716-446655440000",
			mockError:      errors.New("some error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"message":"Failed to delete product"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mocks.NewMockProductUsecase(ctrl)
			h := product.NewProductHandler(mockUsecase)

			mockUsecase.EXPECT().
				DeleteLastProduct(gomock.Any(), tt.pvzID).
				Return(tt.mockError).
				Times(1)

			req := httptest.NewRequest("DELETE", "/products/"+tt.pvzID, nil)
			req = mux.SetURLVars(req, map[string]string{"pvzId": tt.pvzID})
			w := httptest.NewRecorder()

			h.DeleteLastProduct(w, req)

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