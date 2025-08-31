package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	_ "modernc.org/sqlite"
)

type Job struct {
	ID      string `json:"id"`
	Project string `json:"project"`
	Path    string `json:"path"`
}

func main() {
	log.Println("Worker starting")
	rdb := redis.NewClient(&redis.Options{Addr: getenv("REDIS_ADDR", "redis:6379")})
	db, err := sql.Open("sqlite", "/data/app.db")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	for {
		res, err := rdb.BRPop(ctx, 0, "jobs").Result()
		if err != nil {
			log.Println("redis:", err)
			time.Sleep(time.Second)
			continue
		}
		if len(res) != 2 {
			continue
		}
		var job Job
		if err := json.Unmarshal([]byte(res[1]), &job); err != nil {
			log.Println("bad job:", err)
			continue
		}
		if err := process(job, db); err != nil {
			log.Println("process:", err)
		}
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func process(j Job, db *sql.DB) error {
	log.Println("Process", j.ID, j.Path)
	// Detect type and extract to allure-results directory
	workdir, err := os.MkdirTemp("", "allure-work-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workdir)
	resultsDir := filepath.Join(workdir, "allure-results")
	if err := os.MkdirAll(resultsDir, 0o755); err != nil {
		return err
	}

	if strings.HasSuffix(j.Path, ".tar.zst") {
		cmd := exec.Command("tar", "-I", "zstd", "-xf", j.Path, "-C", resultsDir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("extract tar.zst: %v out=%s", err, string(out))
		}
	} else if strings.HasSuffix(j.Path, ".zip") {
		cmd := exec.Command("unzip", "-q", j.Path, "-d", workdir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("unzip: %v out=%s", err, string(out))
		}
		// if zip produced a top-level allure-results, move it
		if _, err := os.Stat(filepath.Join(workdir, "allure-results")); err == nil {
			// already correct
		} else {
			// find it
			filepath.WalkDir(workdir, func(path string, d os.DirEntry, err error) error {
				if d != nil && d.IsDir() && filepath.Base(path) == "allure-results" {
					resultsDir = path
					return io.EOF
				}
				return nil
			})
		}
	} else {
		return fmt.Errorf("unsupported file: %s", j.Path)
	}

	// Count statuses (naive scan)
	counts := map[string]int{"total": 0, "passed": 0, "failed": 0}
	filepath.WalkDir(resultsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), "-result.json") {
			f, _ := os.Open(path)
			defer f.Close()
			s := bufio.NewScanner(f)
			for s.Scan() {
				line := s.Text()
				if strings.Contains(line, "status") {
					counts["total"]++
					if strings.Contains(line, "passed") {
						counts["passed"]++
					}
					if strings.Contains(line, "failed") {
						counts["failed"]++
					}
					break
				}
			}
		}
		return nil
	})

	// Generate allure report
	reportsBase := filepath.Join("/reports", j.Project)
	if err := os.MkdirAll(reportsBase, 0o755); err != nil {
		return err
	}
	reportDir := filepath.Join(reportsBase, j.ID)
	log.Printf("Generating report %s (total=%d passed=%d failed=%d)", reportDir, counts["total"], counts["passed"], counts["failed"])
	cmd := exec.Command("allure", "generate", resultsDir, "-o", reportDir, "--clean")
	log.Printf("Running: %s", strings.Join(cmd.Args, " "))
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("allure generate error: %v out=%s", err, string(out))
	}

	// Update DB
	_, _ = db.Exec(`INSERT OR REPLACE INTO reports(id,project,created_at,status,total,passed,failed) VALUES(?,?,?,?,?,?,?)`,
		j.ID, j.Project, time.Now().UTC().Format(time.RFC3339), status(counts), counts["total"], counts["passed"], counts["failed"])

	// Update 'latest' symlink
	os.Remove(filepath.Join(reportsBase, "latest"))
	os.Symlink(j.ID, filepath.Join(reportsBase, "latest"))

	// Enforce retention per project
	enforceRetention(db, j.Project, reportsBase, atoi(getenv("MAX_REPORTS_PER_PROJECT", "20")))

	return nil
}

func status(c map[string]int) string {
	if c["failed"] > 0 {
		return "failed"
	}
	return "passed"
}

func atoi(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	if i == 0 {
		i = 20
	}
	return i
}

func enforceRetention(db *sql.DB, project, base string, max int) {
	rows, err := db.Query(`SELECT id FROM reports WHERE project=? ORDER BY created_at DESC`, project)
	if err != nil {
		return
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		rows.Scan(&id)
		ids = append(ids, id)
	}
	if len(ids) <= max {
		return
	}
	for _, id := range ids[max:] {
		os.RemoveAll(filepath.Join(base, id))
		_, _ = db.Exec(`DELETE FROM reports WHERE id=?`, id)
	}
}
