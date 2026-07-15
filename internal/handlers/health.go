package handlers

import (
	"net/http"
	"encoding/json"
)


type HealthResponse struct{
	Status string `json:"status"`
}


func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method()!=http.MethodGet{
		http.Error(w,"method not allowed",http.StatusMethodNotAllowed)
		return
	}
	response:=HealthResponse{
		Status:"ok"
	}
	w.Header().Set("Content-Type","application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	} 
}
