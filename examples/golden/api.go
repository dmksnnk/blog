package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GreetRequest struct {
	Name string `json:"name"`
}
type GreetResponse struct {
	Message string `json:"message"`
}

type API struct{}

func NewAPI() API {
	return API{}
}

func (a API) Greet(w http.ResponseWriter, r *http.Request) {
	var req GreetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	resp := GreetResponse{
		Message: fmt.Sprintf("Hello, %s!", req.Name),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func NewRouter(api API) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /greet", api.Greet)

	return mux
}
