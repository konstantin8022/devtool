package loadgenerator

import (
	"net/http"
	"strconv"

	"github.com/golang/glog"
	"gitlab.slurm.io/sre_main/controlplane/apiclient"
	"gitlab.slurm.io/sre_main/controlplane/sitemap"
	"gitlab.slurm.io/sre_main/controlplane/teams"
)

var HTML = `
<!doctype html>
<html>
  <head>
    <title>Load generator</title>
	<meta http-equiv="refresh" content="30">

	<body>
	{{ sitemap }}
    Load runners:
    <form action="/loadgen" method="post">
    <table cellpadding="5">
      <thead>
        <tr><th>Name<th>Running / Desired workers<th>Start<th>Stop<th>Params
	  <tbody>
	  <tr>
		<td>{{ .MainGenericRunner.GetName }}
		<td>
			<input type="number" name="{{ .MainGenericRunner.GetName }}-workers" value="{{ .MainGenericRunner.Workers }}" style="width: 40px" />
			<input type="submit" name="{{ .MainGenericRunner.GetName }}" value="Adjust" />
		</td>
		<td><input type="submit" name="{{ .MainGenericRunner.GetName }}" value="Go" />
		<td><input type="submit" name="{{ .MainGenericRunner.GetName }}" value="Stop" />
		<td><input type="input" name="{{ .MainGenericRunner.GetName }}-params" value="{{ .MainGenericRunner.GetParams }}" style="width: 200px" />
	  </tr>
	  <tr>
	  	<td>{{ .MainSeancesRunner.GetName }}
	  	<td>
			<input type="number" name="{{ .MainSeancesRunner.GetName }}-workers" value="{{ .MainSeancesRunner.Workers }}" style="width: 40px" />
			<input type="submit" name="{{ .MainSeancesRunner.GetName }}" value="Adjust" />
		</td>
		<td><input type="submit" name="{{ .MainSeancesRunner.GetName }}" value="Go" />
		<td><input type="submit" name="{{ .MainSeancesRunner.GetName }}" value="Stop" />
		<td><input type="input" name="{{ .MainSeancesRunner.GetName }}-params" value="{{ .MainSeancesRunner.GetParams }}" style="width: 200px" />
	</tr>
      {{ range .Runners }}
      <tr>
        <td>{{ .Strategy.Name }}
        <td>
          {{ .WorkersRunning }} /
          <input type="number" name="{{ .Strategy.Name }}-workers" value="{{ .DesiredWorkers }}" style="width: 40px" />
          <input type="submit" name="{{ .Strategy.Name }}" value="Adjust" {{ if not .IsRunning }} disabled {{ end }}/>
        </td>
        <td><input type="submit" name="{{ .Strategy.Name }}" value="Go" {{ if .IsRunning }} disabled {{ end }}/>
        <td><input type="submit" name="{{ .Strategy.Name }}" value="Stop" {{ if not .IsRunning }} disabled {{ end }}/>
        <td><input type="input" name="{{ .Strategy.Name }}-params" value="{{ .Strategy.GetParams }}" {{ if .IsRunning }} disabled {{ end }} style="width: 200px" />
      {{ end }}
    </table>
    </form>
`

type handler struct {
	mapi              *apiclient.Main
	runners           []*StrategyRunner
	mainGenericRunner *mainGenericRunner
	mainSeancesRunner *mainSeancesRunner
}

func Register(mapi *apiclient.Main) {
	h := &handler{mapi: mapi}
	sitemap.Handle("/loadgen", h)

	h.mainGenericRunner = &mainGenericRunner{Workers: 1}
	h.mainSeancesRunner = &mainSeancesRunner{Workers: 1}
	h.mainGenericRunner.params.RPS = 10
	h.mainGenericRunner.params.Lifetime = 120
	h.mainGenericRunner.params.Direct = false
	h.mainSeancesRunner.params.RPS = 10

	for _, team := range teams.GetAll() {
		if strategy, err := NewGenericStrategy(team, h.mapi); err != nil {
			glog.Errorf("failed to build NewGenericStrategy(): %v", err)
		} else {
			h.runners = append(h.runners, NewStrategyRunner(strategy))
		}
	}

	for _, team := range teams.GetAll() {
		if strategy, err := NewSeanceCreatorStrategy(team, h.mapi); err != nil {
			glog.Errorf("failed to build NewSeanceCreatorStrategy(): %v", err)
		} else {
			h.runners = append(h.runners, NewStrategyRunner(strategy))
		}
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		runMainLoadGenerator(h, r)

		for _, runner := range h.runners {
			name := runner.Strategy().Name()
			val := r.PostFormValue(name)
			if val == "" {
				continue
			}

			switch val {
			case "Go":
				workers, _ := strconv.Atoi(r.PostFormValue(name + "-workers"))
				params := r.PostFormValue(name + "-params")
				runner.Start(workers, params)
			case "Adjust":
				workers, _ := strconv.Atoi(r.PostFormValue(name + "-workers"))
				runner.Adjust(workers)
			case "Stop":
				runner.Stop()
			default:
				glog.Errorf("Invalid POST form values: %+v\n", r.PostForm)
				http.Error(w, "invalid POST input", http.StatusBadRequest)
				return
			}
		}
	}

	p := struct {
		Runners           []*StrategyRunner
		MainGenericRunner *mainGenericRunner
		MainSeancesRunner *mainSeancesRunner
	}{
		Runners:           h.runners,
		MainGenericRunner: h.mainGenericRunner,
		MainSeancesRunner: h.mainSeancesRunner,
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
