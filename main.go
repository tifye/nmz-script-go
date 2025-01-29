package main

import (
	"github.com/charmbracelet/log"
	"github.com/go-vgo/robotgo"
)

var targetDisplay = 2

func main() {
	logger := log.Default()
	logger.SetLevel(log.DebugLevel)

	numDisplays := robotgo.DisplaysNum()
	logger.Debugf("%d displays found", numDisplays)
	logger.Debugf("no display set, defaulting to display %d", targetDisplay)

	robotgo.DisplayID = targetDisplay

	x, y, w, h := robotgo.GetDisplayBounds(targetDisplay)
	logger.Debug("target display bounds", "x", x, "y", y, "w", w, "h", h)

	var windowOffsetX uint = 2560
	wf := float32(int(windowOffsetX) + w)
	hf := float32(h)
	pconfig := machineConfig{
		PrayerOrbX: windowOffsetX + 1666,
		PrayerOrbY: 148,
		BlackPotPositions: []position{
			{x: uint(float32(0.870833) * wf), y: uint(float32(0.620370) * hf)},
			{x: uint(float32(0.898958) * wf), y: uint(float32(0.624074) * hf)},
			{x: uint(float32(0.925000) * wf), y: uint(float32(0.619444) * hf)},
			{x: uint(float32(0.955208) * wf), y: uint(float32(0.618518) * hf)},
			{x: uint(float32(0.871354) * wf), y: uint(float32(0.662037) * hf)},
			{x: uint(float32(0.901041) * wf), y: uint(float32(0.660185) * hf)},
			{x: uint(float32(0.928125) * wf), y: uint(float32(0.666666) * hf)},
			{x: uint(float32(0.957291) * wf), y: uint(float32(0.659259) * hf)},
			{x: uint(float32(0.872395) * wf), y: uint(float32(0.701851) * hf)},
			{x: uint(float32(0.900000) * wf), y: uint(float32(0.705555) * hf)},
			{x: uint(float32(0.924479) * wf), y: uint(float32(0.710185) * hf)},
			{x: uint(float32(0.956250) * wf), y: uint(float32(0.700000) * hf)},
		},
		AbsorbPotPositions: []position{
			{x: 1675 + windowOffsetX, y: 809},
			{x: 1725 + windowOffsetX, y: 805},
			{x: 1779 + windowOffsetX, y: 808},
			{x: 1829 + windowOffsetX, y: 803},
			{x: 1674 + windowOffsetX, y: 846},
			{x: 1726 + windowOffsetX, y: 848},
			{x: 1782 + windowOffsetX, y: 847},
			{x: 1833 + windowOffsetX, y: 855},
			{x: 1677 + windowOffsetX, y: 893},
			{x: 1726 + windowOffsetX, y: 900},
			{x: 1779 + windowOffsetX, y: 992},
			{x: 1833 + windowOffsetX, y: 889},
			{x: 1673 + windowOffsetX, y: 937},
			{x: 1725 + windowOffsetX, y: 941},
		},
	}

	m := newMachine(logger, true, simulatedClock{timeScale: 1}, pconfig)
	m.run()
}
