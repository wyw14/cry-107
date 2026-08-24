package console

import (
	"net/http"
)

type Handler struct {
	pages map[string]Page
}

func NewHandler() *Handler {
	pages := []Page{KilnPage(), BurnerPage(), CoolerPage(), IncidentPage()}
	values := make(map[string]Page, len(pages))
	for _, page := range pages {
		values[page.Path] = page
	}
	return &Handler{pages: values}
}

func (h *Handler) ServePage(writer http.ResponseWriter, request *http.Request) {
	page, ok := h.pages[request.URL.Path]
	if !ok {
		http.NotFound(writer, request)
		return
	}
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := Render(writer, page); err != nil {
		http.Error(writer, "render operator page", http.StatusInternalServerError)
	}
}

func (h *Handler) Paths() []string {
	return []string{"/kiln", "/burner", "/cooler", "/incidents"}
}
