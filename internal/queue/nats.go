package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	natsStream  = "TF_AGENT"
	natsAckWait = 5 * time.Minute
)

func natsSubject(name string) string { return "tf.tasks." + name }
func natsDurable(name string) string { return "tf-workers-" + name }

// NATSQueue is a durable NATS JetStream-backed Queue.
// Tasks survive server restarts and can be consumed by multiple workers.
type NATSQueue struct {
	nc   *nats.Conn
	js   nats.JetStreamContext
	sub  *nats.Subscription
	name string // queue name, used as subject suffix
}

// NewNATSQueue connects to NATS, ensures the shared stream exists, creates a
// per-name durable pull consumer, and returns a queue ready to push and pop.
// name is used as the subject suffix (e.g. "default" → tf.tasks.default).
func NewNATSQueue(url, name string) (*NATSQueue, error) {
	if name == "" {
		name = "default"
	}
	nc, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("nats jetstream: %w", err)
	}

	subject := natsSubject(name)
	durable := natsDurable(name)

	// Create (or no-op) the shared stream covering all named queues (tf.tasks.>).
	_, err = js.AddStream(&nats.StreamConfig{
		Name:      natsStream,
		Subjects:  []string{"tf.tasks.>"},
		Storage:   nats.FileStorage,
		Retention: nats.WorkQueuePolicy, // each message delivered to exactly one consumer
		Replicas:  1,
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		// Stream exists with a different config — attempt update.
		_, err = js.UpdateStream(&nats.StreamConfig{
			Name:      natsStream,
			Subjects:  []string{"tf.tasks.>"},
			Storage:   nats.FileStorage,
			Retention: nats.WorkQueuePolicy,
			Replicas:  1,
		})
		if err != nil {
			nc.Close()
			return nil, fmt.Errorf("nats stream: %w", err)
		}
	}

	// Per-name pull consumer — durable so it survives reconnects.
	sub, err := js.PullSubscribe(subject, durable,
		nats.BindStream(natsStream),
		nats.AckWait(natsAckWait),
		nats.MaxDeliver(5),
	)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("nats pull subscribe (%s): %w", name, err)
	}

	return &NATSQueue{nc: nc, js: js, sub: sub, name: name}, nil
}

func (q *NATSQueue) Push(ctx context.Context, item Item) error {
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal item: %w", err)
	}
	_, err = q.js.Publish(natsSubject(q.name), data)
	return err
}

// Name returns the queue name.
func (q *NATSQueue) Name() string { return q.name }

// Pop blocks until a task is available or ctx is cancelled.
func (q *NATSQueue) Pop(ctx context.Context) (Item, error) {
	for {
		select {
		case <-ctx.Done():
			return Item{}, ctx.Err()
		default:
		}

		msgs, err := q.sub.Fetch(1, nats.MaxWait(500*time.Millisecond))
		if err == nats.ErrTimeout {
			continue
		}
		if err != nil {
			// On connection errors, back off and retry.
			select {
			case <-ctx.Done():
				return Item{}, ctx.Err()
			case <-time.After(time.Second):
			}
			continue
		}
		if len(msgs) == 0 {
			continue
		}

		msg := msgs[0]
		var item Item
		if err := json.Unmarshal(msg.Data, &item); err != nil {
			_ = msg.Nak()
			continue
		}
		_ = msg.Ack()
		return item, nil
	}
}

// Len returns the approximate number of pending messages in the stream.
func (q *NATSQueue) Len() int {
	info, err := q.js.StreamInfo(natsStream)
	if err != nil {
		return 0
	}
	return int(info.State.Msgs)
}

// Close drains and closes the NATS connection.
func (q *NATSQueue) Close() error {
	_ = q.sub.Unsubscribe()
	q.nc.Close()
	return nil
}
