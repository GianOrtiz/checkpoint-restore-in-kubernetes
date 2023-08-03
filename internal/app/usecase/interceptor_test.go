package usecase

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/app/entity"
	"github.com/google/uuid"
)

type fakeHandler struct{}

func (h *fakeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestInterceptRequest(t *testing.T) {
	t.Run("when receiving a http request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://localhost:8000", nil)

		t.Run("it should add the request to buffer with an unique identifier", func(t *testing.T) {
			monitoredContainer := entity.Container{
				ID:      uuid.NewString(),
				HTTPUrl: "http://localhost:5000",
			}
			interceptor := entity.Interceptor{
				ID:                    uuid.NewString(),
				MonitoringContainerID: monitoredContainer.ID,
				MonitoredContainer:    &monitoredContainer,
				Config: &entity.Config{
					CheckpointingInterval: time.Duration(time.Minute * 5),
				},
			}
			useCase, _ := Interceptor(&interceptor)

			reqID := uuid.NewString()
			useCase.InterceptRequest(reqID, req)

			requests := useCase.(*interceptorUseCase).buffer
			requestIsInBufferAsUnsolved := false
			for _, r := range requests {
				if r.request == req {
					requestIsInBufferAsUnsolved = !r.solved
					break
				}
			}
			if !requestIsInBufferAsUnsolved {
				t.Error("request is not set as unsolved in buffer")
			}
		})

		t.Run("when container responds to request", func(t *testing.T) {
			// Create a fake http server and respond to the request
			handler := fakeHandler{}
			testServer := httptest.NewServer(&handler)

			monitoredContainer := entity.Container{
				ID:      uuid.NewString(),
				HTTPUrl: testServer.URL,
			}
			interceptor := entity.Interceptor{
				ID:                    uuid.NewString(),
				MonitoringContainerID: monitoredContainer.ID,
				MonitoredContainer:    &monitoredContainer,
				Config: &entity.Config{
					CheckpointingInterval: time.Duration(time.Minute * 5),
				},
			}
			useCase, _ := Interceptor(&interceptor)
			defer testServer.Close()

			req := httptest.NewRequest(http.MethodGet, testServer.URL, nil)

			t.Run("it should mark the request as solved", func(t *testing.T) {
				reqID := uuid.NewString()
				useCase.InterceptRequest(reqID, req)
				requests := useCase.(*interceptorUseCase).buffer
				requestIsInBufferAsSolved := false
				for _, r := range requests {
					if r.id == reqID {
						requestIsInBufferAsSolved = r.solved
						break
					}
				}
				if !requestIsInBufferAsSolved {
					t.Error("request is not set as solved in buffer")
				}
			})

			t.Run("it should send the response back", func(t *testing.T) {
				reqID := uuid.NewString()
				res, _ := useCase.InterceptRequest(reqID, req)
				if res.StatusCode != http.StatusOK {
					t.Errorf("expected response with status code 200, received %d", res.StatusCode)
				}
			})
		})
	})
}
