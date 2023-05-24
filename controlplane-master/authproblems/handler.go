package authproblems

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-redis/redis/v7"
	"github.com/golang/glog"

	"gitlab.slurm.io/sre_main/controlplane/sitemap"
	"gitlab.slurm.io/sre_main/controlplane/teams"
)

const maxSlowdown = 100_000

const (
	tmpl = `
		<html>
		<head>
			<style>
				body {
					margin:40px auto;
					line-height:1.6;
					font-size:18px;
					color:#444;
					padding:0 10px
				}

				h1,h2,h3 {
					line-height:1.2;
				}

				div.city {
					width: 100%;
					margin-top: 10px;
					margin-bottom: 10px;
				}

				span.field {
					margin-right: 10px;
					display: block;
				}

				input.small {
					width: 50px;
				}

				div.error {
					color: red;
					font-weight: bold;
				}
				input[type=number]::-webkit-inner-spin-button, 
				input[type=number]::-webkit-outer-spin-button { 
					-webkit-appearance: none; 
					margin: 0; 
				}

				input[type=number] {
					-moz-appearance:textfield;
				}
			</style>
		</head>
		<body>
		{{ sitemap }}
		<h1> Auth problems </h1>

		City-specific overrides default if both active. Modify one team, then submit.
		<div>
			{{ range . }}
			<div class="city">
				<form method="POST">
					<h2> {{ .Name }} </h2>
					<span class="field">
						<input name="authproblems-{{ .Name }}" type="submit" value="set {{ .Name }}"/>
						- last updated at: {{ .LastUpdate }}
					</span>
					<span class="field">
						<label for="authproblems-active-{{ .Name }}">
							Active:
						</label>

						<input 
							name="authproblems-active-{{ .Name }}"
							type="checkbox"
							{{ if .Active }}
							checked
							{{ end }}
							value="active-{{ .Name }}"
						/>
							
					<span class="field">
						<label for="authproblems-slowdown-probability-{{ .Name }}">
							Slowdown probability:
						</label>
						<input 
							name="authproblems-slowdown-probability-{{ .Name }}" 
							type="number" 
							value="{{.SlowdownProbability}}"
							step="1" min="0" max="100" value="0" 
							class="small" /><strong>%</strong>
					</span>
					<span class="field">
						<span> Slowdown range: </span>
						<input 
							name="authproblems-slowdown-min-{{ .Name }}" 
							type="number" 
							value="{{ .SlowdownMin }}"
							step="10" min="0" max="1000000" value="0" /><strong>ms</strong>
						<label for="authproblems-slowdown-max-{{ .Name }}">
						-
						</label>
						<input 
							name="authproblems-slowdown-max-{{ .Name }}" 
							type="number" 
							value="{{ .SlowdownMax }}"
							step="10" min="0" max="1000000" value="0" /><strong>ms</strong>
					</span>
					<span class="field">
						<label for="authproblems-error-probability-{{ .Name }}">
							Error probability:
						</label>
						<input 
							name="authproblems-error-probability-{{ .Name }}" 
							type="number" 
							step="1" min="0" max="100"
							value="{{ .ErrorProbability }}"
							class="small" /><strong>%</strong>
					</span>

					<span class="field">
						{{ range .Errors }}
						<div class="error"> {{ . }} </div>
						{{ end }}
					</span>
				</form>
			</div>
			<hr>
			{{ end }}
		</div>
		</body>
		</html>
	`
)

type handler struct {
	redis *redis.Client
}

func Register(redis *redis.Client) {
	sitemap.Handle("/auth-problems", &handler{
		redis: redis,
	})
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		r.ParseForm()
		var workingTeam string
		// set one team at a time
		for _, team := range teams.GetAll() {
			if r.PostForm.Get(fmt.Sprintf("authproblems-%s", team.Name)) != "" {
				workingTeam = team.Name
				break
			}
		}

		if r.PostForm.Get("authproblems-default") != "" {
			workingTeam = "default"
		}

		if workingTeam == "" {
			glog.Errorf("authproblems form submitted with no usable data (no recognized team name)")
			http.Error(w, "internal error", http.StatusBadRequest)
			return
		}

		desiredSlowdownProbRaw := r.PostForm.Get(fmt.Sprintf("authproblems-slowdown-probability-%s", workingTeam))
		desiredSlowdownProb, err := strconv.Atoi(desiredSlowdownProbRaw)
		if err != nil {
			glog.Errorf("invalid desired slowdown probability for team %q: %q / %v", workingTeam, desiredSlowdownProb, err)
			http.Error(w, fmt.Sprintf("error converting slowdown probability for team %q: %v", workingTeam, err), http.StatusBadRequest)
			return
		}

		if desiredSlowdownProb > 100 || desiredSlowdownProb < 0 {
			glog.Errorf("invalid desired slowdown probability for team %q: %q is not 0 <= x <= 100", workingTeam, desiredSlowdownProb)
			http.Error(w, "invalid desired slowdown: must be 0-100", http.StatusBadRequest)
			return
		}

		desiredSlowdownMinRaw := r.PostForm.Get(fmt.Sprintf("authproblems-slowdown-min-%s", workingTeam))
		desiredSlowdownMin, err := strconv.Atoi(desiredSlowdownMinRaw)
		if err != nil {
			glog.Errorf("invalid desired slowdown min for team %q: %q / %v", workingTeam, desiredSlowdownProb, err)
			http.Error(w, fmt.Sprintf("error converting slowdown min for team %q: %v", workingTeam, err), http.StatusBadRequest)
			return
		}

		if desiredSlowdownMin < 0 {
			glog.Errorf("invalid desired slowdown min for team %q: %q is not >= 0", workingTeam, desiredSlowdownMin)
			http.Error(w, "invalid desired min slowdown: must be >0", http.StatusBadRequest)
			return
		}

		desiredSlowdownMaxRaw := r.PostForm.Get(fmt.Sprintf("authproblems-slowdown-max-%s", workingTeam))
		desiredSlowdownMax, err := strconv.Atoi(desiredSlowdownMaxRaw)
		if err != nil {
			glog.Errorf("invalid desired slowdown max for team %q: %q / %v", workingTeam, desiredSlowdownProb, err)
			http.Error(w, fmt.Sprintf("error converting slowdown max for team %q: %v", workingTeam, err), http.StatusBadRequest)
			return
		}

		if desiredSlowdownMax > maxSlowdown || desiredSlowdownMax < 0 {
			glog.Errorf("invalid desired slowdown max for team %q: %q is >= 0", workingTeam, desiredSlowdownMin)
			http.Error(w, "invalid desired slowdown max: must be 0 <= x <= 100_000", http.StatusBadRequest)
			return
		}

		if desiredSlowdownMin > desiredSlowdownMax {
			glog.Errorf("invalid desired slowdown for team %q: min (%q) must be < max (%q)", workingTeam, desiredSlowdownMin, desiredSlowdownMax)
			http.Error(w, "invalid desired slowdown: min must be < max", http.StatusBadRequest)
			return
		}

		desiredErrorProbRaw := r.PostForm.Get(fmt.Sprintf("authproblems-error-probability-%s", workingTeam))
		desiredErrorProb, err := strconv.Atoi(desiredErrorProbRaw)
		if err != nil {
			glog.Errorf("invalid desired error probability for team %q: %q / %v", workingTeam, desiredErrorProb, err)
			http.Error(w, fmt.Sprintf("error converting desired error probability for team %q: %v", workingTeam, err), http.StatusBadRequest)
			return
		}

		if desiredErrorProb > 100 || desiredErrorProb < 0 {
			glog.Errorf("invalid desired error probability for team %q: %q is not 0 <= x <= 100", workingTeam, desiredErrorProb)
			http.Error(w, "invalid desired error: must be 0-100", http.StatusBadRequest)
			return
		}

		activeRaw := r.PostForm.Get(fmt.Sprintf("authproblems-active-%s", workingTeam))
		var active bool
		if activeRaw == fmt.Sprintf("active-%s", workingTeam) {
			active = true
		} else {
			active = false
		}

		if !active && workingTeam == "default" {
			glog.Errorf("cannot deactivate default team")
			http.Error(w, "cannot deactivate default team", http.StatusBadRequest)
			return
		}

		st := NewState(workingTeam)
		st.SlowdownProbability = desiredSlowdownProb
		st.SlowdownMin = desiredSlowdownMin
		st.SlowdownMax = desiredSlowdownMax
		st.ErrorProbability = desiredErrorProb
		st.Active = active
		err = st.Save(h.redis)
		if err != nil {
			glog.Errorf("failed to update desired state for %q: %v", workingTeam, err)
			http.Error(w, fmt.Sprintf("error setting state: %v", err), http.StatusBadRequest)
			return
		}

		fallthrough

	case http.MethodGet:
		var states []state

		// special case the "default" setting, since it isn't present as a team in K8s
		st := NewState("default")
		if err := st.ReadState(h.redis); err != nil {
			st.Errors = append(st.Errors, fmt.Errorf("failed to fetch actual state for %q: %v", "default", err))
		}
		states = append(states, st)

		for _, team := range teams.GetAll() {
			st := NewState(team.Name)
			if err := st.ReadState(h.redis); err != nil {
				st.Errors = append(st.Errors, fmt.Errorf("failed to fetch actual state for %q: %v", team.Name, err))
			}

			states = append(states, st)
		}

		if t, err := sitemap.Template("default").Parse(tmpl); err != nil {
			glog.Errorf("failed to parse template: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "text/html")
			if err := t.Execute(w, states); err != nil {
				glog.Errorf("failed to execute template: %v", err)
			}
		}
	default:
		http.Error(w, "unknown method", http.StatusMethodNotAllowed)
	}
}
