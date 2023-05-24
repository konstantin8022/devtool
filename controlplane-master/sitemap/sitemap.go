package sitemap

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"
)

var (
	lock     sync.RWMutex
	patterns []string
)

func init() {
	patterns = append(patterns, "/metrics")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		index := Index()
		w.Write([]byte(index))
	})
}

func Handle(pattern string, handler http.Handler) {
	lock.Lock()
	defer lock.Unlock()
	patterns = append(patterns, pattern)
	http.Handle(pattern, handler)
}

func Embed() template.HTML {
	lock.RLock()
	defer lock.RUnlock()

	var result []string
	for _, pattern := range patterns {
		result = append(result, fmt.Sprintf(`<a href="%s">%s</a>`, pattern, pattern))
	}

	return template.HTML("<div>" + strings.Join(result, " | ") + "</div>")
}

func Index() string {
	var result []string
	for _, pattern := range patterns {
		result = append(result, fmt.Sprintf(`<a href="%s">%s</a>`, pattern, pattern))
	}

	return "<head> <title>Controlplane</title></head><body>" + string(Embed()) + "<body>"
}

func Template(name string) *template.Template {
	return template.New(name).Funcs(map[string]interface{}{
		"sitemap": Embed,
	})
}
