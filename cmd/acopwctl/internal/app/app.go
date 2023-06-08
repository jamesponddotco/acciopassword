// Package app is where the core logic for acopwctl lives.
package app

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"git.sr.ht/~jamesponddotco/acciopassword/internal/build"
	"git.sr.ht/~jamesponddotco/acciopassword/internal/config"
	"git.sr.ht/~jamesponddotco/acciopassword/internal/database"
	"git.sr.ht/~jamesponddotco/acciopassword/internal/server"
	"git.sr.ht/~jamesponddotco/xstd-go/xlog"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func Run() int {
	logger, err := zap.NewProduction()
	if err != nil && !errors.Is(err, syscall.ENOTTY) {
		xlog.Printf("Failed to create logger: %v\n", err)

		return 1
	}

	rootCmd := &cobra.Command{
		Use:               "acopwctl",
		Short:             "acopwctl is a CLI tool for controlling the Accio Password server.",
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
		Version:           Version(),
	}

	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)

	AddCommands(rootCmd, logger)

	if err := rootCmd.Execute(); err != nil {
		xlog.Printf("error: %s", err.Error())

		return 1
	}

	return 0
}

func AddCommands(rootCmd *cobra.Command, logger *zap.Logger) {
	addStartCommand(rootCmd, logger)
	addStopCommand(rootCmd, logger)
}

func addStartCommand(rootCmd *cobra.Command, logger *zap.Logger) {
	var configPath string

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the server.",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				logger.Error("Failed to load config", zap.Error(err))

				return
			}

			db, err := database.Open(logger, cfg.Database.DSN)
			if err != nil {
				logger.Error("Failed to open database", zap.Error(err))

				return
			}
			defer db.Close()

			srv, err := server.New(cfg, db, logger)
			if err != nil {
				logger.Error("Failed to create server", zap.Error(err))

				return
			}

			if _, err = os.Stat(cfg.Server.PID); !os.IsNotExist(err) {
				logger.Error("Server is already running")

				return
			}

			pid := os.Getpid()

			pidFile, err := os.Create(cfg.Server.PID)
			if err != nil {
				logger.Error("Failed to create PID file", zap.Error(err))

				return
			}
			defer pidFile.Close()

			_, err = fmt.Fprintf(pidFile, "%d\n", pid)
			if err != nil {
				logger.Error("Failed to write PID file", zap.Error(err))

				return
			}

			if err := srv.Start(); err != nil {
				logger.Error("Failed to run server", zap.Error(err))

				return
			}
		},
	}

	startCmd.Flags().StringVarP(&configPath, "config", "c", "config.json", "Path to the configuration file.")

	rootCmd.AddCommand(startCmd)
}

func addStopCommand(rootCmd *cobra.Command, logger *zap.Logger) {
	var cfgPath string

	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the server.",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadConfig(cfgPath)
			if err != nil {
				logger.Error("Failed to load config", zap.Error(err))

				return
			}

			pidFileData, err := os.ReadFile(cfg.Server.PID)
			if err != nil {
				logger.Error("Failed to read PID file", zap.Error(err))

				return
			}

			pid, err := strconv.Atoi(strings.TrimSpace(string(pidFileData)))
			if err != nil {
				logger.Error("Failed to parse PID file", zap.Error(err))

				return
			}

			process, err := os.FindProcess(pid)
			if err != nil {
				logger.Error("Failed to find process", zap.Error(err))

				return
			}

			if err = process.Signal(os.Interrupt); err != nil {
				logger.Error("Failed to send interrupt signal", zap.Error(err))

				return
			}

			if err = os.Remove(cfg.Server.PID); err != nil {
				logger.Error("Failed to remove PID file", zap.Error(err))

				return
			}
		},
	}

	stopCmd.Flags().StringVarP(&cfgPath, "config", "c", "config.json", "Path to the configuration file.")

	rootCmd.AddCommand(stopCmd)
}

func Version() string {
	var builder strings.Builder

	builder.WriteString("Accio Password: ")
	builder.WriteString(build.Version)

	return builder.String()
}
