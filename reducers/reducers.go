package reducers

import(
	"github.com/kjirou/tower-of-go/models"
	"github.com/kjirou/tower-of-go/utils"
	"time"
)

type FourDirection int
const (
	FourDirectionUp FourDirection = iota
	FourDirectionRight
	FourDirectionDown
	FourDirectionLeft
)

func StartOrRestartGame(state models.State) (*models.State, bool, error) {
	game := state.GetGame()
	field := state.GetField()

	// Generate a new maze.
	// Remove the hero.
	err := field.ResetMaze()
	if err != nil {
		return &state, false, err
	}

	// Replace the hero.
	heroFieldElement, _ := field.At(models.HeroPosition)
	heroFieldElement.UpdateObjectClass("hero")

	// Start the new game.
	game.Reset()
	game.Start(state.GetExecutionTime())

	return &state, true, nil
}

func AdvanceTime(state models.State, delta time.Duration) (*models.State, bool, error) {
	game := state.GetGame()
	field := state.GetField()

	// In the game.
	if game.IsStarted() && !game.IsFinished() {
		// The hero climbs up the stairs.
		heroFieldElement, getElementOfHeroErr := field.GetElementOfHero()
		if getElementOfHeroErr != nil {
			return &state, false, getElementOfHeroErr
		}
		if (heroFieldElement.GetFloorObjectClass() == "upstairs") {
			// Generate a new maze.
			// Remove the hero.
			err := field.ResetMaze()
			if err != nil {
				return &state, false, err
			}

			// Relocate the hero to the entrance.
			replacedHeroFieldElement, _ := field.At(models.HeroPosition)
			replacedHeroFieldElement.UpdateObjectClass("hero")

			game.IncrementFloorNumber()
		}

		// Time over of this game.
		remainingTime := game.CalculateRemainingTime(state.GetExecutionTime())
		if remainingTime == 0 {
			game.Finish()
		}
	}

	state.AlterExecutionTime(delta)

	return &state, true, nil
}

func WalkHero(state models.State, direction FourDirection) (*models.State, bool, error) {
	game := state.GetGame()
	if game.IsFinished() {
		return &state, true, nil
	}

	field := state.GetField()
	element, getElementOfHeroErr := field.GetElementOfHero()
	if getElementOfHeroErr != nil {
		return &state, false, getElementOfHeroErr
	}
	nextY := element.GetPosition().GetY()
	nextX := element.GetPosition().GetX()
	switch direction {
	case FourDirectionUp:
		nextY -= 1
	case FourDirectionRight:
		nextX += 1
	case FourDirectionDown:
		nextY += 1
	case FourDirectionLeft:
		nextX -= 1
	}
	position := element.GetPosition()
	nextPosition := &utils.MatrixPosition{
		Y: nextY,
		X: nextX,
	}
	if nextPosition.Validate(field.MeasureRowLength(), field.MeasureColumnLength()) {
		element, err := field.At(nextPosition)
		if err != nil {
			return &state, false, err
		} else if element.IsObjectEmpty() {
			err := field.MoveObject(position, nextPosition)
			return &state, err == nil, err
		}
	}
	return &state, true, nil
}
