package main

import (
	"fmt"
	"os"

	"ble-radar.klederson.com/internal/app"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	flagDemo    bool
	flagAdapter string
	flagRange   float64
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ble-radar",
		Short: "BLE Radar - Terminal Bluetooth device scanner with radar display",
		Long: `BLE Radar scans for Bluetooth Low Energy and classic Bluetooth devices,
displaying them on a circular ASCII radar with a Matrix-inspired aesthetic.

Requires sudo or CAP_NET_ADMIN capability for real Bluetooth scanning.
Use --demo flag for demonstration mode without Bluetooth hardware.`,
		RunE: run,
	}

	rootCmd.Flags().BoolVar(&flagDemo, "demo", false, "Run in demo mode with fake devices (no Bluetooth required)")
	rootCmd.Flags().StringVar(&flagAdapter, "adapter", "hci0", "Bluetooth adapter to use")
	rootCmd.Flags().Float64Var(&flagRange, "range", 30.0, "Maximum radar range in meters")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	model := app.New(flagDemo, flagAdapter)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithFPS(30),
	)

	// Start scanners with reference to the tea program
	if err := model.StartScanners(p); err != nil {
		if !flagDemo {
			fmt.Fprintf(os.Stderr, "\nError: %v\n\n", err)
			fmt.Fprintln(os.Stderr, "Bluetooth scanning requires elevated permissions.")
			fmt.Fprintln(os.Stderr, "Try one of:")
			fmt.Fprintln(os.Stderr, "  sudo ./ble-radar")
			fmt.Fprintln(os.Stderr, "  sudo setcap cap_net_admin+ep ./ble-radar")
			fmt.Fprintln(os.Stderr, "  ./ble-radar --demo    (demo mode, no hardware needed)")
			return err
		}
	}

	_, err := p.Run()
	return err
}
