package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rustdesk/rustdesk-api-server"
	"github.com/rustdesk/rustdesk-api-server/config"
	"github.com/rustdesk/rustdesk-api-server/internal/database"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"github.com/rustdesk/rustdesk-api-server/internal/server"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "rustdesk-api-server",
		Short: "RustDesk API Server — Web admin console backend",
		Long: `RustDesk API Server provides REST API and WebSocket services for the
RustDesk Web Admin Console. It handles device heartbeat, sysinfo,
address book sync, audit logging, and user management.

Commands:
  serve    Start the API server
  sync     Initialize and migrate the database
  user     Manage user accounts (user add, ...)`,
	}

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "path to config file (default: config/config.yaml)")

	rootCmd.AddCommand(serveCmd())
	rootCmd.AddCommand(syncCmd())
	rootCmd.AddCommand(userCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func serveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Setup logging
			level, err := logrus.ParseLevel(cfg.Log.Level)
			if err != nil {
				level = logrus.InfoLevel
			}
			logrus.SetLevel(level)
			logrus.SetFormatter(&logrus.TextFormatter{
				FullTimestamp: true,
			})

			logrus.Infof("Starting RustDesk API Server v0.1.0")
			logrus.Infof("Config loaded from: %s", cfgFile)

			// Initialize database
			if err := database.Init(cfg.Database); err != nil {
				return fmt.Errorf("failed to init database: %w", err)
			}
			logrus.Info("Database initialized")

			if err := database.Migrate(); err != nil {
				return fmt.Errorf("failed to migrate database: %w", err)
			}
			logrus.Info("Database migration completed")

			// Seed initial admin user if users table is empty
			seedAdminUser()

			// Create server and start with timeouts, TLS, and graceful shutdown
			srv := server.NewServer(cfg, embedded.Frontend)

			httpServer := &http.Server{
				Addr:              cfg.Server.Addr,
				Handler:           srv,
				ReadTimeout:       10 * time.Second,
				WriteTimeout:      10 * time.Second,
				IdleTimeout:       60 * time.Second,
				ReadHeaderTimeout: 5 * time.Second,
			}

			go func() {
				var err error
				if cfg.Server.CertFile != "" && cfg.Server.KeyFile != "" {
					logrus.Infof("Starting HTTPS server on %s", cfg.Server.Addr)
					err = httpServer.ListenAndServeTLS(cfg.Server.CertFile, cfg.Server.KeyFile)
				} else {
					logrus.Warnf("Starting HTTP server on %s (TLS not configured)", cfg.Server.Addr)
					err = httpServer.ListenAndServe()
				}
				if err != nil && err != http.ErrServerClosed {
					logrus.Fatalf("server error: %v", err)
				}
			}()

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit
			logrus.Info("Shutting down server...")

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := httpServer.Shutdown(ctx); err != nil {
				logrus.Fatalf("server forced to shutdown: %v", err)
			}
			logrus.Info("Server exited")
			return nil
		},
	}
}

func syncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Initialize and migrate the database (no server start)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			logrus.SetLevel(logrus.InfoLevel)
			logrus.SetFormatter(&logrus.TextFormatter{
				FullTimestamp: true,
			})

			if err := database.Init(cfg.Database); err != nil {
				return fmt.Errorf("failed to init database: %w", err)
			}

			if err := database.Migrate(); err != nil {
				return fmt.Errorf("failed to migrate database: %w", err)
			}

			logrus.Info("Database sync completed successfully")
			fmt.Println("Database sync completed — all tables created/updated.")
			return nil
		},
	}
}

func userCmd() *cobra.Command {
	var isAdmin bool

	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage user accounts",
	}

	addCmd := &cobra.Command{
		Use:   "add <username> <password>",
		Short: "Add a new user",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			username := args[0]
			password := args[1]

			cfg, err := config.Load(cfgFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			logrus.SetLevel(logrus.WarnLevel)
			if err := database.Init(cfg.Database); err != nil {
				return fmt.Errorf("failed to init database: %w", err)
			}

			if err := database.Migrate(); err != nil {
				return fmt.Errorf("failed to migrate database: %w", err)
			}

			// Create user via service layer
			user, err := service.CreateUser(username, password, isAdmin)
			if err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}

			fmt.Printf("User created: %s (id=%d, admin=%v)\n", user.Username, user.ID, user.IsAdmin)
			return nil
		},
	}

	addCmd.Flags().BoolVar(&isAdmin, "admin", false, "Grant admin privileges")
	cmd.AddCommand(addCmd)
	return cmd
}

// seedAdminUser checks if the users table is empty and auto-creates an admin
// account with a random password. Credentials are printed to stdout.
func seedAdminUser() {
	var count int64
	database.DB.Model(&model.User{}).Count(&count)
	if count > 0 {
		return
	}

	password := service.GenerateRandomPassword()
	user, err := service.CreateUser("admin", password, true)
	if err != nil {
		logrus.Errorf("Failed to seed admin user: %v", err)
		return
	}

	// Print to stderr directly, not through the logging system
	fmt.Fprintf(os.Stderr, "\n=== ADMIN ACCOUNT CREATED ===\n  Username: %s\n  Password: %s\n=== CHANGE THIS PASSWORD IMMEDIATELY ===\n\n", user.Username, password)
}
