package e2e

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/bot/basic"
)

// mcConnect attempts to connect to a Minecraft server using the Minecraft protocol.
// This triggers lazymc to start the backend server.
// It uses offline mode authentication (no Microsoft account needed).
func mcConnect(host string, port int, playerName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	addr := net.JoinHostPort(host, strconv.Itoa(port))

	// Create a new bot client
	client := bot.NewClient()
	client.Auth.Name = playerName

	// Try to dial the server using a dialer with context
	//nolint:exhaustruct // Only Timeout is needed
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}
	conn, dialErr := dialer.DialContext(ctx, "tcp", addr)
	if dialErr != nil {
		return dialErr
	}

	// Set deadline for the connection
	if deadlineErr := conn.SetDeadline(time.Now().Add(10 * time.Second)); deadlineErr != nil {
		_ = conn.Close()
		return deadlineErr
	}

	defer func() {
		_ = conn.Close()
	}()

	// Join the server (this sends the login request that triggers lazymc)
	//nolint:exhaustruct // Only Context is needed for our use case
	joinErr := client.JoinServerWithOptions(addr, bot.JoinOptions{
		Context: ctx,
	})
	if joinErr != nil {
		// Connection might be refused or timeout, but the important thing is
		// that we sent the login request which triggers lazymc
		// Even if it fails, the handshake should trigger the server start
		return nil // Don't return error as the purpose is just to trigger
	}

	// Register basic handlers to prevent crashes
	//nolint:exhaustruct // Only required fields for triggering server
	basic.NewPlayer(client, basic.DefaultSettings, basic.EventsListener{})

	return nil
}

// mcTriggerServerStart attempts to trigger lazymc to start the Minecraft server
// by sending a Minecraft protocol handshake and login request.
func mcTriggerServerStart(host string, port int) error {
	return mcConnect(host, port, "E2ETestPlayer")
}
