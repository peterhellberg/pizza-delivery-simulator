package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"time"

	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
)

const (
	port                  = "8234"
	namespace             = "default"
	queryUnassignedOrders = `WorkflowType="PlaceOrder" AND ExecutionStatus="Running" AND DriverAssigned=false`
)

// Available drivers
var drivers = []Driver{
	{ID: "tommy", Emoji: "🧒", Name: "Tommy Brown"},
	{ID: "walter", Emoji: "👴", Name: "Walter Smith"},
	{ID: "james", Emoji: "🧔", Name: "James O'Connor"},
}

type Driver struct {
	ID    string
	Emoji string
	Name  string
}

type DriverNote struct {
	Driver
	Note string
}

// Order represents a workflow execution of type PlaceOrder without an assigned driver
type Order struct {
	WorkflowID string
	RunID      string
	Info       OrderInfo
}

// OrderInfo is the information about an order
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

// IndexData contains the data used for the index template
type IndexData struct {
	OrdersCount int64
}

// OrdersData contains the data used for the orders template
type OrdersData struct {
	Orders  []Order
	Drivers []Driver
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	jsFS, err := fs.Sub(staticFS, "static/js")
	if err != nil {
		logger.Info("Unable construct js filesystem", "Error", err)
		os.Exit(1)
	}

	cssFS, err := fs.Sub(staticFS, "static/css")
	if err != nil {
		logger.Info("Unable construct css filesystem", "Error", err)
		os.Exit(1)
	}

	temporal, err := client.NewLazyClient(client.Options{
		Logger: log.NewStructuredLogger(logger),
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

	http.HandleFunc("GET /{$}", d.index)
	http.HandleFunc("POST /assign", d.assign)
	http.HandleFunc("GET /style.css", d.styleCSS)
	http.HandleFunc("GET /orders", d.orders)
	http.HandleFunc("GET /orders/count", d.ordersCount)
	http.HandleFunc("GET /orders/count.stream", d.ordersCountStream)
	http.Handle("GET /js/", http.StripPrefix("/js/", http.FileServer(http.FS(jsFS))))
	http.Handle("GET /css/", http.StripPrefix("/css/", http.FileServer(http.FS(cssFS))))

	logger.Info("Pizza dashboard started at http://localhost:" + port)
	http.ListenAndServe(":"+port, nil)
}

// The main type for the dashboard, such as the endpoints
type Dashboard struct {
	client.Client
	log.Logger
}

func (d *Dashboard) listUnassignedOrders(ctx context.Context, pageSize int32) ([]Order, error) {
	resp, err := d.WorkflowService().ListWorkflowExecutions(ctx,
		&workflowservice.ListWorkflowExecutionsRequest{
			Namespace: namespace,
			PageSize:  pageSize,
			Query:     queryUnassignedOrders,
		},
	)
	if err != nil {
		return nil, err
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

	return orders, nil
}

// index handler queries workflows and renders the template
func (d *Dashboard) index(w http.ResponseWriter, r *http.Request) {
	count, err := d.countUnassignedOrders(r.Context())
	if err != nil {
		http.Error(w, "Failed to count unassigned orders: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := indexTemplate.Execute(w, IndexData{
		OrdersCount: count,
	}); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// assign signals a workflow that a driver has been assigned and reloads dashboard
func (d *Dashboard) assign(w http.ResponseWriter, r *http.Request) {
	var (
		workflowID = r.FormValue("workflowID")
		runID      = r.FormValue("runID")
		driver     = r.FormValue("driver")
		note       = r.FormValue("note")
		ctx        = context.Background()
	)

	idx := slices.IndexFunc(drivers, func(d Driver) bool {
		return d.Name == driver
	})

	if err := d.SignalWorkflow(ctx, workflowID, runID, "DriverAssigned", DriverNote{
		Driver: drivers[idx],
		Note:   note,
	}); err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to signal workflow: %v", err),
			http.StatusInternalServerError,
		)

		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (d *Dashboard) styleCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	w.Write(styleCSS)
}

func (d *Dashboard) orders(w http.ResponseWriter, r *http.Request) {
	orders, err := d.listUnassignedOrders(r.Context(), 25)
	if err != nil {
		http.Error(w, "Failed to query workflows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := ordersTemplate.Execute(w, OrdersData{
		Orders:  orders,
		Drivers: drivers,
	}); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (d *Dashboard) countUnassignedOrders(ctx context.Context) (int64, error) {
	resp, err := d.WorkflowService().CountWorkflowExecutions(ctx,
		&workflowservice.CountWorkflowExecutionsRequest{
			Namespace: namespace,
			Query:     queryUnassignedOrders,
		},
	)

	if err != nil {
		return 0, err
	}

	return resp.Count, nil
}

// Endpoint method for /orders/count
func (d *Dashboard) ordersCount(w http.ResponseWriter, r *http.Request) {
	count, err := d.countUnassignedOrders(r.Context())
	if err != nil {
		http.Error(w, "Failed to count unassigned orders: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%d", count)
}

// Endpoint method for /orders/count.stream
func (d *Dashboard) ordersCountStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Check if w is a HTTP flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send a first "keepalive" comment so Firefox commits the stream
	fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastCount int64 = -1

	checkCount := func(ctx context.Context) {
		count, err := d.countUnassignedOrders(ctx)
		if err != nil {
			return
		}

		if count != lastCount {
			fmt.Fprintf(w, "event: count-changed\ndata: %d\n\n", count)
			flusher.Flush()
			lastCount = count
		}
	}

	checkCount(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			checkCount(ctx)
		}
	}
}

func getMemoField[T any](fields map[string]*common.Payload, key string, out *T) error {
	if f, ok := fields[key]; ok && f != nil {
		if data := f.GetData(); data != nil {
			return json.Unmarshal(data, out)
		}
	}
	return nil
}

var (
	//go:embed style.css
	styleCSS []byte

	//go:embed static/*
	staticFS embed.FS

	//go:embed templates/index.html
	indexHTML string

	//go:embed templates/orders.html
	ordersHTML string

	// Parsed templates
	indexTemplate  = template.Must(template.New("index").Parse(indexHTML))
	ordersTemplate = template.Must(template.New("orders").Parse(ordersHTML))
)
