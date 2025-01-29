package main

import (
	"context"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-vgo/robotgo"
)

type position struct {
	X uint
	Y uint
}

type machineConfig struct {
	PrayerOrbX         uint
	PrayerOrbY         uint
	BlackPotPositions  []position
	AbsorbPotPositions []position
}

type machine struct {
	ctx                 context.Context
	sleepTimer          *time.Timer
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

func newMachine(ctx context.Context, logger *log.Logger, dryRun bool, tclock clock, pconfig machineConfig) *machine {
	return &machine{
		dryRun:              dryRun,
		logger:              logger,
		pconfig:             pconfig,
		blackPotBag:         newPotionBag(pconfig.BlackPotPositions, dryRun),
		absorbPotBag:        newPotionBag(pconfig.AbsorbPotPositions, dryRun),
		tclock:              tclock,
		nextAbsorbRepotTime: time.Now().Add(-5 * time.Hour),
		nextBlackRepotTime:  time.Now().Add(-5 * time.Hour),
		ctx:                 ctx,
		sleepTimer:          time.NewTimer(0),
	}
}

func (m *machine) run() {
	for state := flashPrayerOrb; state != nil; {
		state = state(m)
	}
}

func (m *machine) sleep(d time.Duration) (time.Time, error) {
	d = m.tclock.Scale(d)
	wasActive := m.sleepTimer.Reset(d)
	if !wasActive {
		select {
		case <-m.sleepTimer.C:
		default:
		}
	}

	select {
	case t := <-m.sleepTimer.C:
		return t, nil
	case <-m.ctx.Done():
		return time.Time{}, context.Cause(m.ctx)
	}
}

func flashPrayerOrb(m *machine) stateFunc {
	m.logger.Info("flashing prayer orb")

	if _, err := m.sleep(randomMilisecondDuration(300, 300)); err != nil {
		return errState(err)
	}

	robotgo.Move(int(m.pconfig.PrayerOrbX), int(m.pconfig.PrayerOrbY), 10.0, 1.0, 1000)
	if !m.dryRun {
		robotgo.Click("left", true)
	}
	m.drankPots = false

	if _, err := m.sleep(randomMilisecondDuration(300, 300)); err != nil {
		return errState(err)
	}
	return drinkBlackPots
}

func drinkBlackPots(m *machine) stateFunc {
	if !m.nextBlackRepotTime.Before(time.Now()) {
		return drinkAbsorbsPots
	}

	m.logger.Info("drinking black pots", "potions", m.blackPotBag.size(), "effective", m.blackPotBag.effectiveSize())

	if _, err := m.sleep(randomMilisecondDuration(100, 15)); err != nil {
		return errState(err)
	}
	err := m.blackPotBag.drink()
	if err != nil {
		m.logger.Info("out of black pots")
		return drinkAbsorbsPots
	}

	m.nextBlackRepotTime = m.tclock.Future(randomMilisecondDuration(300, 15))
	m.drankPots = true

	if _, err := m.sleep(randomSecondDuration(1, 2)); err != nil {
		return errState(err)
	}
	return drinkAbsorbsPots
}

func drinkAbsorbsPots(m *machine) stateFunc {
	if !m.nextAbsorbRepotTime.Before(time.Now()) {
		return waitForReset
	}

	m.logger.Info("drinking absorbs", "potions", m.absorbPotBag.size(), "effective", m.absorbPotBag.effectiveSize())

	if _, err := m.sleep(randomMilisecondDuration(100, 15)); err != nil {
		return errState(err)
	}
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
	ticker := time.NewTicker(m.tclock.Scale(time.Second))
	defer ticker.Stop()
	ticks := 0
	for cur := range ticker.C {
		select {
		case <-m.ctx.Done():
			return errState(context.Cause(m.ctx))
		default:
		}

		if deadline.Before(cur) {
			m.logger.Infof("waited for %v", time.Since(start))

			ticker.Stop()
			break
		}

		ticks = ticks + 1
		if ticks%5 == 0 || deadline.Sub(cur) < m.tclock.Scale(11*time.Second) {
			m.logger.Infof("%v elapsed", time.Since(start))
		}
	}
	return flashPrayerOrb
}

func errState(err error) stateFunc {
	return func(m *machine) stateFunc {
		m.logger.Error(err)
		return nil
	}
}
