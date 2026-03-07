package orchestrate

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RunsCreated tracks number of runs created
	RunsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "agentruntime_runs_created_total",
		Help: "Total number of runs created",
	})

	// RunsActive tracks number of currently active runs
	RunsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "agentruntime_runs_active",
		Help: "Number of currently active runs",
	})

	// RunDuration tracks run completion time
	RunDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "agentruntime_run_duration_seconds",
		Help:    "Duration of run execution in seconds",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
	})

	// APIRequestDuration tracks API request duration
	APIRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "agentruntime_api_request_duration_seconds",
		Help:    "API request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "endpoint", "status"})

	// ToolExecutions tracks tool execution counts
	ToolExecutions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "agentruntime_tool_executions_total",
		Help: "Total number of tool executions",
	}, []string{"tool", "status"})

	// PolicyDenials tracks policy denials
	PolicyDenials = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "agentruntime_policy_denials_total",
		Help: "Total number of policy denials",
	}, []string{"rule"})

	// RunRetries tracks retry attempts for failed runs.
	RunRetries = promauto.NewCounter(prometheus.CounterOpts{
		Name: "agentruntime_run_retries_total",
		Help: "Total number of run retry attempts",
	})

	// RunDeadLetters tracks terminal failures written to dead-letter storage.
	RunDeadLetters = promauto.NewCounter(prometheus.CounterOpts{
		Name: "agentruntime_run_dead_letters_total",
		Help: "Total number of dead-lettered runs",
	})

	// RunQueueDepth tracks runner queue depth.
	RunQueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "agentruntime_run_queue_depth",
		Help: "Current number of queued run requests",
	})

	// RunQueueRejected tracks run submissions rejected due to full queue.
	RunQueueRejected = promauto.NewCounter(prometheus.CounterOpts{
		Name: "agentruntime_run_queue_rejected_total",
		Help: "Total run submissions rejected because queue was full",
	})

	// RunFailures tracks terminal run failures grouped by reason/source.
	RunFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "agentruntime_run_failures_total",
		Help: "Total number of terminal run failures by reason and source",
	}, []string{"reason", "source"})

	// DeadLettersPruned tracks dead letters removed by retention pruner.
	DeadLettersPruned = promauto.NewCounter(prometheus.CounterOpts{
		Name: "agentruntime_dead_letters_pruned_total",
		Help: "Total number of dead letters pruned by retention",
	})

	// DeadLettersArchived tracks dead letters archived before prune.
	DeadLettersArchived = promauto.NewCounter(prometheus.CounterOpts{
		Name: "agentruntime_dead_letters_archived_total",
		Help: "Total number of dead letters archived to disk",
	})
)
