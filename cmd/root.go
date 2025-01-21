package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	log "github.com/starttoaster/proxmox-exporter/internal/logger"

	"github.com/starttoaster/proxmox-exporter/internal/http"
	"github.com/starttoaster/proxmox-exporter/internal/proxmox"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "proxmox-exporter",
	Short: "Prometheus exporter for Proxmox",

	Run: func(cmd *cobra.Command, args []string) {
		// Init logger
		log.Init(viper.GetString("log-level"))

		// Initialize proxmox client package
		err := proxmox.Init(
			strings.Split(viper.GetString("proxmox-endpoints"), ","),
			viper.GetString("proxmox-token-id"),
			viper.GetString("proxmox-token"),
			viper.GetBool("proxmox-api-insecure"))
		if err != nil {
			log.Logger.Error(err.Error())
			os.Exit(1)
		}

		// Create http server
		m, err := http.NewServer(string(viper.GetString("server-addr")), uint16(viper.GetUint("server-port")))
		if err != nil {
			log.Logger.Error(err.Error())
			os.Exit(1)
		}

		// Start http server
		err = m.StartServer()
		if err != nil {
			log.Logger.Error(err.Error())
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// Read in environment variables that match defined config pattern
	viper.SetEnvPrefix("PROXMOX_EXPORTER")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	rootCmd.PersistentFlags().String("log-level", "info", "The log-level for the application, can be one of info, warn, error, debug.")
	rootCmd.PersistentFlags().String("server-addr", "0.0.0.0", "The address on which the exporter listens")
	rootCmd.PersistentFlags().Uint16("server-port", 8080, "The port the metrics server binds to.")
	rootCmd.PersistentFlags().String("proxmox-endpoints", "", "The Proxmox API endpoint, you can pass in multiple endpoints separated by commas (ex: https://localhost:8006/)")
	rootCmd.PersistentFlags().String("proxmox-token-id", "", "Proxmox API token ID")
	rootCmd.PersistentFlags().String("proxmox-token", "", "Proxmox API token")
	rootCmd.PersistentFlags().Bool("proxmox-api-insecure", false, "Whether or not this client should accept insecure connections to Proxmox (default: false)")

	err := viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	if err != nil {
		log.Logger.Error(err.Error())
		os.Exit(1)
	}

	err = viper.BindPFlag("server-addr", rootCmd.PersistentFlags().Lookup("server-addr"))
	if err != nil {
		log.Logger.Error(err.Error())
		os.Exit(1)
	}

	err = viper.BindPFlag("server-port", rootCmd.PersistentFlags().Lookup("server-port"))
	if err != nil {
		log.Logger.Error(err.Error())
		os.Exit(1)
	}

	err = viper.BindPFlag("proxmox-endpoints", rootCmd.PersistentFlags().Lookup("proxmox-endpoints"))
	if err != nil {
		log.Logger.Error(err.Error())
		os.Exit(1)
	}

	err = viper.BindPFlag("proxmox-token-id", rootCmd.PersistentFlags().Lookup("proxmox-token-id"))
	if err != nil {
		log.Logger.Error(err.Error())
		os.Exit(1)
	}

	err = viper.BindPFlag("proxmox-token", rootCmd.PersistentFlags().Lookup("proxmox-token"))
	if err != nil {
		log.Logger.Error(err.Error())
		os.Exit(1)
	}

	err = viper.BindPFlag("proxmox-api-insecure", rootCmd.PersistentFlags().Lookup("proxmox-api-insecure"))
	if err != nil {
		log.Logger.Error(err.Error())
		os.Exit(1)
	}
}
