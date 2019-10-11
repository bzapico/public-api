/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package commands

import (
	"github.com/nalej/public-api/internal/pkg/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var config = server.Config{}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Launch the server API",
	Long:  `Launch the server API`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		log.Info().Msg("Launching API!")
		config.Debug = debugLevel
		server := server.NewService(config)
		server.Run()
	},
}

func init() {
	runCmd.Flags().IntVar(&config.Port, "port", 8081, "Port to launch the Public gRPC API")
	runCmd.Flags().IntVar(&config.HTTPPort, "httpPort", 8082, "Port to launch the Public HTTP API")
	runCmd.PersistentFlags().StringVar(&config.SystemModelAddress, "systemModelAddress", "localhost:8800",
		"System Model address (host:port)")
	runCmd.PersistentFlags().StringVar(&config.InfrastructureManagerAddress, "infrastructureManagerAddress", "localhost:8860",
		"Infrastructure Manager address (host:port)")
	runCmd.PersistentFlags().StringVar(&config.ApplicationsManagerAddress, "applicationsManagerAddress", "localhost:8910",
		"Applications Manager address (host:port)")
	runCmd.PersistentFlags().StringVar(&config.UserManagerAddress, "userManagerAddress", "localhost:8920",
		"User Manager address (host:port)")
	runCmd.PersistentFlags().StringVar(&config.AuthHeader, "authHeader", "", "Authorization Header")
	runCmd.PersistentFlags().StringVar(&config.AuthSecret, "authSecret", "", "Authorization secret")
	runCmd.PersistentFlags().StringVar(&config.AuthConfigPath, "authConfigPath", "", "Authorization config path")
	runCmd.PersistentFlags().StringVar(&config.DeviceManagerAddress, "deviceManagerAddress", "localhost:6010",
		"Device Manager address (host:port)")
	runCmd.PersistentFlags().StringVar(&config.UnifiedLoggingAddress, "unifiedLoggingAddress", "localhost:8323",
		"Unified Logging Coordinator address (host:port)")
	runCmd.PersistentFlags().StringVar(&config.MonitoringManagerAddress, "monitoringManagerAddress", "localhost:8423",
		"Monitoring Manager address (host:port)")
	runCmd.PersistentFlags().StringVar(&config.ProvisionerManagerAddress, "provisionerManagerAddress", "localhost:8930",
		"Provisioner Manager address (host:port)")
	runCmd.PersistentFlags().StringVar(&config.InventoryManagerAddress, "inventoryManagerAddress", "localhost:5510",
		"Inventory Manager address (host:port)")

	rootCmd.AddCommand(runCmd)
}
