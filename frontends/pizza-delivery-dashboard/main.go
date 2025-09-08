package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go.temporal.io/api/common/v1"
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

// Order represents a workflow execution
type Order struct {
	WorkflowID string
	RunID      string
	Info       OrderInfo
}

type OrderInfo struct {
	Customer struct {
		Name  string
		Addr  string
		Error error
	}
	Pizza struct {
		Nr    int
		Name  string
		Error error
	}
}

// DashboardData contains workflows and drivers for template
type DashboardData struct {
	Orders  []Order
	Drivers []string
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	temporal, err := client.NewLazyClient(client.Options{
		Namespace: "default",
		Logger:    log.NewStructuredLogger(logger),
	})

	if err != nil {
		logger.Info("Unable to connect to Temporal:", "Error", err)
		os.Exit(1)
	}
	defer temporal.Close()

	d := &Dashboard{
		Client: temporal,
		Logger: logger,
	}

	http.HandleFunc("/", d.listWorkflows)
	http.HandleFunc("/assign", d.assignDriver)

	logger.Info("Pizza dashboard started at http://localhost:" + port)
	http.ListenAndServe(":"+port, nil)
}

// The main type for the dashboard, such as the endpoints
type Dashboard struct {
	client.Client
	log.Logger
}

// listWorkflows handler queries workflows and renders the template
func (d *Dashboard) listWorkflows(w http.ResponseWriter, r *http.Request) {
	resp, err := d.WorkflowService().ListWorkflowExecutions(r.Context(),
		&workflowservice.ListWorkflowExecutionsRequest{
			Namespace: "default",
			PageSize:  100,
			Query:     `DriverAssigned = false`,
		},
	)
	if err != nil {
		http.Error(w, "Failed to query workflows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var orders []Order

	for _, exec := range resp.Executions {
		memo := OrderInfo{}

		if exec.Memo != nil {
			fields := exec.Memo.GetFields()

			if err := getMemoField(fields, "Customer", &memo.Customer); err != nil {
				d.Info("Failed to get memo field Customer", "error", err)
			}

			if err := getMemoField(fields, "Pizza", &memo.Pizza); err != nil {
				d.Info("Failed to get memo field Pizza", "error", err)
			}
		}

		orders = append(orders, Order{
			WorkflowID: exec.Execution.WorkflowId,
			RunID:      exec.Execution.RunId,
			Info:       memo,
		})
	}

	if err := page.Execute(w, DashboardData{
		Orders:  orders,
		Drivers: drivers,
	}); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// assignDriver signals a workflow that a driver has been assigned and reloads dashboard
func (d *Dashboard) assignDriver(w http.ResponseWriter, r *http.Request) {
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

	if err := d.SignalWorkflow(ctx, workflowID, runID, "DriverAccepted", driver); err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to signal workflow: %v", err),
			http.StatusInternalServerError,
		)

		return
	}

	time.Sleep(500 * time.Millisecond)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getMemoField[T any](fields map[string]*common.Payload, key string, out *T) error {
	if f, ok := fields[key]; ok && f != nil {
		if data := f.GetData(); data != nil {
			return json.Unmarshal(data, out)
		}
	}
	return nil
}

// Template for dashboard
var page = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html data-theme="dark">
  <head>
    <title>🍕 Pizza Delivery Dashboard 🍕</title>
    <link
      rel="stylesheet"
      href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.orange.min.css"
    >
    <style>
      :root { --pico-font-size: 1.8rem; }
			header {
			  position: sticky;
			  top: 0;
			  background: var(--pico-background-color);
			  padding: 0.5rem 0;
				z-index: 10;
			}
			header h4 { padding: 0.5rem 0; }
			.container {
  			max-width: 950px;
  			margin: 0 auto;
  			padding: 0 0.5rem;
			}
			.order-card {
			  padding: 1rem;
			  border: 1px solid var(--pico-muted-border-color);
			  border-radius: 0.75rem;
			  box-shadow: 0 2px 5px rgba(0,0,0,0.05);
			  margin-bottom: 1rem;
			}
			.order-card:hover { box-shadow: 0 4px 8px rgba(0,0,0,0.1); }
			.order-meta {
			}
			.order-meta ul { padding: 0; }
			.order-meta li { list-style: none; }
			.order-meta .debug {
				font-size: 0.5rem; 
				float: right; 
				color: var(--pico-muted-color); 
			}
    </style>
  </head>
  <body>
    <header class="container">
      <h4>🚚 Orders waiting for driver ({{len .Orders}})</h4>
    </header>
    <main class="container">
      {{range .Orders}}
        <article class="order-card">
          <div class="order-header">
            <h2>
							<u>{{.Info.Pizza.Name}}</u> to <em>{{.Info.Customer.Name}}</em>
						</h2>
          </div>
					<div class="order-meta">
					  <ul>
					    <li>🍕 {{.Info.Pizza.Name}}</li>
					    <li>🏠 <strong>{{.Info.Customer.Name}}</strong>
					      <ul>
					        <li>
										<address>{{.Info.Customer.Addr}}</address>
									</li>
					      </ul>
					    </li>
					  </ul>
					</div>
          <form method="POST" action="/assign">
            <input type="hidden" name="workflowID" value="{{.WorkflowID}}">
            <input type="hidden" name="runID" value="{{.RunID}}">
            <fieldset role="group">
              <select name="driver">
                {{range $.Drivers}}
                  <option value="{{.}}">{{.}}</option>
                {{end}}
              </select>
              <input type="submit" value="Assign 🚗">
            </fieldset>
          </form>
        </article>
      {{else}}
        <p><em>No orders waiting for a driver ✨</em></p>
      {{end}}
    </main>
  </body>
</html>
`))
