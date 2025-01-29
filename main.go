package main

import (
	"context"
	"errors"
	"os"
	"os/signal"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/log"
	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

type config struct {
	DryRun        bool
	TimeScale     float32
	WindowXOffset uint
	PrayerOrb     point
	BlackPotions  []point
	AbsorbPotions []point
}

type point struct {
	X float32
	Y float32
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

	numDisplays := robotgo.DisplaysNum()
	logger.Debugf("%d displays found", numDisplays)
	logger.Debugf("no display set, defaulting to display %d", targetDisplay)

	robotgo.DisplayID = targetDisplay

	x, y, w, h := robotgo.GetDisplayBounds(targetDisplay)
	logger.Debug("target display bounds", "x", x, "y", y, "w", w, "h", h)

	wf := float32(w)
	hf := float32(h)
	pconfig := machineConfig{
		PrayerOrbX:         uint(conf.PrayerOrb.X) + conf.WindowXOffset,
		PrayerOrbY:         uint(conf.PrayerOrb.Y),
		BlackPotPositions:  make([]position, len(conf.BlackPotions)),
		AbsorbPotPositions: make([]position, len(conf.AbsorbPotions)),
	}
	for i, p := range conf.BlackPotions {
		pconfig.BlackPotPositions[i].X = uint(p.X*wf) + conf.WindowXOffset
		pconfig.BlackPotPositions[i].Y = uint(p.Y * hf)
	}
	for i, p := range conf.AbsorbPotions {
		pconfig.AbsorbPotPositions[i].X = uint(p.X*wf) + conf.WindowXOffset
		pconfig.AbsorbPotPositions[i].Y = uint(p.Y * hf)
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
