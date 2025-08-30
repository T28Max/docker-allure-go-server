package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	_ "modernc.org/sqlite"
)

type Server struct {
	DB  *sql.DB
	RDB *redis.Client
	AuthToken string
	MaxReports int
}

type Report struct {
	ID string `json:"id"`
	Project string `json:"project"`
	CreatedAt time.Time `json:"created_at"`
	Status string `json:"status"`
	Total int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

func getenv(k, d string) string { if v := os.Getenv(k); v != "" { return v }; return d }

func main(){
	os.MkdirAll("/data", 0o755)
	db, err := sql.Open("sqlite", "/data/app.db")
	if err != nil { log.Fatal(err) }
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS reports(
		id TEXT PRIMARY KEY, project TEXT, created_at TEXT, status TEXT, total INT, passed INT, failed INT
	)`); err != nil { log.Fatal(err) }

	rdb := redis.NewClient(&redis.Options{Addr: getenv("REDIS_ADDR", "redis:6379")})
	s := &Server{DB: db, RDB: rdb, AuthToken: getenv("API_AUTH_TOKEN",""), MaxReports: atoi(getenv("MAX_REPORTS_PER_PROJECT","20"))}

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Logger)
	r.Use(cors.Handler(cors.Options{ AllowedOrigins: []string{"*"}, AllowedMethods: []string{"GET","POST","DELETE"}, AllowedHeaders: []string{"*"} }))
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request){ w.Write([]byte("ok")) })
	r.Get("/api/reports", s.listReports)
	r.Post("/api/uploads", s.requireAuth(s.handleUpload))
	r.Delete("/api/reports/{id}", s.requireAuth(s.deleteReport))

	log.Println("API listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}

func atoi(s string) int { var i int; fmt.Sscanf(s, "%d", &i); if i==0 { i=20 }; return i }

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		if s.AuthToken == "" { next(w,r); return }
		if !strings.HasPrefix(strings.ToLower(r.Header.Get("Authorization")), "bearer ") || strings.TrimSpace(r.Header.Get("Authorization")[7:]) != s.AuthToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized); return
		}
		next(w,r)
	}
}

func (s *Server) listReports(w http.ResponseWriter, r *http.Request){
	project := r.URL.Query().Get("project")
	if project == "" { project = "demo" }
	rows, err := s.DB.Query(`SELECT id, project, created_at, status, total, passed, failed FROM reports WHERE project=? ORDER BY created_at DESC`, project)
	if err != nil { http.Error(w, err.Error(), 500); return }
	defer rows.Close()
	var out []Report
	for rows.Next() {
		var rr Report; var created string
		if err := rows.Scan(&rr.ID, &rr.Project, &created, &rr.Status, &rr.Total, &rr.Passed, &rr.Failed); err != nil { continue }
		rr.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, rr)
	}
	w.Header().Set("Content-Type","application/json")
	json.NewEncoder(w).Encode(out)
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request){
	r.Body = http.MaxBytesReader(w, r.Body, 20<<30) // 20GB cap for safety (streamed to disk)
	if err := r.ParseMultipartForm(64<<20); err != nil { http.Error(w, "parse form: "+err.Error(), 400); return }
	project := r.FormValue("project"); if project=="" { project="demo" }
	file, header, err := r.FormFile("file"); if err != nil { http.Error(w, "file required", 400); return }
	defer file.Close()

	id := uuid.New().String()
	uploadDir := filepath.Join("/data","uploads", project, id)
	os.MkdirAll(uploadDir, 0o755)
	dstPath := filepath.Join(uploadDir, header.Filename)
	dst, err := os.Create(dstPath); if err != nil { http.Error(w, err.Error(), 500); return }
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil { http.Error(w, "copy: "+err.Error(), 500); return }

	// enqueue job for worker
	job := map[string]string{"id": id, "project": project, "path": dstPath}
	b, _ := json.Marshal(job)
	if err := s.RDB.LPush(context.Background(), "jobs", b).Err(); err != nil { http.Error(w, err.Error(), 500); return }

	w.Header().Set("Content-Type","application/json")
	w.Write([]byte(`{"ok":true,"id":"`+id+`"}`))
}

func (s *Server) deleteReport(w http.ResponseWriter, r *http.Request){
	id := chi.URLParam(r, "id")
	if id == "" { http.Error(w,"missing id",400); return }
	var project string
	err := s.DB.QueryRow(`SELECT project FROM reports WHERE id=?`, id).Scan(&project)
	if err != nil { http.Error(w, "not found", 404); return }

	// Remove files
	reportDir := filepath.Join("/reports", project, id)
	os.RemoveAll(reportDir)
	// If latest symlink points to it, remove and ignore
	os.Remove(filepath.Join("/reports", project, "latest"))

	_, _ = s.DB.Exec(`DELETE FROM reports WHERE id=?`, id)
	w.WriteHeader(204)
}
