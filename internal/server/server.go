package server

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/blck-snwmn/banago/internal/config"
	"github.com/blck-snwmn/banago/internal/history"
)

//go:embed templates/*.html
var templateFS embed.FS

// Server represents the web server for browsing images
type Server struct {
	projectRoot string
	port        int
	templates   *template.Template
}

// New creates a new Server instance
func New(projectRoot string, port int) *Server {
	return &Server{
		projectRoot: projectRoot,
		port:        port,
	}
}

// Start starts the web server
func (s *Server) Start() error {
	var err error
	s.templates, err = template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/subprojects/", s.handleSubproject)
	mux.HandleFunc("/entry/", s.handleEntry)
	mux.HandleFunc("/images/", s.handleImage)

	addr := fmt.Sprintf(":%d", s.port)
	return http.ListenAndServe(addr, mux)
}

// SubprojectInfo contains subproject information for templates
type SubprojectInfo struct {
	Name        string
	Description string
	EntryCount  int
}

// handleIndex shows the list of subprojects
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	subprojects, err := s.listSubprojects()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	projectConfig, _ := config.LoadProjectConfig(s.projectRoot)
	projectName := ""
	if projectConfig != nil {
		projectName = projectConfig.Name
	}

	data := struct {
		ProjectName string
		Subprojects []SubprojectInfo
	}{
		ProjectName: projectName,
		Subprojects: subprojects,
	}

	if err := s.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// EntryInfo contains entry information for templates
type EntryInfo struct {
	ID           string
	CreatedAt    string
	Success      bool
	OutputImages []string
	ImageCount   int
}

// handleSubproject shows the history entries of a subproject
func (s *Server) handleSubproject(w http.ResponseWriter, r *http.Request) {
	// Extract subproject name from /subprojects/{name}
	name := r.URL.Path[len("/subprojects/"):]
	if name == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	subprojectDir := config.GetSubprojectDir(s.projectRoot, name)
	if !config.SubprojectConfigExists(subprojectDir) {
		http.NotFound(w, r)
		return
	}

	historyDir := config.GetHistoryDir(subprojectDir)
	entries, err := history.ListEntries(historyDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reverse order (newest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ID > entries[j].ID
	})

	var entryInfos []EntryInfo
	for _, e := range entries {
		entryInfos = append(entryInfos, EntryInfo{
			ID:           e.ID,
			CreatedAt:    e.CreatedAt,
			Success:      e.Result.Success,
			OutputImages: e.Result.OutputImages,
			ImageCount:   len(e.Result.OutputImages),
		})
	}

	subprojectConfig, _ := config.LoadSubprojectConfig(subprojectDir)
	description := ""
	if subprojectConfig != nil {
		description = subprojectConfig.Description
	}

	data := struct {
		Name        string
		Description string
		Entries     []EntryInfo
	}{
		Name:        name,
		Description: description,
		Entries:     entryInfos,
	}

	if err := s.templates.ExecuteTemplate(w, "subproject.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleEntry shows a single entry with images and prompt
func (s *Server) handleEntry(w http.ResponseWriter, r *http.Request) {
	// Extract from /entry/{subproject}/{id}
	path := r.URL.Path[len("/entry/"):]
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		http.NotFound(w, r)
		return
	}

	subproject := parts[0]
	id := parts[1]
	s.renderEntry(w, subproject, id)
}

func (s *Server) renderEntry(w http.ResponseWriter, subprojectName, entryID string) {
	subprojectDir := config.GetSubprojectDir(s.projectRoot, subprojectName)
	historyDir := config.GetHistoryDir(subprojectDir)

	entry, err := history.GetEntryByID(historyDir, entryID)
	if err != nil {
		http.NotFound(w, nil)
		return
	}

	entryDir := filepath.Join(historyDir, entryID)
	prompt, _ := history.LoadPrompt(entryDir)

	// Build output image URLs
	var imageURLs []string
	for _, img := range entry.Result.OutputImages {
		imageURLs = append(imageURLs, fmt.Sprintf("/images/%s/%s/%s", subprojectName, entryID, img))
	}

	// Build input image URLs
	var inputImageURLs []string
	for _, img := range entry.Generation.InputImages {
		inputImageURLs = append(inputImageURLs, fmt.Sprintf("/images/%s/%s/%s", subprojectName, entryID, img))
	}

	// Get prev/next entry IDs for navigation
	prevID, nextID := s.getAdjacentEntryIDs(historyDir, entryID)

	data := struct {
		SubprojectName string
		Entry          *history.Entry
		Prompt         string
		ImageURLs      []string
		InputImageURLs []string
		PrevEntryID    string
		NextEntryID    string
	}{
		SubprojectName: subprojectName,
		Entry:          entry,
		Prompt:         prompt,
		ImageURLs:      imageURLs,
		InputImageURLs: inputImageURLs,
		PrevEntryID:    prevID,
		NextEntryID:    nextID,
	}

	if err := s.templates.ExecuteTemplate(w, "entry.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// getAdjacentEntryIDs returns the previous and next entry IDs for navigation
// Entries are sorted newest first, so "prev" is newer and "next" is older
func (s *Server) getAdjacentEntryIDs(historyDir, currentID string) (prevID, nextID string) {
	entries, err := history.ListEntries(historyDir)
	if err != nil || len(entries) == 0 {
		return "", ""
	}

	// Reverse to newest first (same as subproject page order)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ID > entries[j].ID
	})

	for i, e := range entries {
		if e.ID == currentID {
			if i > 0 {
				prevID = entries[i-1].ID
			}
			if i < len(entries)-1 {
				nextID = entries[i+1].ID
			}
			break
		}
	}

	return prevID, nextID
}

// handleImage serves image files from history
func (s *Server) handleImage(w http.ResponseWriter, r *http.Request) {
	// /images/{subproject}/{entryID}/{filename}
	path := r.URL.Path[len("/images/"):]

	parts := strings.SplitN(path, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		http.NotFound(w, r)
		return
	}
	subproject, entryID, filename := parts[0], parts[1], parts[2]

	subprojectDir := config.GetSubprojectDir(s.projectRoot, subproject)
	historyDir := config.GetHistoryDir(subprojectDir)
	imagePath := filepath.Join(historyDir, entryID, filename)

	// Security: ensure the path is within the history directory
	absImagePath, err := filepath.Abs(imagePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	absHistoryDir, err := filepath.Abs(historyDir)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if !strings.HasPrefix(absImagePath, absHistoryDir+string(filepath.Separator)) {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, imagePath)
}

func (s *Server) listSubprojects() ([]SubprojectInfo, error) {
	subprojectsDir := filepath.Join(s.projectRoot, config.SubprojectsDir)

	entries, err := os.ReadDir(subprojectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SubprojectInfo{}, nil
		}
		return nil, err
	}

	var result []SubprojectInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		subprojectDir := filepath.Join(subprojectsDir, entry.Name())
		if !config.SubprojectConfigExists(subprojectDir) {
			continue
		}

		cfg, _ := config.LoadSubprojectConfig(subprojectDir)
		description := ""
		if cfg != nil {
			description = cfg.Description
		}

		historyDir := config.GetHistoryDir(subprojectDir)
		historyEntries, _ := history.ListEntries(historyDir)

		result = append(result, SubprojectInfo{
			Name:        entry.Name(),
			Description: description,
			EntryCount:  len(historyEntries),
		})
	}

	return result, nil
}
