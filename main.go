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
	wf := float32(w)
	hf := float32(h)
	pconfig := machineConfig{
		PrayerOrbX: uint(0.867708*wf) + windowOffsetX,
		PrayerOrbY: uint(0.137037 * hf),
		BlackPotPositions: []position{
			{x: uint(float32(0.870833)*wf) + windowOffsetX, y: uint(float32(0.620370) * hf)},
			{x: uint(float32(0.898958)*wf) + windowOffsetX, y: uint(float32(0.624074) * hf)},
			{x: uint(float32(0.925000)*wf) + windowOffsetX, y: uint(float32(0.619444) * hf)},
			{x: uint(float32(0.955208)*wf) + windowOffsetX, y: uint(float32(0.618518) * hf)},
			{x: uint(float32(0.871354)*wf) + windowOffsetX, y: uint(float32(0.662037) * hf)},
			{x: uint(float32(0.901041)*wf) + windowOffsetX, y: uint(float32(0.660185) * hf)},
			{x: uint(float32(0.928125)*wf) + windowOffsetX, y: uint(float32(0.666666) * hf)},
			{x: uint(float32(0.957291)*wf) + windowOffsetX, y: uint(float32(0.659259) * hf)},
			{x: uint(float32(0.872395)*wf) + windowOffsetX, y: uint(float32(0.701851) * hf)},
			{x: uint(float32(0.900000)*wf) + windowOffsetX, y: uint(float32(0.705555) * hf)},
			{x: uint(float32(0.924479)*wf) + windowOffsetX, y: uint(float32(0.710185) * hf)},
			{x: uint(float32(0.956250)*wf) + windowOffsetX, y: uint(float32(0.700000) * hf)},
		},
		AbsorbPotPositions: []position{
			{x: uint(float32(0.872395)*wf) + windowOffsetX, y: uint(float32(0.749074) * hf)},
			{x: uint(float32(0.898437)*wf) + windowOffsetX, y: uint(float32(0.745370) * hf)},
			{x: uint(float32(0.926562)*wf) + windowOffsetX, y: uint(float32(0.748148) * hf)},
			{x: uint(float32(0.952604)*wf) + windowOffsetX, y: uint(float32(0.743518) * hf)},
			{x: uint(float32(0.871875)*wf) + windowOffsetX, y: uint(float32(0.783333) * hf)},
			{x: uint(float32(0.898958)*wf) + windowOffsetX, y: uint(float32(0.785185) * hf)},
			{x: uint(float32(0.928125)*wf) + windowOffsetX, y: uint(float32(0.784259) * hf)},
			{x: uint(float32(0.954687)*wf) + windowOffsetX, y: uint(float32(0.791666) * hf)},
			{x: uint(float32(0.873437)*wf) + windowOffsetX, y: uint(float32(0.826851) * hf)},
			{x: uint(float32(0.898958)*wf) + windowOffsetX, y: uint(float32(0.833333) * hf)},
			{x: uint(float32(0.926562)*wf) + windowOffsetX, y: uint(float32(0.918518) * hf)},
			{x: uint(float32(0.954687)*wf) + windowOffsetX, y: uint(float32(0.823148) * hf)},
			{x: uint(float32(0.871354)*wf) + windowOffsetX, y: uint(float32(0.867592) * hf)},
			{x: uint(float32(0.898437)*wf) + windowOffsetX, y: uint(float32(0.871296) * hf)},
		},
	}

	m := newMachine(logger, true, simulatedClock{timeScale: 1}, pconfig)
	m.run()
}
