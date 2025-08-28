package main

import (
	"io"
	"log/slog"
	"os"

	pds "pizza-delivery-simulator"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"
)

const taskQueue = "pizza"

func main() {
	logger := newLogger(os.Stdout)

	c, err := client.Dial(newOptions(logger))
	if err != nil {
		logger.Error("Unable to create client", slog.Any("error", err))
		os.Exit(1)
	}
	defer c.Close()

	w := newWorker(c, worker.Options{})

	if err := w.Run(worker.InterruptCh()); err != nil {
		logger.Error("Unable to start worker", slog.Any("error", err))
		os.Exit(1)
	}
}

func newWorker(client client.Client, options worker.Options) worker.Worker {
	w := worker.New(client, taskQueue, options)

	register(w, pds.GetMenu, pds.PlaceOrder)

	w.RegisterActivity(&pds.Activities{})

	return w
}

func register(w worker.Worker, workflows ...any) {
	for _, workflow := range workflows {
		w.RegisterWorkflow(workflow)
	}
}

func newLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, nil))
}

func newOptions(logger *slog.Logger) client.Options {
	return client.Options{
		Logger: log.NewStructuredLogger(logger),
	}
}
