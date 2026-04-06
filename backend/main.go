package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type PublishRequest struct {
	Title   string `json:"title"`
	Date    string `json:"date"`
	Slug    string `json:"slug"`
	Content string `json:"content"`
	Token   string `json:"token"`
}

type VerifyTokenRequest struct {
	Token string `json:"token"`
}

type DeleteRequest struct {
	File  string `json:"file"`
	Token string `json:"token"`
}

type UpdateAboutRequest struct {
	Token     string   `json:"token"`
	Bio       string   `json:"bio"`
	TechStack []string `json:"techStack"`
	Contact   []string `json:"contact"`
}

type AboutData struct {
	Bio       string   `json:"bio"`
	TechStack []string `json:"techStack"`
	Contact   []string `json:"contact"`
	UpdatedAt string   `json:"updatedAt,omitempty"`
}

type PostItem struct {
	Title string `json:"title"`
	Date  string `json:"date"`
	URL   string `json:"url"`
}

type APIResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	File    string `json:"file,omitempty"`
}

var (
	slugRegex    = regexp.MustCompile(`^[a-z0-9-]+$`)
	dateRegex    = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	appRoot      = filepath.Clean(filepath.Join(".", ".."))
	defaultToken = "changeme-demo-token"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/verify-token", handleVerifyToken)
	mux.HandleFunc("/api/publish", handlePublish)
	mux.HandleFunc("/api/delete", handleDelete)
	mux.HandleFunc("/api/update-about", handleUpdateAbout)

	// Serve static blog files from repo root.
	mux.Handle("/", http.FileServer(http.Dir(appRoot)))

	addr := ":8080"
	log.Printf("blog api + static server running at http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, withCORS(mux)))
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,X-Admin-Token")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{OK: false, Message: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{OK: true, Message: "ok"})
}

func handleVerifyToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{OK: false, Message: "method not allowed"})
		return
	}
	var req VerifyTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{OK: false, Message: "invalid json body"})
		return
	}
	if isValidToken(req.Token) {
		writeJSON(w, http.StatusOK, APIResponse{OK: true, Message: "token valid"})
		return
	}
	writeJSON(w, http.StatusUnauthorized, APIResponse{OK: false, Message: "token invalid"})
}

func handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{OK: false, Message: "method not allowed"})
		return
	}

	var req PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{OK: false, Message: "invalid json body"})
		return
	}
	if !authorize(r.Header.Get("X-Admin-Token"), req.Token) {
		writeJSON(w, http.StatusUnauthorized, APIResponse{OK: false, Message: "token invalid"})
		return
	}
	if err := validateRequest(req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{OK: false, Message: err.Error()})
		return
	}

	postsDir := filepath.Join(appRoot, "posts")
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{OK: false, Message: "failed to create posts dir"})
		return
	}

	fileName := req.Slug + ".md"
	filePath := filepath.Join(postsDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		writeJSON(w, http.StatusConflict, APIResponse{OK: false, Message: "post file already exists"})
		return
	}

	mdContent := buildPostMarkdown(req)
	if err := os.WriteFile(filePath, []byte(mdContent), 0644); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{OK: false, Message: "failed to write post file"})
		return
	}

	url := "./post.html?file=./posts/" + fileName
	if err := upsertPostsJSON(appRoot, PostItem{Title: req.Title, Date: req.Date, URL: url}); err != nil {
		_ = os.Remove(filePath)
		writeJSON(w, http.StatusInternalServerError, APIResponse{OK: false, Message: "failed to update posts.json"})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{OK: true, Message: "markdown post published locally", File: filepath.ToSlash(filepath.Join("posts", fileName))})
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{OK: false, Message: "method not allowed"})
		return
	}
	var req DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{OK: false, Message: "invalid json body"})
		return
	}
	if !authorize(r.Header.Get("X-Admin-Token"), req.Token) {
		writeJSON(w, http.StatusUnauthorized, APIResponse{OK: false, Message: "token invalid"})
		return
	}
	if err := validateDeleteFile(req.File); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{OK: false, Message: err.Error()})
		return
	}

	rel := strings.TrimPrefix(req.File, "./")
	rel = filepath.ToSlash(rel)
	abs := filepath.Join(appRoot, filepath.FromSlash(rel))

	if _, err := os.Stat(abs); err != nil {
		writeJSON(w, http.StatusNotFound, APIResponse{OK: false, Message: "post file not found"})
		return
	}
	if err := os.Remove(abs); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{OK: false, Message: "failed to delete post file"})
		return
	}

	targetURL := "./post.html?file=./" + rel
	if err := removePostFromPostsJSON(appRoot, targetURL); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{OK: false, Message: "deleted file but failed to update posts.json"})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{OK: true, Message: "post deleted", File: req.File})
}

func handleUpdateAbout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{OK: false, Message: "method not allowed"})
		return
	}
	var req UpdateAboutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{OK: false, Message: "invalid json body"})
		return
	}
	if !authorize(r.Header.Get("X-Admin-Token"), req.Token) {
		writeJSON(w, http.StatusUnauthorized, APIResponse{OK: false, Message: "token invalid"})
		return
	}

	req.Bio = strings.TrimSpace(req.Bio)
	if req.Bio == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{OK: false, Message: "bio is required"})
		return
	}
	about := AboutData{
		Bio:       req.Bio,
		TechStack: cleanLines(req.TechStack),
		Contact:   cleanLines(req.Contact),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}
	if len(about.TechStack) == 0 {
		writeJSON(w, http.StatusBadRequest, APIResponse{OK: false, Message: "techStack is required"})
		return
	}
	if len(about.Contact) == 0 {
		writeJSON(w, http.StatusBadRequest, APIResponse{OK: false, Message: "contact is required"})
		return
	}
	if err := writeAboutJSON(appRoot, about); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{OK: false, Message: "failed to update about.json"})
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{OK: true, Message: "about updated"})
}

func validateDeleteFile(file string) error {
	file = strings.TrimSpace(file)
	if file == "" {
		return errors.New("file is required")
	}
	if !strings.HasPrefix(file, "./posts/") || !strings.HasSuffix(file, ".md") {
		return errors.New("file must be in ./posts/*.md")
	}
	if strings.Contains(file, "..") {
		return errors.New("invalid file path")
	}
	return nil
}

func authorize(headerToken, bodyToken string) bool {
	if strings.TrimSpace(headerToken) != "" {
		return isValidToken(headerToken)
	}
	return isValidToken(bodyToken)
}

func isValidToken(token string) bool {
	expected := os.Getenv("BLOG_ADMIN_TOKEN")
	if strings.TrimSpace(expected) == "" {
		expected = defaultToken
	}
	return strings.TrimSpace(token) == expected
}

func validateRequest(req PublishRequest) error {
	req.Title = strings.TrimSpace(req.Title)
	req.Date = strings.TrimSpace(req.Date)
	req.Slug = strings.TrimSpace(req.Slug)
	req.Content = strings.TrimSpace(req.Content)

	if req.Title == "" || req.Date == "" || req.Slug == "" || req.Content == "" {
		return errors.New("title/date/slug/content are required")
	}
	if !dateRegex.MatchString(req.Date) {
		return errors.New("date must be YYYY-MM-DD")
	}
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		return errors.New("date is invalid")
	}
	if !slugRegex.MatchString(req.Slug) {
		return errors.New("slug only allows lowercase letters, numbers, and '-'")
	}
	return nil
}

func buildPostMarkdown(req PublishRequest) string {
	title := html.EscapeString(strings.TrimSpace(req.Title))
	date := html.EscapeString(strings.TrimSpace(req.Date))
	content := strings.TrimSpace(req.Content)
	return fmt.Sprintf("# %s\n\n%s\n\n%s\n", title, date, content)
}

func upsertPostsJSON(rootDir string, post PostItem) error {
	file := filepath.Join(rootDir, "posts.json")
	posts := make([]PostItem, 0)

	if b, err := os.ReadFile(file); err == nil {
		_ = json.Unmarshal(b, &posts)
	}

	posts = append(posts, post)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date > posts[j].Date
	})

	out, err := json.MarshalIndent(posts, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return os.WriteFile(file, out, 0644)
}

func removePostFromPostsJSON(rootDir, targetURL string) error {
	file := filepath.Join(rootDir, "posts.json")
	posts := make([]PostItem, 0)

	if b, err := os.ReadFile(file); err == nil {
		_ = json.Unmarshal(b, &posts)
	} else {
		return err
	}

	filtered := make([]PostItem, 0, len(posts))
	for _, p := range posts {
		if p.URL != targetURL {
			filtered = append(filtered, p)
		}
	}

	out, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return os.WriteFile(file, out, 0644)
}

func writeAboutJSON(rootDir string, about AboutData) error {
	file := filepath.Join(rootDir, "about.json")
	out, err := json.MarshalIndent(about, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return os.WriteFile(file, out, 0644)
}

func cleanLines(items []string) []string {
	out := make([]string, 0, len(items))
	for _, s := range items {
		t := strings.TrimSpace(s)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
