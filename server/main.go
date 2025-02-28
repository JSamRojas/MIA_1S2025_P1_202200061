package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	analizador "server/Analyzer"
	Utilities "server/Utilities"
)

type InputData struct {
	Code string `json:"code"`
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var inputs []string
	err := json.NewDecoder(r.Body).Decode(&inputs)
	if err != nil {
		http.Error(w, "Error al decodificar el JSON", http.StatusBadRequest)
		return
	}

	results, errors := analizador.Analyzer(inputs)

	response := map[string]interface{}{
		"results": results,
		"errors":  errors,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/analyze", handleAnalyze)

	corsHandler := Utilities.EnableCors(mux)

	fmt.Println("Servidor corriendo en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))

}
