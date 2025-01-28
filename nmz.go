package main

import (
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-vgo/robotgo"
)

type position struct {
	x uint
	y uint
}

type machineConfig struct {
	PrayerOrbX         uint
	PrayerOrbY         uint
	BlackPotPositions  []position
	AbsorbPotPositions []position
}

type machine struct {
	dryRun              bool
	logger              *log.Logger
	pconfig             machineConfig
	drankPots           bool
	nextBlackRepotTime  time.Time
	nextAbsorbRepotTime time.Time
	blackPotBag         *potionBag
	absorbPotBag        *potionBag
	tclock              clock
}

type stateFunc func(*machine) stateFunc

func newMachine(logger *log.Logger, dryRun bool, tclock clock, pconfig machineConfig) *machine {

	return &machine{
		dryRun:       dryRun,
		logger:       logger,
		pconfig:      pconfig,
		blackPotBag:  newPotionBag(pconfig.BlackPotPositions, dryRun),
		absorbPotBag: newPotionBag(pconfig.AbsorbPotPositions, dryRun),
		tclock:       tclock,
	}
}

func (m *machine) run() {
	for state := flashPrayerOrb; state != nil; {
		state = state(m)
	}
}

func flashPrayerOrb(m *machine) stateFunc {
	m.logger.Info("flashing prayer orb")

	m.tclock.Sleep(randomMilisecondDuration(300, 300))
	robotgo.Move(int(m.pconfig.PrayerOrbX), int(m.pconfig.PrayerOrbY))
	if !m.dryRun {
		robotgo.Click("left", true)
	}
	m.drankPots = false
	return drinkBlackPots
}

func drinkBlackPots(m *machine) stateFunc {
	m.logger.Debug("before drink back pots", "now", time.Now(), "next", m.nextAbsorbRepotTime)
	if m.nextBlackRepotTime.Before(time.Now()) {
		return drinkAbsorbsPots
	}

	m.logger.Info("drinking pots", "potions", m.blackPotBag.size(), "effective", m.blackPotBag.effectiveSize())

	m.tclock.Sleep(randomMilisecondDuration(100, 15))
	err := m.blackPotBag.drink()
	if err != nil {
		m.logger.Info("out of black pots")
		return drinkAbsorbsPots
	}

	m.nextBlackRepotTime = m.tclock.Future(randomMilisecondDuration(300, 15))
	m.drankPots = true

	m.tclock.Sleep(randomSecondDuration(1, 2))
	return drinkAbsorbsPots
}

func drinkAbsorbsPots(m *machine) stateFunc {
	m.logger.Debug("before drink absorb pots", "now", time.Now(), "next", m.nextAbsorbRepotTime)
	if m.nextAbsorbRepotTime.Before(time.Now()) {
		return waitForReset
	}

	m.logger.Info("drinking absorbs", "potions", m.absorbPotBag.size(), "effective", m.absorbPotBag.effectiveSize())

	m.tclock.Sleep(randomMilisecondDuration(100, 15))
	err := m.absorbPotBag.drink()
	if err != nil {
		m.logger.Info("out of absorb pots")
		return waitForReset
	}

	m.nextAbsorbRepotTime = m.tclock.Future(randomMilisecondDuration(300, 15))
	m.drankPots = true

	return waitForReset
}

func waitForReset(m *machine) stateFunc {
	m.logger.Info("wait for reset", "drankPots", m.drankPots)

	var waitDuration time.Duration
	if m.drankPots {
		waitDuration = m.tclock.Scale(randomSecondDuration(35, 5))
	} else {
		waitDuration = m.tclock.Scale(randomSecondDuration(45, 5))
	}

	start := time.Now()
	deadline := start.Add(waitDuration)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	ticks := 0
	for cur := range ticker.C {
		if deadline.Before(cur) {
			m.logger.Infof("waited for %v", time.Since(start))

			ticker.Stop()
			break
		}

		ticks = ticks + 1
		if ticks%5 == 0 || deadline.Sub(cur) < 11*time.Second {
			m.logger.Infof("%v elapsed", time.Since(start))
		}
	}
	return flashPrayerOrb
}
