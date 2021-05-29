package m

import (
	"fmt"
	"regexp"

	"github.com/walles/moar/twin"
)

type ScreenLines struct {
	inputLines             *InputLines
	firstInputLineOneBased int
	leftColumnZeroBased    int

	width  int
	height int

	showLineNumbers bool
	wrapLongLines   bool
}

func (sl *ScreenLines) getScreenLines(searchPattern *regexp.Regexp) [][]twin.Cell {
	// Count the length of the last line number
	//
	// Offsets figured out through trial-and-error...
	lastLineOneBased := sl.inputLines.firstLineOneBased + len(sl.inputLines.lines) - 1
	numberPrefixLength := len(formatNumber(uint(lastLineOneBased))) + 1
	if numberPrefixLength < 4 {
		// 4 = space for 3 digits followed by one whitespace
		//
		// https://github.com/walles/moar/issues/38
		numberPrefixLength = 4
	}

	if !sl.showLineNumbers {
		numberPrefixLength = 0
	}

	returnLines := make([][]twin.Cell, 0, sl.height)
	screenFull := false
	for lineIndex, line := range sl.inputLines.lines {
		lineNumber := sl.firstLineOneBased() + lineIndex

		highlighted := line.HighlightedTokens(searchPattern)
		var wrapped [][]twin.Cell
		if sl.wrapLongLines {
			wrapped = wrapLine(sl.width-numberPrefixLength, highlighted)
		} else {
			// All on one line
			wrapped = [][]twin.Cell{highlighted}
		}

		for wrapIndex, inputLinePart := range wrapped {
			visibleLineNumber := &lineNumber
			if wrapIndex > 0 {
				visibleLineNumber = nil
			}

			newLine := make([]twin.Cell, 0, sl.width)
			newLine = append(newLine, createLineNumberPrefix(visibleLineNumber, numberPrefixLength)...)
			newLine = append(newLine, inputLinePart...)

			if sl.leftColumnZeroBased > 0 && len(inputLinePart) > 0 {
				// Add can-scroll-left marker
				newLine[0] = twin.Cell{
					Rune:  '<',
					Style: twin.StyleDefault.WithAttr(twin.AttrReverse),
				}
			}

			if len(inputLinePart)+numberPrefixLength > sl.width {
				newLine[sl.width-1] = twin.Cell{
					Rune:  '>',
					Style: twin.StyleDefault.WithAttr(twin.AttrReverse),
				}
			}

			returnLines = append(returnLines, newLine)

			if len(returnLines) >= sl.height {
				// We have shown all the lines that can fit on the screen
				screenFull = true
				break
			}
		}

		if screenFull {
			break
		}
	}

	return returnLines
}

// Generate a line number prefix. Can be empty or all-whitespace depending on parameters.
func createLineNumberPrefix(fileLineNumber *int, numberPrefixLength int) []twin.Cell {
	if numberPrefixLength == 0 {
		return []twin.Cell{}
	}

	lineNumberPrefix := make([]twin.Cell, 0, numberPrefixLength)
	if fileLineNumber == nil {
		for len(lineNumberPrefix) < numberPrefixLength {
			lineNumberPrefix = append(lineNumberPrefix, twin.Cell{Rune: ' '})
		}
		return lineNumberPrefix
	}

	lineNumberString := formatNumber(uint(*fileLineNumber))
	lineNumberString = fmt.Sprintf("%*s ", numberPrefixLength-1, lineNumberString)
	if len(lineNumberString) > numberPrefixLength {
		panic(fmt.Errorf(
			"lineNumberString <%s> longer than numberPrefixLength %d",
			lineNumberString, numberPrefixLength))
	}

	for column, digit := range lineNumberString {
		if column >= numberPrefixLength {
			break
		}

		lineNumberPrefix = append(lineNumberPrefix, twin.NewCell(digit, _numberStyle))
	}

	return lineNumberPrefix
}

func (sl *ScreenLines) firstLineOneBased() int {
	// FIXME: This is wrong when wrapping is enabled
	return sl.inputLines.firstLineOneBased
}
