package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/log"
	hook "github.com/robotn/gohook"
)

const (
	InventoryRows     = 7
	InventoryColumns  = 4
	MaxInventorySlots = InventoryRows * InventoryColumns
	MaxDoses          = 4
)

type config struct {
	DryRun                bool
	VisualDebug           bool
	TimeScale             float32
	NumberOfBlackPotions  uint
	NumberOfAbsorbPotions uint
	DosesOfFirstBlack     uint
	DosesOfFirstAbsorb    uint
}

func main() {
	logger := log.Default()
	logger.SetLevel(log.DebugLevel)

	var conf config
	_, err := toml.DecodeFile("config.toml", &conf)
	if err != nil {
		logger.Fatalf("failed to load config: %s", err)
	}

	err = validateConfig(conf)
	if err != nil {
		logger.Fatalf("invalid config: %s", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	ctx, cancelCause := context.WithCancelCause(ctx)

	hook.Register(hook.KeyDown, []string{"esc"}, func(e hook.Event) {
		cancelCause(errors.New("user interrupt"))
	})
	hookEvsChan := hook.Start()
	defer hook.End()
	go hook.Process(hookEvsChan)

	m := newMachine(
		ctx,
		logger,
		simulatedClock{timeScale: conf.TimeScale},
		conf,
	)
	m.run()
}

func validateConfig(conf config) error {
	totalPotions := conf.NumberOfAbsorbPotions + conf.NumberOfBlackPotions
	if totalPotions > MaxInventorySlots {
		return fmt.Errorf("total number of potions exceed inventory size, total: %d", totalPotions)
	}
	return nil
}
