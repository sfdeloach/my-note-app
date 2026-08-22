// Command agenda-service serves the note listing and the four rendered
// views (Agenda, Red-Letter Agenda, Minutes, Action Items) for Session
// Meeting notes stored in the shared Joplin Postgres database. See
// docs/agenda-service/initial-prompt.md for the full brief and
// docs/agenda-service/roadmap.md for how it was built.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sfdeloach/my-note-app/agenda-service/reader"
	"github.com/sfdeloach/my-note-app/agenda-service/server"
	"github.com/sfdeloach/my-note-app/agenda-service/settings"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	pool, err := pgxpool.New(ctx, cfg.dsn())
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()

	rdr, err := reader.New(ctx, pool)
	if err != nil {
		return fmt.Errorf("initializing reader: %w", err)
	}

	// Fail fast on a broken settings file at startup. Every request after
	// this re-reads and re-parses it fresh — settings.Load never caches.
	if _, err := settings.Load(cfg.settingsPath); err != nil {
		return fmt.Errorf("loading settings: %w", err)
	}

	srv := &http.Server{
		Addr:              cfg.listenAddr,
		Handler:           server.New(rdr, cfg.settingsPath),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("agenda-service listening on %s", cfg.listenAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("serving: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

type config struct {
	dbHost       string
	dbPort       string
	dbName       string
	dbUser       string
	dbPassword   string
	settingsPath string
	listenAddr   string
}

// loadConfig reads every environment variable this service needs and
// fails fast, naming every missing one at once, rather than dying on the
// first one a caller happens to have forgotten.
func loadConfig() (config, error) {
	c := config{
		dbHost:       os.Getenv("AGENDA_DB_HOST"),
		dbPort:       os.Getenv("AGENDA_DB_PORT"),
		dbName:       os.Getenv("AGENDA_DB_NAME"),
		dbUser:       os.Getenv("AGENDA_DB_USER"),
		dbPassword:   os.Getenv("AGENDA_DB_PASSWORD"),
		settingsPath: os.Getenv("AGENDA_SETTINGS_PATH"),
		listenAddr:   os.Getenv("AGENDA_LISTEN_ADDR"),
	}

	named := map[string]string{
		"AGENDA_DB_HOST":       c.dbHost,
		"AGENDA_DB_PORT":       c.dbPort,
		"AGENDA_DB_NAME":       c.dbName,
		"AGENDA_DB_USER":       c.dbUser,
		"AGENDA_DB_PASSWORD":   c.dbPassword,
		"AGENDA_SETTINGS_PATH": c.settingsPath,
		"AGENDA_LISTEN_ADDR":   c.listenAddr,
	}
	var missing []string
	for name, val := range named {
		if val == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return config{}, fmt.Errorf("missing required environment variable(s): %v", missing)
	}

	return c, nil
}

func (c config) dsn() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.dbUser, c.dbPassword, c.dbHost, c.dbPort, c.dbName)
}
