// Package listener bridges Postgres LISTEN/NOTIFY into the Hub.
// It holds one dedicated connection: WaitForNotification occupies it,
// so pooled handles must never be used here.
package listener

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"go-lab/storage/realtime/internal/bus"
)

// tenantProbe extracts only the routing key from a notification payload.
type tenantProbe struct {
	TenantID int `json:"tenant_id"`
}

// Start connects, LISTENs on channel, and publishes every notification
// to the Hub until ctx ends. It returns only on error or cancel.
func Start(ctx context.Context, dsn, channel string, hub *bus.Hub) error {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return err
	}

	defer func() {
		_ = conn.Close(ctx)
	}()

	_, err = conn.Exec(ctx, "LISTEN "+channel)
	if err != nil {
		return err
	}

	slog.Info("listening for notifications", "channel", channel)

	for {
		notification, err := conn.WaitForNotification(ctx)
		if err != nil {
			return err
		}

		var probe tenantProbe

		err = json.Unmarshal([]byte(notification.Payload), &probe)
		if err != nil {
			slog.Warn("dropping unparsable notification", "error", err)

			continue
		}

		hub.Publish(probe.TenantID, []byte(notification.Payload))
	}
}
