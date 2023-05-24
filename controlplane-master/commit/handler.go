package commit

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v7"
	"github.com/golang/glog"
	"gitlab.slurm.io/sre_main/controlplane/gitlabapi"
	"gitlab.slurm.io/sre_main/controlplane/sitemap"
	"gitlab.slurm.io/sre_main/controlplane/teams"
)

const (
	tmpl = `
        <html>
        <head>
            <title>Commit to students repos</title>
			<meta http-equiv="refresh" content="30"/>
			<style>
				span.field {
					margin-right: 10px;
					display: block;
				}
			</style>
        </head>
        <body>
		<div>
			{{ sitemap }}
			<h1>Commit/MR to students repos</h1>

			<form method="POST">
				{{ $teams := .GetTeams }}
				{{ range .GetActions}}
				<div>
					<span class="field">
						Type: <strong>{{ .GetType }}</strong>
						<input name="{{ .GetName }}-action" type="submit" value="create" />
					<span>
					<span class="field">
						Name: <strong>{{ .GetName }}</strong>
					<span>
					<span class="field">
						Team:
						<select name="{{ .GetName }}-team">
							<option value="all">all</option>
						{{ range $teams}}
							<option value="{{ .Name }}">{{ .Name }} ({{ .ProviderBackendURL }})</option>
						{{ end }}
						</select>
					</span>
					<span class="field">
						Message: <strong>{{ .GetMessage }}</strong>
					</span>
					<span class="field">
						Files: <strong>{{ .GetFiles }}</strong>
					</span>
					<span class="field">
						Status:
						<table cellpadding="5" border="1">
							<tr>
								{{ range $teams }}
									<th>{{ .Name }}</th>
								{{ end }}
							</tr>
							<tr>
								{{ $action := . }}
								{{ range $teams }}
									{{ prettytd $action . }}
								{{ end }}
							</tr>
						</table>
					</span>
				</div>
				<hr style="margin-top: 30px"/>
				{{ end }}
			</form>
		</div>
        </body>
        </html>
	`

	all = "all"
)

type handler struct {
	actions []Action
}

func Register(glab *gitlabapi.Client, redis *redis.Client) {
	actions := []Action{
		&CommitCreateBookingsResponse201,
		&MergeRequestAdd404OnMissingSeance,
		&MergeRequestAddBookingsMetric,
	}

	for _, action := range actions {
		if commit, ok := action.(*Commit); ok {
			commit.glab = glab
			commit.redis = redis
		} else if mr, ok := action.(*MergeRequest); ok {
			mr.glab = glab
			mr.redis = redis
		}
	}

	sitemap.Handle("/commit", &handler{actions: actions})
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for key := range r.PostForm {
			if !strings.HasSuffix(key, "-action") {
				continue
			}

			name := strings.TrimSuffix(key, "-action")
			action, ok := h.GetAction(name)
			if !ok {
				continue
			}

			var teamList []teams.Team
			if team := r.PostForm.Get(name + "-team"); team == all {
				teamList = teams.GetAll()
			} else if t, ok := teams.Get(team); ok {
				teamList = append(teamList, t)
			}

			if len(teamList) == 0 {
				continue
			}

			for _, team := range teamList {
				if err := action.Do(team); err != nil {
					glog.Errorf("failed to create %s/%s: %v", action.GetName(), team.Name, err)
				} else {
					glog.Infof("successfully created %s/%s", action.GetName(), team.Name)
				}
			}
		}

		fallthrough

	case http.MethodGet:
		funcmap := map[string]interface{}{
			"prettytd": GetPrettyTableCell,
		}

		if t, err := sitemap.Template("default").Funcs(funcmap).Parse(tmpl); err != nil {
			glog.Errorf("failed to parse template: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "text/html")
			if err := t.Execute(w, h); err != nil {
				glog.Errorf("failed to execute template: %v", err)
			}
		}
	default:
		http.Error(w, "unknown method", http.StatusMethodNotAllowed)
	}
}

func (h *handler) GetAction(name string) (Action, bool) {
	for _, action := range h.actions {
		if action.GetName() == name {
			return action, true
		}
	}

	return nil, false
}

func (h *handler) GetActions() []Action {
	return h.actions
}

func (h *handler) GetTeams() []teams.Team {
	return teams.GetAll()
}

func GetPrettyTableCell(action Action, team teams.Team) template.HTML {
	status, url, err := action.GetStatus(team)
	if err != nil {
		glog.Errorf("failed to get status of %q team %q: %v", action.GetName(), team.Name, err)
		return template.HTML(`<td style="background-color: red;">error, see logs</td>`)
	}

	if status == "not commited" || status == "not created" {
		return template.HTML("<td>" + status + "</td>")
	}

	var color string
	switch status {
	case "commited + running", "can_be_merged + running":
		color = "cyan"
	case "commited + success", "can_be_merged + success":
		color = "green"
	case "commited + canceled", "can_be_merged + canceled":
		color = "yellow"
	}

	return template.HTML(fmt.Sprintf(
		`<td style="background-color: %s;"><a href="%s" target="_blank">%s</a></td>`,
		color, url, status,
	))
}
