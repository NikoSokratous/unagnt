package orchestrate

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisQueueKey = "agentruntime:queue"
const redisDequeueTimeout = 30 * time.Second

// redisQueueItem is the serializable representation of RunRequest for Redis.
type redisQueueItem struct {
	RunID        string                 `json:"run_id"`
	AgentName    string                 `json:"agent_name"`
	Goal         string                 `json:"goal"`
	Source       string                 `json:"source"`
	Outputs      map[string]interface{} `json:"outputs,omitempty"`
	MaxRetries   int                    `json:"max_retries"`
	RetryBackoff int64                  `json:"retry_backoff_ns"`
	Timeout      int64                  `json:"timeout_ns"`
}

// NewRedisQueue creates a Redis-backed durable queue.
func NewRedisQueue(client *redis.Client, queueKey string) QueueBackend {
	if queueKey == "" {
		queueKey = redisQueueKey
	}
	return &redisQueue{
		client:   client,
		queueKey: queueKey,
	}
}

type redisQueue struct {
	client   *redis.Client
	queueKey string
	once     sync.Once
}

func (r *redisQueue) Enqueue(ctx context.Context, req RunRequest) error {
	item := redisQueueItem{
		RunID:        req.RunID,
		AgentName:    req.AgentName,
		Goal:         req.Goal,
		Source:       req.Source,
		Outputs:      req.Outputs,
		MaxRetries:   req.MaxRetries,
		RetryBackoff: int64(req.RetryBackoff),
		Timeout:      int64(req.Timeout),
	}
	b, err := json.Marshal(item)
	if err != nil {
		return err
	}
	added := r.client.LPush(ctx, r.queueKey, b)
	if err := added.Err(); err != nil {
		RunQueueRejected.Inc()
		return err
	}
	RunQueueDepth.Set(float64(r.client.LLen(ctx, r.queueKey).Val()))
	return nil
}

func (r *redisQueue) Dequeue(ctx context.Context) (RunRequest, bool, error) {
	// BRPOP blocks until an item is available or context is done
	result, err := r.client.BRPop(ctx, redisDequeueTimeout, r.queueKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return RunRequest{}, false, err
	}
	if len(result) < 2 {
		return RunRequest{}, false, nil
	}
	raw := result[1]
	var item redisQueueItem
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		return RunRequest{}, false, err
	}
	req := RunRequest{
		RunID:        item.RunID,
		AgentName:    item.AgentName,
		Goal:         item.Goal,
		Source:       item.Source,
		Outputs:      item.Outputs,
		MaxRetries:   item.MaxRetries,
		RetryBackoff: time.Duration(item.RetryBackoff),
		Timeout:      time.Duration(item.Timeout),
	}
	RunQueueDepth.Set(float64(r.client.LLen(ctx, r.queueKey).Val()))
	return req, true, nil
}

func (r *redisQueue) Len() int {
	n, err := r.client.LLen(context.Background(), r.queueKey).Result()
	if err != nil {
		return 0
	}
	return int(n)
}

func (r *redisQueue) Close() error {
	r.once.Do(func() {})
	return nil
}
