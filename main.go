package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/lib/pq"
)

type ProjectRequest struct {
	ProjectName string `json:"project_name"`
	SourceText  string `json:"source_text"`
}

type ProjectReport struct {
	Name       string   `json:"project_name"`
	Modules    []string `json:"modules"`
	Entities   []string `json:"entities"`
	Complexity string   `json:"complexity"`
	SQLSchema  string   `json:"sql_schema"`
	GoStructs  string   `json:"go_structs"`
	APIDoc     string   `json:"api_doc"`
}

func main() {
	InitDB()
	defer DB.Close()

	http.HandleFunc("GET /", handleHome)
	http.HandleFunc("POST /api/analyze", handleAnalyze)
	http.HandleFunc("GET /api/projects", handleGetProjects)

	fmt.Println("Сервер запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "./ui/index.html")
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req ProjectRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.ProjectName == "" || req.SourceText == "" {
		http.Error(w, "Некорректные данные запроса", http.StatusBadRequest)
		return
	}

	var rawDataID int
	queryRaw := "INSERT INTO raw_data (project_name, source_text) VALUES ($1, $2) RETURNING id"
	err = DB.QueryRow(queryRaw, req.ProjectName, req.SourceText).Scan(&rawDataID)
	if err != nil {
		log.Println("Ошибка сохранения сырых данных:", err)
		http.Error(w, "Ошибка сохранения данных", http.StatusInternalServerError)
		return
	}

	analysis := AnalyzePrimaryData(req.SourceText)

	queryAnalysis := `INSERT INTO crm_analysis (raw_data_id, recommended_modules, extracted_entities, complexity_level, generated_sql, generated_structs, api_documentation) 
                      VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = DB.Exec(queryAnalysis, rawDataID, pq.Array(analysis.Modules), pq.Array(analysis.Entities), analysis.Complexity, analysis.SQLSchema, analysis.GoStructs, analysis.APIDoc)
	if err != nil {
		log.Println("Ошибка сохранения результатов анализа:", err)
		http.Error(w, "Ошибка сохранения анализа", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysis)
}

func handleGetProjects(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query(`SELECT r.project_name, a.recommended_modules, a.extracted_entities, a.complexity_level, a.generated_sql, a.generated_structs, a.api_documentation 
                           FROM raw_data r JOIN crm_analysis a ON r.id = a.raw_data_id ORDER BY a.created_at DESC`)
	if err != nil {
		log.Println("Ошибка получения истории проектов:", err)
		http.Error(w, "Ошибка получения данных", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reports []ProjectReport
	for rows.Next() {
		var rep ProjectReport
		err := rows.Scan(&rep.Name, pq.Array(&rep.Modules), pq.Array(&rep.Entities), &rep.Complexity, &rep.SQLSchema, &rep.GoStructs, &rep.APIDoc)
		if err == nil {
			reports = append(reports, rep)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}
