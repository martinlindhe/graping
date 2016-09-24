package graping

import (
	"fmt"
	"time"

	"github.com/gizak/termui"
)

func (app *App) updateUIPositions() {
	app.height = termui.TermHeight()
	app.width = termui.TermWidth()
	lc0.Width = app.width
	lc0.Height = app.height
	avgPar.X = app.width - len(avgPar.Text) - 1
	avgPar.Y = app.height - 1
	currentlyShowingPoints = app.width - 8
}

func (app *App) repaintScreen(history []savedResult) {

	pos := len(history) - currentlyShowingPoints
	if pos < 0 {
		pos = 0
	}
	currHistLen := len(history[pos:])
	var sum time.Duration
	data := make([]float64, currHistLen)
	max := 0.
	min := 9999.

	for i, hist := range history[pos:] {
		if hist.dead {
			data[i] = 0
		} else {
			// in ms
			data[i] = hist.rtt.Seconds() * 1000
			if max < data[i] {
				max = data[i]
			}
			if min > data[i] {
				min = data[i]
			}
			sum += hist.rtt
		}
	}
	lc0.Data = data

	if currHistLen > 0 {
		txt := ""

		duration := time.Since(app.started)
		txt += fmt.Sprintf("running %v, ", duration-(duration%time.Second))
		if app.width > 40 {
			txt += fmt.Sprintf("min %1.f ms, ", min)
			txt += fmt.Sprintf("max %.1f ms, ", max)
		}

		avg := sum / time.Duration(currHistLen)
		last := history[len(history)-1]
		txt += fmt.Sprintf("avg %.1f ms, ", avg.Seconds()*1000)

		if last.rtt != 0. {
			txt += fmt.Sprintf("now %.1f ms", last.rtt.Seconds()*1000)
		} else {
			txt += "now n/a ms"
		}
		avgPar.Text = txt
		avgPar.X = termui.TermWidth() - len(avgPar.Text) - 2
		avgPar.Width = len(avgPar.Text)
	}

	termui.Render(lc0, avgPar)
}
