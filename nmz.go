package main

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"image"
	"math/rand"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-vgo/robotgo"
	"github.com/vcaesar/gcv"
)

type position struct {
	X uint
	Y uint
}

type machineConfig struct {
	PrayerOrbPosition position
	NumBlackPotions   uint
	NumAbsorbPotions  uint
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
		blackPotBag:         newPotionBag(pconfig.NumBlackPotions, dryRun),
		absorbPotBag:        newPotionBag(pconfig.NumAbsorbPotions, dryRun),
		tclock:              tclock,
		nextAbsorbRepotTime: time.Now().Add(-5 * time.Hour),
		nextBlackRepotTime:  time.Now().Add(-5 * time.Hour),
		ctx:                 ctx,
		sleepTimer:          time.NewTimer(0),
	}
}

func (m *machine) run() {
	for state := calibrate; state != nil; {
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

	moveDeviateRandom(m.pconfig.PrayerOrbPosition.X, m.pconfig.PrayerOrbPosition.Y)
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

//go:embed calibration/landmark_1.png
var landmark1CalibrationRef []byte

//go:embed calibration/landmark_2.png
var landmark2CalibrationRef []byte

//go:embed calibration/landmark_3.png
var landmark3CalibrationRef []byte

func calibrate(m *machine) stateFunc {
	m.logger.Info("begining calibration")

	landmark1, _, err := image.Decode(bytes.NewReader(landmark1CalibrationRef))
	if err != nil {
		return errState(err)
	}
	landmark2, _, err := image.Decode(bytes.NewReader(landmark2CalibrationRef))
	if err != nil {
		return errState(err)
	}
	landmark3, _, err := image.Decode(bytes.NewReader(landmark3CalibrationRef))
	if err != nil {
		return errState(err)
	}

	landmarks := []image.Image{landmark1, landmark2, landmark3}

	displayId := -1
	x, y, w, h := 0, 0, 0, 0
	var img image.Image
displays:
	for i := range robotgo.DisplaysNum() {
		x, y, w, h = robotgo.GetDisplayBounds(i)
		m.logger.Debugf("display %d, [x,y][%d,%d] [w,h][%d,%d]", i, x, y, w, h)

		robotgo.DisplayID = i
		img, err = robotgo.CaptureImg(x, y, w, h)
		if err != nil {
			return errState(err)
		}

		for j, l := range landmarks {
			results := gcv.FindAllImg(l, img)
			if len(results) > 0 {
				m.logger.Debugf("landmark %d matched on display %d", j+1, i)
				displayId = i
				break displays
			} else {
				m.logger.Debugf("landmark %d failed", j+1)
			}
		}
	}

	if displayId == -1 {
		return errState(errors.New("could not find display with game"))
	}

	robotgo.DisplayID = displayId
	xOffset := 0
	for i := range robotgo.DisplaysNum() {
		dx, _, dw, _ := robotgo.GetDisplayBounds(i)
		if dx < x {
			xOffset = xOffset + dw
		}
	}

	return calibratePrayerOrb(uint(xOffset), img)
}

//go:embed calibration/prayer_orb.png
var prayerOrbCalibrationRef []byte

func calibratePrayerOrb(xOffset uint, screenshot image.Image) stateFunc {
	return func(m *machine) stateFunc {
		m.logger.Info("calibrating prayer orb")

		refImageFile := bytes.NewReader(prayerOrbCalibrationRef)
		refImage, _, err := image.Decode(refImageFile)
		if err != nil {
			return errState(err)
		}

		results := gcv.FindAllImg(refImage, screenshot)
		m.logger.Debugf("found %d prayer orb matches", len(results))
		if len(results) <= 0 {
			return errState(errors.New("prayer orb calibration: could not locate prayer orb"))
		}

		m.pconfig.PrayerOrbPosition.X = uint(results[0].Middle.X) + xOffset
		m.pconfig.PrayerOrbPosition.Y = uint(results[0].Middle.Y)

		moveDeviateRandom(m.pconfig.PrayerOrbPosition.X, m.pconfig.PrayerOrbPosition.Y)
		if _, err := m.sleep(1 * time.Second); err != nil {
			return errState(err)
		}

		return calibrateInventory(xOffset, screenshot)
	}
}

var (
	//go:embed calibration/inventory_bottom_right_corner.png
	inventoryBottomRightCalibrationRef []byte
	//go:embed calibration/inventory_top_left_corner.png
	inventoryTopLeftCalibrationRef []byte
)

func calibrateInventory(xOffset uint, screenshot image.Image) stateFunc {
	return func(m *machine) stateFunc {
		topLeftRefFile := bytes.NewReader(inventoryTopLeftCalibrationRef)
		topLeftRef, _, err := image.Decode(topLeftRefFile)
		if err != nil {
			return errState(err)
		}
		results := gcv.FindAllImg(topLeftRef, screenshot)
		m.logger.Debugf("found %d top left inventory matches", len(results))
		if len(results) <= 0 {
			return errState(errors.New("inventory calibration: could not locate inventory top left corner"))
		}
		topLeft := results[0]

		bottomRightRefFile := bytes.NewReader(inventoryBottomRightCalibrationRef)
		bottomRightRef, _, err := image.Decode(bottomRightRefFile)
		if err != nil {
			return errState(err)
		}
		results = gcv.FindAllImg(bottomRightRef, screenshot)
		m.logger.Debugf("found %d bottom right inventory matches", len(results))
		if len(results) <= 0 {
			return errState(errors.New("inventory calibration: could not locate inventory bottom right corner"))
		}
		bottomRight := results[0]

		moveDeviateRandom(uint(topLeft.Middle.X)+xOffset, uint(topLeft.Middle.Y))
		if _, err := m.sleep(time.Second); err != nil {
			return errState(err)
		}
		moveDeviateRandom(uint(bottomRight.Middle.X)+xOffset, uint(bottomRight.Middle.Y))
		if _, err := m.sleep(time.Second); err != nil {
			return errState(err)
		}

		width := bottomRight.Middle.X - topLeft.Middle.X
		height := bottomRight.Middle.Y - topLeft.Middle.Y
		cellWidth := int(float32(width) / 4.0)
		cellHeight := int(float32(height) / 7.0)
		cellMiddleXOffset := cellWidth / 2.0
		cellMiddleYOffset := cellHeight / 2.0

		inventorySlots := make([]position, MaxInventorySlots)
		for cy := range InventoryRows {
			for cx := range InventoryColumns {
				x := topLeft.Middle.X + cellWidth*cx + cellMiddleXOffset
				y := topLeft.Middle.Y + cellHeight*cy + cellMiddleYOffset

				si := cy*InventoryColumns + cx
				inventorySlots[si].X = uint(x) + xOffset
				inventorySlots[si].Y = uint(y)
			}
		}

		delay := 200 * time.Millisecond
		for i := range m.pconfig.NumBlackPotions {
			if _, err := m.sleep(delay); err != nil {
				return errState(err)
			}

			slot := inventorySlots[i]
			m.blackPotBag.potions[i].x = slot.X
			m.blackPotBag.potions[i].y = slot.Y
			moveDeviateRandom(slot.X, slot.Y)
		}

		for i := range m.pconfig.NumAbsorbPotions {
			if _, err := m.sleep(delay); err != nil {
				return errState(err)
			}

			slot := inventorySlots[m.pconfig.NumBlackPotions+i]
			m.absorbPotBag.potions[i].x = slot.X
			m.absorbPotBag.potions[i].y = slot.Y
			moveDeviateRandom(slot.X, slot.Y)
		}

		return flashPrayerOrb
	}
}

func moveDeviateRandom(x, y uint) {
	n := rand.Intn(8) - 4
	nx := int(x) + n
	ny := int(y) + n
	robotgo.Move(nx, ny)
}
