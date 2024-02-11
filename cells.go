package main

import (
	"fmt"
	"strings"

	"github.com/tonkeeper/tongo/boc"
)

func fromStdHexString(s string) ([]*boc.Cell, error) {
	lines := strings.Split(s, "\n")
	cells, _, err := recursiveFromStdHexString(0, lines)
	return cells, err
}

func recursiveFromStdHexString(depth int, lines []string) ([]*boc.Cell, int, error) {
	var cells []*boc.Cell
	i := 0
	for ; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}
		currentDepth := reprDepth(lines[i])
		if currentDepth == depth {
			b, err := boc.BitStringFromFiftHex(strings.Trim(lines[i], " }{x"))
			if err != nil {
				return nil, 0, err
			}
			c := boc.NewCellWithBits(*b)
			cells = append(cells, c)
		} else if currentDepth > depth {
			refCell, read, err := recursiveFromStdHexString(currentDepth, lines[i:])
			if err != nil {
				return nil, 0, err
			}
			if len(refCell) > 4 {
				return nil, 0, fmt.Errorf("too many childre cells: %v, want 4", len(refCell))
			}
			for _, c := range refCell {
				cells[len(cells)-1].AddRef(c)
			}
			i += read
		}
	}
	return cells, i, nil
}

func reprDepth(s string) int {
	for i, l := range s {
		if l != ' ' {
			return i
		}
	}
	return len(s)
}
