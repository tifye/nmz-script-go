package main

import (
	"fmt"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/vcaesar/gcv"
)

const inventoryTopLeftCorner = "inventory_top_left_corner.png"
const inventoryBottomRightCorner = "inventory_bottom_right_corner.png"
const xOffset = 0

func main() {
	robotgo.DisplayID = 1
	robotgo.SaveCapture("test.png", -2560, 0, 2560, 1440)

	results := gcv.FindAllImgFile(inventoryTopLeftCorner, "test.png")
	fmt.Printf("found %d results for %s\n", len(results), inventoryTopLeftCorner)
	if len(results) <= 0 {
		return
	}
	topLeft := results[0]
	robotgo.Move(topLeft.Middle.X+xOffset, topLeft.Middle.Y)

	time.Sleep(2 * time.Second)
	results = gcv.FindAllImgFile(inventoryBottomRightCorner, "test.png")
	fmt.Printf("found %d results for %s\n", len(results), inventoryBottomRightCorner)
	if len(results) <= 0 {
		return
	}
	bottomRight := results[0]
	robotgo.Move(bottomRight.Middle.X+xOffset, bottomRight.Middle.Y)

	width := bottomRight.Middle.X - topLeft.Middle.X
	height := bottomRight.Middle.Y - topLeft.Middle.Y
	cellWidth := int(float32(width) / 4.0)
	cellHeight := int(float32(height) / 7.0)
	cellMiddleXOffset := cellWidth / 2.0
	cellMiddleYOffset := cellHeight / 2.0

	for cy := range 7 {
		for cx := range 4 {
			time.Sleep(300 * time.Millisecond)
			x := topLeft.Middle.X + cellWidth*cx + cellMiddleXOffset
			y := topLeft.Middle.Y + cellHeight*cy + cellMiddleYOffset
			robotgo.Move(x+xOffset, y)
		}
	}
}
