# 🍕 Pizza Delivery Dashboard

`pizza-delivery-dashboard` is a lightweight Go web server that provides 
a simple dashboard for managing pizza delivery workflows using [Temporal](https://temporal.io/). 

It allows you to view workflows waiting for a driver and assign 
available drivers directly from the web interface.

## Features

- Displays a list of workflows where a driver has not yet been assigned.
- Allows assignment of drivers to workflows via a simple HTML dashboard.
- Integrates with Temporal's workflow service for querying and signaling workflows.
- Lightweight, minimal UI using [PicoCSS](https://picocss.com/).

## Getting Started

### Prerequisites

- Go 1.21+  
- Temporal server running (default namespace: `default`)  

### Usage

Once the server is running, open your browser: <http://localhost:8234>

You will see a list of workflows waiting for drivers. 
Each workflow can be assigned a driver from a dropdown menu. 
After assignment, the workflow is signaled, and the dashboard reloads.

### Configurable Drivers

The server comes with a default set of drivers:

```go
var drivers = []string{
    "🧒",
    "👴",
    "👲",
}
```

You can modify this list in `main.go` to match your drivers.

### Endpoints

| Endpoint       | Method | Description |
|----------------|--------|-------------|
| `/`            | GET    | Lists workflows waiting for driver assignment. |
| `/assign`      | POST   | Assigns a driver to a workflow and signals Temporal. |

## How It Works

1. The server queries Temporal for workflows where `DriverAssigned = false`.
2. The dashboard displays the list of workflows and available drivers.
3. When a driver is assigned, the server signals the workflow using Temporal's `SignalWorkflow` API.
4. The dashboard reloads to reflect updated workflow statuses.

## Dependencies

- [Go Temporal SDK](https://pkg.go.dev/go.temporal.io/sdk)
- [PicoCSS](https://picocss.com/) for styling
- Go standard libraries: `net/http`, `html/template`, `context`, `log/slog`, `os`, `time`
