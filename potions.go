package main

import (
	"errors"

	"github.com/go-vgo/robotgo"
)

var (
	ErrOutOfPotions = errors.New("out of potions")
)

type pot struct {
	x     uint
	y     uint
	doses uint8
}

type potionBag struct {
	dryRun  bool
	potions []pot
}

func newPotionBag(potPositions []position, dryRun bool) *potionBag {
	potBag := &potionBag{
		potions: make([]pot, len(potPositions)),
		dryRun:  dryRun,
	}
	for i := range potBag.potions {
		potBag.potions[i] = pot{
			x:     potPositions[i].x,
			y:     potPositions[i].y,
			doses: 4,
		}
	}
	return potBag
}

func (pb *potionBag) size() uint {
	return uint(len(pb.potions))
}

func (pb *potionBag) effectiveSize() uint {
	sum := 0
	for _, p := range pb.potions {
		sum = sum + int(p.doses)
	}
	return uint(sum)
}

func (pb *potionBag) drink() error {
	if pb.size() == 0 {
		return ErrOutOfPotions
	}

	idx := len(pb.potions) - 1
	pot := pb.potions[idx]

	robotgo.Move(int(pot.x), int(pot.y))
	if !pb.dryRun {
		robotgo.Click("left")
	}

	newDoses := pot.doses - 1
	pb.potions[idx].doses = newDoses
	if newDoses <= 0 {
		pb.potions = pb.potions[:idx] // Acceptable leak
	}

	return nil
}
