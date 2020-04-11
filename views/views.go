package views

import (
	"fmt"
	"github.com/kjirou/tower_of_go/utils"
	"github.com/nsf/termbox-go"
	"strings"
)

type ScreenElement struct {
	Symbol          rune
	ForegroundColor termbox.Attribute
	BackgroundColor termbox.Attribute
}

func (screenElement *ScreenElement) renderWithFieldElement(fieldElement utils.IFieldElement) {
	symbol := '.'
	fg := termbox.ColorWhite
	bg := termbox.ColorBlack
	if !fieldElement.IsObjectEmpty() {
		switch fieldElement.GetObjectClass() {
		case "hero":
			symbol = '@'
			fg = termbox.ColorMagenta
		case "wall":
			symbol = '#'
			fg = termbox.ColorYellow
		default:
			symbol = '?'
		}
	} else {
		switch fieldElement.GetFloorObjectClass() {
		case "upstairs":
			symbol = '<'
			fg = termbox.ColorGreen
		}
	}
	screenElement.Symbol = symbol
	screenElement.ForegroundColor = fg
	screenElement.BackgroundColor = bg
}

type ScreenText struct {
	Position utils.IMatrixPosition
	// ASCII only. Line breaks are not allowed.
	Text string
	Foreground termbox.Attribute
}

func createSequentialScreenTexts(position utils.IMatrixPosition, parts []*ScreenText) []*ScreenText {
	texts := make([]*ScreenText, 0)
	deltaX := 0
	for _, part := range parts {
		var pos utils.IMatrixPosition = &utils.MatrixPosition{
			Y: position.GetY(),
			X: position.GetX() + deltaX,
		}
		deltaX += len(part.Text)
		fg := termbox.ColorWhite
		if part.Foreground != 0 {
			fg = part.Foreground
		}
		text := ScreenText {
			Position: pos,
			Text: part.Text,
			Foreground: fg,
		}
		texts = append(texts, &text)
	}
	return texts
}

//
// A layer that avoids to write logics tightly coupled with "termbox".
//
type Screen struct {
	matrix [][]ScreenElement
	staticTexts []*ScreenText
}

func (screen *Screen) GetMatrix() [][]ScreenElement {
	return screen.matrix
}

func (screen *Screen) measureRowLength() int {
	return len(screen.matrix)
}

func (screen *Screen) measureColumnLength() int {
	return len(screen.matrix[0])
}

func (screen *Screen) At(position utils.IMatrixPosition) *ScreenElement {
	y := position.GetY()
	x := position.GetX()
	// TODO: Error handling.
	if y < 0 || y > screen.measureRowLength() {
		panic(fmt.Sprintf("That position (Y=%d) does not exist on the screen.", y))
	} else if x < 0 || x > screen.measureColumnLength() {
		panic(fmt.Sprintf("That position (X=%d) does not exist on the screen.", x))
	}
	return &(screen.matrix[y][x])
}

func (screen *Screen) renderField(startPosition utils.IMatrixPosition, field utils.IField) {
	rowLength := field.MeasureRowLength()
	columnLength := field.MeasureColumnLength()
	for y := 0; y < rowLength; y++ {
		for x := 0; x < columnLength; x++ {
			var screenElementPosition utils.IMatrixPosition = &utils.MatrixPosition{
				Y: startPosition.GetY() + y,
				X: startPosition.GetX() + x,
			}
			element := screen.At(screenElementPosition)
			var fieldElementPosition utils.IMatrixPosition = &utils.MatrixPosition{Y: y, X: x}
			// TODO: Error handling.
			var fieldElement, _ = field.At(fieldElementPosition)
			element.renderWithFieldElement(fieldElement)
		}
	}
}

func (screen *Screen) Render(state utils.IState) {
	game := state.GetGame()

	rowLength := screen.measureRowLength()
	columnLength := screen.measureColumnLength()

	// Pad elements with blanks.
	// Set borders.
	for y := 0; y < rowLength; y++ {
		for x := 0; x < columnLength; x++ {
			isTopOrBottomEdge := y == 0 || y == rowLength-1
			isLeftOrRightEdge := x == 0 || x == columnLength-1
			symbol := ' '
			switch {
			case isTopOrBottomEdge && isLeftOrRightEdge:
				symbol = '+'
			case isTopOrBottomEdge && !isLeftOrRightEdge:
				symbol = '-'
			case !isTopOrBottomEdge && isLeftOrRightEdge:
				symbol = '|'
			}
			screen.matrix[y][x].Symbol = symbol
		}
	}

	// Place the field.
	var fieldPosition utils.IMatrixPosition = &utils.MatrixPosition{Y: 2, X: 2}
	screen.renderField(fieldPosition, state.GetField())

	// Prepare texts.
	texts := make([]*ScreenText, 0)
	texts = append(texts, screen.staticTexts...)
	var timeTextPosition utils.IMatrixPosition = &utils.MatrixPosition{Y: 3, X: 25}
	remainingTime := game.CalculateRemainingTime(state.GetExecutionTime()).Seconds()
	remainingTimeText := fmt.Sprintf("%4.1f", remainingTime)
	timeText := ScreenText{
		Position: timeTextPosition,
		Text: fmt.Sprintf("Time : %s", remainingTimeText),
		Foreground: termbox.ColorWhite,
	}
	texts = append(texts, &timeText)
	var floorNumberTextPosition utils.IMatrixPosition = &utils.MatrixPosition{Y: 4, X: 25}
	floorNumberText := ScreenText{
		Position: floorNumberTextPosition,
		Text: fmt.Sprintf("Floor: %2d", game.GetFloorNumber()),
		Foreground: termbox.ColorWhite,
	}
	texts = append(texts, &floorNumberText)
	if game.IsFinished() {
		score := game.GetFloorNumber()
		var lankTextPosition utils.IMatrixPosition = &utils.MatrixPosition{Y: 5, X: 27}
		message := "No good..."
		fg := termbox.ColorWhite
		switch {
			case score == 3:
				message = "Good!"
				fg = termbox.ColorGreen
			case score == 4:
				message = "Excellent!"
				fg = termbox.ColorGreen
			case score == 5:
				message = "Marvelous!"
				fg = termbox.ColorGreen
			case score >= 6:
				message = "Gopher!!"
				fg = termbox.ColorCyan
		}
		lankText := ScreenText{
			Position: lankTextPosition,
			Text: message,
			Foreground: fg,
		}
		texts = append(texts, &lankText)
	}

	// Place texts.
	for _, textInstance := range texts {
		for deltaX, character := range textInstance.Text {
			element := &screen.matrix[textInstance.Position.GetY()][textInstance.Position.GetX() + deltaX]
			element.Symbol = character
			element.ForegroundColor = textInstance.Foreground
		}
	}
}

func (screen *Screen) AsText() string {
	rowLength := screen.measureRowLength()
	columnLength := screen.measureColumnLength()
	lines := make([]string, rowLength)
	for y := 0; y < rowLength; y++ {
		line := make([]rune, columnLength)
		for x := 0; x < columnLength; x++ {
			line[x] = screen.matrix[y][x].Symbol
		}
		lines[y] = string(line)
	}
	return strings.Join(lines, "\n")
}

func CreateScreen(rowLength int, columnLength int) Screen {
	matrix := make([][]ScreenElement, rowLength)
	for rowIndex := 0; rowIndex < rowLength; rowIndex++ {
		row := make([]ScreenElement, columnLength)
		for columnIndex := 0; columnIndex < columnLength; columnIndex++ {
			row[columnIndex] = ScreenElement{
				Symbol:          '_',
				ForegroundColor: termbox.ColorWhite,
				BackgroundColor: termbox.ColorBlack,
			}
		}
		matrix[rowIndex] = row
	}

	staticTexts := make([]*ScreenText, 0)

	var titleTextPosition utils.IMatrixPosition = &utils.MatrixPosition{Y: 0, X: 2}
	titleText := ScreenText{
		Position: titleTextPosition,
		Text: "[ A Tower of Go ]",
		Foreground: termbox.ColorWhite,
	}
	staticTexts = append(staticTexts, &titleText)

	var urlTextPosition utils.IMatrixPosition = &utils.MatrixPosition{Y: 22, X: 41}
	urlText := ScreenText{
		Position: urlTextPosition,
		Text: "https://github.com/kjirou/tower_of_go",
		Foreground: termbox.ColorWhite | termbox.AttrUnderline,
	}
	staticTexts = append(staticTexts, &urlText)

	var operationTitleTextPosition utils.IMatrixPosition = &utils.MatrixPosition{Y: 11, X: 25}
	operationTitleText := ScreenText{
		Position: operationTitleTextPosition,
		Text: "[ Operations ]",
		Foreground: termbox.ColorWhite,
	}
	staticTexts = append(staticTexts, &operationTitleText)

	var sKeyHelpTextPosition utils.IMatrixPosition = &utils.MatrixPosition{Y: 12, X: 25}
	var sKeyHelpTextParts = make([]*ScreenText, 0)
	sKeyHelpTextParts = append(sKeyHelpTextParts, &ScreenText{Text: "\""})
	sKeyHelpTextParts = append(sKeyHelpTextParts, &ScreenText{Text: "s", Foreground: termbox.ColorYellow})
	sKeyHelpTextParts = append(sKeyHelpTextParts, &ScreenText{Text: "\" ... Start or restart a new game."})
	sKeyHelpTexts := createSequentialScreenTexts(sKeyHelpTextPosition, sKeyHelpTextParts)
	staticTexts = append(staticTexts, sKeyHelpTexts...)

	var moveKeysHelpTextPosition utils.IMatrixPosition = &utils.MatrixPosition{Y: 13, X: 25}
	var moveKeysHelpTextParts = make([]*ScreenText, 0)
	moveKeysHelpTextParts =
		append(moveKeysHelpTextParts, &ScreenText{Text: "Arrow keys", Foreground: termbox.ColorYellow})
	moveKeysHelpTextParts = append(moveKeysHelpTextParts, &ScreenText{Text: " or \""})
	moveKeysHelpTextParts =
		append(moveKeysHelpTextParts, &ScreenText{Text: "k,l,j,h", Foreground: termbox.ColorYellow})
	moveKeysHelpTextParts = append(moveKeysHelpTextParts, &ScreenText{Text: "\" ... Move the player."})
	staticTexts = append(
		staticTexts,
		createSequentialScreenTexts(moveKeysHelpTextPosition, moveKeysHelpTextParts)...
	)

	var description1TextPosition utils.IMatrixPosition = &utils.MatrixPosition{Y: 17, X: 3}
	description1Text := ScreenText{
		Position: description1TextPosition,
		Text: "Move the player in the upper left to reach the stairs in the lower right.",
		Foreground: termbox.ColorWhite,
	}
	staticTexts = append(staticTexts, &description1Text)

	var description2TextPosition utils.IMatrixPosition = &utils.MatrixPosition{Y: 18, X: 3}
	description2Text := ScreenText{
		Position: description2TextPosition,
		Text: "The score is the number of floors that can be reached within 30 seconds.",
		Foreground: termbox.ColorWhite,
	}
	staticTexts = append(staticTexts, &description2Text)

	return Screen{
		matrix: matrix,
		staticTexts: staticTexts,
	}
}
