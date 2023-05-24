package redis

import (
	"net/http"

	"github.com/go-redis/redis/v7"
	"github.com/golang/glog"
	"gitlab.slurm.io/sre_main/controlplane/sitemap"
)

var HTML = `
<!doctype html>
<html>
  <head>
    <title>Redis values</title>
	<meta http-equiv="refresh" content="5">

    <body>
	{{ sitemap }}
    <form action="/redis" method="post">
    <table>
      <thead>
        <tr><th>Name<th>Value<th>
      <tbody>
	  {{ range $key, $value := . }}
      <tr>
        <td>{{ $key }}
        <td><input type=text name="{{ $key }}" value="{{ $value }}" />
          <td><input type="submit" value="Update" />
      {{ end }}
    </table>
    </form>
`

var redisKeys = []string{
	"provider_timeout",
	"is_service_mesh",
	"card_service_errors_probability",
}

type handler struct {
	r *redis.Client
}

func Register(r *redis.Client) {
	h := &handler{r: r}
	sitemap.Handle("/redis", h)
}

func (h *handler) getValue(key string) string {
	val, err := h.r.Get(key).Result()
	if err != nil {
		glog.Errorf("Error getting redis value %s: %s", key, err)
		return ""
	}
	return val
}

func (h *handler) setValue(key, value string) {
	h.r.Set(key, value, 0)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, key := range redisKeys {
			h.setValue(key, r.PostFormValue(key))
		}
	}

	p := map[string]string{}
	for _, key := range redisKeys {
		p[key] = h.getValue(key)
	}
	if t, err := sitemap.Template("default").Parse(HTML); err != nil {
		glog.Errorf("failed to parse template: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "text/html")
		if err := t.Execute(w, p); err != nil {
			glog.Errorf("failed to execute template: %v", err)
		}
	}
}
