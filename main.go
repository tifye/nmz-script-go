package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/log"
	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

type config struct {
	DryRun                bool
	TimeScale             float32
	WindowXOffset         uint
	NumberOfBlackPotions  uint
	NumberOfAbsorbPotions uint
}

var targetDisplay = 2

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

	robotgo.DisplayID = targetDisplay
	x, y, w, h := robotgo.GetDisplayBounds(targetDisplay)
	logger.Debug("target display bounds", "x", x, "y", y, "w", w, "h", h)

	pconfig := machineConfig{
		NumBlackPotions:  conf.NumberOfBlackPotions,
		NumAbsorbPotions: conf.NumberOfAbsorbPotions,
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
		conf.DryRun,
		simulatedClock{timeScale: conf.TimeScale},
		pconfig,
	)
	m.run()
}

func validateConfig(conf config) error {
	totalPotions := conf.NumberOfAbsorbPotions + conf.NumberOfBlackPotions
	if totalPotions > 7*4 {
		return fmt.Errorf("total number of potions exceed inventory size, total: %d", totalPotions)
	}
	return nil
}
