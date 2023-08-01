package application

import (
	"bytes"
	"net/http"
)

type InterceptorServer struct {
	InterceptRequest *InterceptRequest
}

func (s *InterceptorServer) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		res, err := s.InterceptRequest.Execute(r)
		if err != nil {
			// TODO: change this to handle error.
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		buf := bytes.NewBuffer(make([]byte, 0, res.ContentLength))
		_, err = buf.ReadFrom(res.Body)
		if err != nil {
			// TODO: change this to handle error.
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		body := buf.Bytes()

		w.Write(body)
		// Disable superfluous WriteHeader call.
		if res.StatusCode != http.StatusOK {
			w.WriteHeader(res.StatusCode)
		}
	})
	return http.ListenAndServe(":8001", mux)
}
