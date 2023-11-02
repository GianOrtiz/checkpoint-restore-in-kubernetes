package delivery

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/usecase"
	"github.com/google/uuid"
)

type interceptorServer struct {
	Port               int
	InterceptorUseCase usecase.InterceptorUseCase
}

func InterceptorServer(port int, interceptorUseCase usecase.InterceptorUseCase) *interceptorServer {
	return &interceptorServer{
		Port:               port,
		InterceptorUseCase: interceptorUseCase,
	}
}

func (s *interceptorServer) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.NewString()
		log.Printf("Handling request %q\n", reqID)
		res, err := s.InterceptorUseCase.InterceptRequest(reqID, r)
		log.Printf("Request %q handled with err %v and response %v\n", reqID, err, res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(res.StatusCode)
		w.Write(responseBody)
		for key, values := range res.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
	})

	mux.HandleFunc("/checkpoint", func(w http.ResponseWriter, r *http.Request) {
		if err := s.InterceptorUseCase.Checkpoint(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("intercept failed with err %v\n", err)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/reproject", func(w http.ResponseWriter, r *http.Request) {
		if err := s.InterceptorUseCase.Reproject(0); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("reprojection failed with err %v\n", err)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("Listening on port %d\n", s.Port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), mux)
}
