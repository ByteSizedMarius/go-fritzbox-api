package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	fritzbox "github.com/ByteSizedMarius/go-fritzbox-api"
	"github.com/ByteSizedMarius/go-fritzbox-api/smart"
)

func main() {
	user := flag.String("user", "", "FRITZ!Box username")
	pass := flag.String("pass", "", "FRITZ!Box password")
	uid := flag.String("uid", "", "Thermostat UID (optional, shows details)")
	flag.Parse()

	if *user == "" || *pass == "" {
		fmt.Fprintln(os.Stderr, "Usage: thermostat -user=<username> -pass=<password> [-uid=<uid>]")
		os.Exit(1)
	}

	client := fritzbox.New(*user, *pass)
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Connection failed: %v\n", err)
		os.Exit(1)
	}

	if *uid == "" {
		listThermostats(client)
	} else {
		showDetails(client, *uid)
	}
}

func listThermostats(client *fritzbox.Client) {
	thermostats, err := smart.GetAllThermostats(client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get thermostats: %v\n", err)
		os.Exit(1)
	}

	if len(thermostats) == 0 {
		fmt.Println("No thermostats found")
		return
	}

	fmt.Printf("Found %d thermostat(s):\n\n", len(thermostats))
	for _, t := range thermostats {
		status := "connected"
		if !t.IsConnected {
			status = "disconnected"
		}
		fmt.Printf("%-30s  Current: %5.1f°C  Target: %5.1f°C  [%s]\n",
			t.Name, t.CurrentTemp, t.TargetTemp, status)
		fmt.Printf("  UID: %s\n\n", t.UID)
	}
}

func showDetails(client *fritzbox.Client, uid string) {
	t, err := smart.GetThermostat(client, uid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get thermostat: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== %s ===\n\n", t.Name)

	fmt.Println("Identification:")
	fmt.Printf("  UID:         %s\n", t.UID)
	fmt.Printf("  AIN:         %s\n", t.AIN)
	fmt.Printf("  Connected:   %v\n", t.IsConnected)

	fmt.Println("\nTemperatures:")
	fmt.Printf("  Current:     %.1f°C\n", t.CurrentTemp)
	fmt.Printf("  Target:      %.1f°C\n", t.TargetTemp)
	fmt.Printf("  Comfort:     %.1f°C\n", t.ComfortTemp)
	fmt.Printf("  Reduced:     %.1f°C\n", t.ReducedTemp)
	fmt.Printf("  Offset:      %.1f°C\n", t.TempOffset)

	fmt.Println("\nStatus:")
	fmt.Printf("  Boost:       %v", t.Boost.Active)
	if t.Boost.Active && !t.Boost.EndTime.IsZero() {
		fmt.Printf(" (until %s)", t.Boost.EndTime.Format("15:04"))
	}
	fmt.Println()
	fmt.Printf("  Window Open: %v", t.WindowOpen.Active)
	if t.WindowOpen.Active && !t.WindowOpen.EndTime.IsZero() {
		fmt.Printf(" (until %s)", t.WindowOpen.EndTime.Format("15:04"))
	}
	fmt.Println()
	fmt.Printf("  Summer:      %v\n", t.IsSummerActive)
	fmt.Printf("  Holiday:     %v\n", t.IsHolidayActive)
	fmt.Printf("  Locked:      %v\n", t.IsLocked)

	fmt.Println("\nBattery:")
	fmt.Printf("  Level:       %d%%\n", t.BatteryLevel)
	fmt.Printf("  Low:         %v\n", t.IsBatteryLow)

	if t.NextChange != nil {
		fmt.Println("\nNext Scheduled Change:")
		fmt.Printf("  Time:        %s\n", t.NextChange.Time.Format("Mon 15:04"))
		fmt.Printf("  Temperature: %.1f°C\n", t.NextChange.Temperature)
	}

	handle := smart.NewThermostatHandle(client, uid)
	cfg, err := handle.GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nFailed to get config: %v\n", err)
		return
	}

	fmt.Println("\nConfiguration:")

	fmt.Printf("  Summer Period: %v", cfg.SummerPeriod.Enabled)
	if cfg.SummerPeriod.Enabled && !cfg.SummerPeriod.StartTime.IsZero() {
		fmt.Printf(" (%s - %s)",
			cfg.SummerPeriod.StartTime.Format("Jan 2"),
			cfg.SummerPeriod.EndTime.Format("Jan 2"))
	}
	fmt.Println()

	if len(cfg.HolidayPeriods) > 0 {
		fmt.Printf("  Holidays:      %d configured\n", len(cfg.HolidayPeriods))
		for i, h := range cfg.HolidayPeriods {
			fmt.Printf("    [%d] %s - %s at %.1f°C\n", i,
				h.StartTime.Format("Jan 2 15:04"),
				h.EndTime.Format("Jan 2 15:04"),
				h.Temperature)
		}
	} else {
		fmt.Println("  Holidays:      none")
	}

	if len(cfg.WeeklySchedule) > 0 {
		fmt.Printf("  Schedule:      %d entries\n", len(cfg.WeeklySchedule))
	} else {
		fmt.Println("  Schedule:      none")
	}
}

// Example write operations (not executed by CLI)

func exampleSetTemperature(client *fritzbox.Client, uid string) {
	handle := smart.NewThermostatHandle(client, uid)

	handle.SetTargetTemperature(21.5)
	handle.TurnOff()
	handle.TurnOn()

	handle.SetComfortPreset(22.0)
	handle.SetReducedPreset(17.0)
	handle.SetTemperatureOffset(1.5)
}

func exampleBoostAndWindowOpen(client *fritzbox.Client, uid string) {
	handle := smart.NewThermostatHandle(client, uid)

	handle.SetBoost(30)
	handle.DeactivateBoost()

	handle.SetWindowOpen(15)
	handle.DeactivateWindowOpen()
	handle.SetWindowOpenDetection(15, "medium")
}

func exampleSchedulesAndPeriods(client *fritzbox.Client, uid string) {
	handle := smart.NewThermostatHandle(client, uid)

	handle.SetSummerPeriod(true, 5, 1, 9, 30)

	start := time.Date(2025, 12, 24, 0, 0, 0, 0, time.Local)
	end := time.Date(2026, 1, 2, 0, 0, 0, 0, time.Local)
	handle.AddHoliday(start, end, 18.0)
	handle.RemoveHoliday(0)
	handle.ClearHolidays()
}

func exampleLocks(client *fritzbox.Client, uid string) {
	handle := smart.NewThermostatHandle(client, uid)
	handle.SetLocks(true, false)
	handle.SetAdaptiveHeating(true)
}
