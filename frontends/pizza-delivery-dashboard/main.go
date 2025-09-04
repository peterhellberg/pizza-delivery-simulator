package main

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
)

const port = "8234"

// Configurable drivers
var drivers = []string{
	"🧒",
	"👴",
	"👲",
}

var temporal client.Client

// WorkflowInfo represents a workflow execution
type WorkflowInfo struct {
	WorkflowID string
	RunID      string
}

// DashboardData contains workflows and drivers for template
type DashboardData struct {
	Workflows []WorkflowInfo
	Drivers   []string
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	var err error

	temporal, err = client.NewLazyClient(client.Options{
		Namespace: "default",
		Logger:    log.NewStructuredLogger(logger),
	})

	if err != nil {
		logger.Info("Unable to connect to Temporal:", "Error", err)
		os.Exit(1)
	}
	defer temporal.Close()

	http.HandleFunc("/", listWorkflows)
	http.HandleFunc("/assign", assignDriver)

	logger.Info("Pizza dashboard started at http://localhost:" + port)
	http.ListenAndServe(":"+port, nil)
}

// listWorkflows handler queries workflows and renders the template
func listWorkflows(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	resp, err := temporal.WorkflowService().ListWorkflowExecutions(ctx,
		&workflowservice.ListWorkflowExecutionsRequest{
			Namespace: "default",
			PageSize:  20,
			Query:     "DriverAssigned = false",
		},
	)
	if err != nil {
		http.Error(w, "Failed to query workflows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var results []WorkflowInfo

	for _, exec := range resp.Executions {
		results = append(results, WorkflowInfo{
			WorkflowID: exec.Execution.WorkflowId,
			RunID:      exec.Execution.RunId,
		})
	}

	data := DashboardData{Workflows: results, Drivers: drivers}

	if err := page.Execute(w, data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// assignDriver signals a workflow that a driver has been assigned and reloads dashboard
func assignDriver(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var (
		workflowID = r.FormValue("workflowID")
		runID      = r.FormValue("runID")
		driver     = r.FormValue("driver")
		ctx        = context.Background()
	)

	if err := temporal.SignalWorkflow(ctx, workflowID, runID, "DriverAccepted", driver); err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to signal workflow: %v", err),
			http.StatusInternalServerError,
		)

		return
	}

	time.Sleep(500 * time.Millisecond)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Template for dashboard
var page = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html data-theme="light">
	<head>
		<title>🍕 Pizza Delivery Dashboard 🍕</title>
		<link
	  	rel="stylesheet"
	  	href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.orange.min.css"
		>
		<style>
			:root {
    		--pico-font-size: 2rem;
  		}
		</style>
	</head>
	<body>
		<main class="container">
			<h1>Workflows waiting for driver</h1>
			<ul>
			{{range .Workflows}}
			  <li>
					<label>
			    	<b>WorkflowID:</b> {{.WorkflowID}}
						<br>
						<i>(RunID: {{.RunID}})</i>
					</label>
			    <form method="POST" action="/assign">
						<input type="hidden" name="workflowID" value="{{.WorkflowID}}">
			      <input type="hidden" name="runID" value="{{.RunID}}">
						<fieldset role="group">
			      <select name="driver">
			        {{range $.Drivers}}
			          <option value="{{.}}">{{.}}</option>
			        {{end}}
			      </select>
			      <input type="submit" value="🚗">
			    	</fieldset>
					</form>
			  </li>
			{{else}}
			<li>No workflows waiting for driver ✨</li>
			{{end}}
			</ul>
		</main>
	</body>
</html>
`))
