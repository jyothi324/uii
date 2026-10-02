package postgreslistener

import (
	"context"
	"log"
	"time"

	"github.com/lib/pq"
)

type EventHandler func(message []byte)

type Listener struct {
	connString string
	channel    string
	handler    EventHandler
	logger     *log.Logger
	listener   *pq.Listener
}

func NewListener(connString, channel string, handler EventHandler, logger *log.Logger) *Listener {
	return &Listener{
		connString: connString,
		channel:    channel,
		handler:    handler,
		logger:     logger,
	}
}

func (l *Listener) Start(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	listener := pq.NewListener(l.connString, 10*time.Second, 90*time.Second, l.reportProblem)
	if err := listener.Listen(l.channel); err != nil {
		return err
	}

	l.listener = listener
	l.logger.Printf("Postgres listener started on channel %s", l.channel)

	for {
		select {
		case n := <-listener.NotificationChannel():
			if n == nil {
				continue
			}
			l.logger.Printf("Received notification on %s: %s", l.channel, n.Extra)
			l.handleNotification(n)
		case <-time.After(30 * time.Second):
			if err := listener.Ping(); err != nil {
				l.logger.Printf("Postgres listener ping failed: %v", err)
			}
		case <-ctx.Done():
			listener.UnlistenAll()
			listener.Close()
			l.logger.Println("Postgres listener shutdown")
			return nil
		}
	}
}

func (l *Listener) handleNotification(notification *pq.Notification) {
	if notification == nil || notification.Extra == "" {
		return
	}

	payload := []byte(notification.Extra)
	l.handler(payload)
}

func (l *Listener) reportProblem(event pq.ListenerEventType, err error) {
	if err != nil {
		l.logger.Printf("Postgres listener problem (event=%v): %v", event, err)
	}
}

func (l *Listener) Close() error {
	if l.listener == nil {
		return nil
	}
	l.listener.UnlistenAll()
	return l.listener.Close()
}
