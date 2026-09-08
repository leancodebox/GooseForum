package job

import (
	"context"
	"log/slog"

	"github.com/leancodebox/GooseForum/app/service/oidcproviderservice"
)

// cleanupOIDC removes expired protocol data so login and token rotation do not
// grow the database indefinitely. Expiry is enforced by request handlers even
// between cleanup runs; replay-detection records retain their required lifetime.
func cleanupOIDC() {
	if !oidcproviderservice.Configured() {
		return
	}
	service, err := oidcproviderservice.Default()
	if err != nil {
		slog.Error("OIDC cleanup unavailable", "err", err)
		return
	}
	if _, err := service.PurgeDefault(context.Background()); err != nil {
		slog.Error("OIDC cleanup failed", "err", err)
	}
}
