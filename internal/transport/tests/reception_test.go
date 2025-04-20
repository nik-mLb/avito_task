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
	models "github.com/nik-mLb/avito_task/internal/models/reception"
	"github.com/nik-mLb/avito_task/internal/transport/dto"
	reception "github.com/nik-mLb/avito_task/internal/transport/reception"
	"github.com/nik-mLb/avito_task/internal/usecase/mocks"
)

func TestReceptionHandler_CreateReception(t *testing.T) {
	testReception := &models.Reception{
		ID:            uuid.New(),
		ReceptionDate: time.Now(),
		PickupPointID: uuid.New(),
		Status:        "in_progress",
	}

	tests := []struct {
		name           string
		requestBody    string
		mockReturn     *models.Reception
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful creation",
			requestBody:    `{"pvzId":"` + testReception.PickupPointID.String() + `"}`,
			mockReturn:     testReception,
			mockError:      nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":"` + testReception.ID.String() + `","dateTime":"` + testReception.ReceptionDate.Format(time.RFC3339) + `","pvzId":"` + testReception.PickupPointID.String() + `","status":"in_progress"}`,
		},
		{
			name:           "active reception exists",
			requestBody:    `{"pvzId":"` + testReception.PickupPointID.String() + `"}`,
			mockReturn:     nil,
			mockError:      errs.ErrActiveReceptionExists,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"message":"Active reception already exists"}`,
		},
		{
			name:           "internal server error",
			requestBody:    `{"pvzId":"` + testReception.PickupPointID.String() + `"}`,
			mockReturn:     nil,
			mockError:      errors.New("some error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"message":"Failed to create reception"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mocks.NewMockReceptionUsecase(ctrl)
			h := reception.NewReceptionHandler(mockUsecase)

			if tt.mockError != nil || tt.mockReturn != nil {
				var req dto.ReceptionRequest
				if err := json.Unmarshal([]byte(tt.requestBody), &req); err == nil {
					mockUsecase.EXPECT().
						CreateReception(gomock.Any(), req.PickupPointID).
						Return(tt.mockReturn, tt.mockError).
						Times(1)
				}
			}

			req := httptest.NewRequest("POST", "/receptions", strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()

			h.CreateReception(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			body := strings.TrimSpace(w.Body.String())
            if normalizeJSONTime(body) != normalizeJSONTime(tt.expectedBody) {
                t.Errorf("expected body %s, got %s", tt.expectedBody, body)
            }
		})
	}
}

func TestReceptionHandler_CloseReception(t *testing.T) {
	testReception := &models.Reception{
		ID:            uuid.New(),
		ReceptionDate: time.Now(),
		PickupPointID: uuid.New(),
		Status:        "close",
	}

	tests := []struct {
		name           string
		pvzID          string
		mockReturn     *models.Reception
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful close",
			pvzID:          testReception.PickupPointID.String(),
			mockReturn:     testReception,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"` + testReception.ID.String() + `","dateTime":"` + testReception.ReceptionDate.Format(time.RFC3339) + `","pvzId":"` + testReception.PickupPointID.String() + `","status":"close"}`,
		},
		{
			name:           "no active reception to close",
			pvzID:          testReception.PickupPointID.String(),
			mockReturn:     nil,
			mockError:      errs.ErrNoActiveReceptionToClose,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"message":"No active reception to close"}`,
		},
		{
			name:           "internal server error",
			pvzID:          testReception.PickupPointID.String(),
			mockReturn:     nil,
			mockError:      errors.New("some error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"message":"Failed to close reception"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mocks.NewMockReceptionUsecase(ctrl)
			h := reception.NewReceptionHandler(mockUsecase)

			mockUsecase.EXPECT().
				CloseReception(gomock.Any(), tt.pvzID).
				Return(tt.mockReturn, tt.mockError).
				Times(1)

			req := httptest.NewRequest("PUT", "/receptions/"+tt.pvzID, nil)
			req = mux.SetURLVars(req, map[string]string{"pvzId": tt.pvzID})
			w := httptest.NewRecorder()

			h.CloseReception(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			body := strings.TrimSpace(w.Body.String())
            if normalizeJSONTime(body) != normalizeJSONTime(tt.expectedBody) {
                t.Errorf("expected body %s, got %s", tt.expectedBody, body)
            }
		})
	}
}