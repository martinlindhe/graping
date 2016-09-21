package graping

import (
	"fmt"
	"time"

	"github.com/gizak/termui"
)

func updateUIPositions() {
	height := termui.TermHeight()
	width := termui.TermWidth()
	lc0.SetWidth(width)
	lc0.Height = height
	avgPar.X = width - len(avgPar.Text) - 1
	avgPar.Y = height - 1
	currentlyShowingPoints = width - 8
}

func repaintScreen(history []savedResult) {

	pos := len(history) - currentlyShowingPoints
	if pos < 0 {
		pos = 0
	}
	currHistLen := len(history[pos:])
	var sum time.Duration
	data := make([]float64, currHistLen)

	for i, hist := range history[pos:] {
		if hist.dead {
			data[i] = 0
		} else {
			// in ms
			data[i] = hist.rtt.Seconds() * 1000
			sum += hist.rtt
		}
	}
	lc0.Data = data

	if currHistLen > 0 {
		avg := sum / time.Duration(currHistLen)
		curr := history[currHistLen-1]
		txt := fmt.Sprintf("avg(%d) %.1f ms, ", currHistLen, avg.Seconds()*1000)

		if curr.rtt != 0 {
			txt += fmt.Sprintf("now %.1f ms", curr.rtt.Seconds()*1000)
		} else {
			txt += "now n/a ms"
		}
		avgPar.Text = txt
		avgPar.X = termui.TermWidth() - len(avgPar.Text) - 2
		avgPar.Width = len(avgPar.Text)
	}

	termui.Render(lc0, avgPar)
}
