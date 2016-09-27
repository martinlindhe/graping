package graping

import (
	"fmt"
	"time"

	"github.com/gizak/termui"
)

func (app *App) updateUIPositions() {
	app.height = termui.TermHeight()
	app.width = termui.TermWidth()
	app.chart.Width = app.width
	app.chart.Height = app.height
	app.footer.X = app.width - len(app.footer.Text) - 1
	app.footer.Y = app.height - 1
	currentlyShowingPoints = app.chart.Width - 8
	if app.chart.Mode == "braille" {
		currentlyShowingPoints *= 2
	}
}

func (app *App) repaintScreen(history []savedResult) {

	// pos is offset for start of history to plot on screen,
	// wrt to the full history buffer
	pos := len(history) - currentlyShowingPoints
	if pos < 0 {
		pos = 0
	}
	histLen := len(history[pos:])
	var sum time.Duration
	data := make([]float64, histLen)

	avgMeasures := 0
	for i, hist := range history[pos:] {
		if hist.dead {
			data[i] = 0
		} else {
			// in ms
			data[i] = hist.rtt.Seconds() * 1000
			if data[i] > app.max {
				app.max = data[i]
			}
			if data[i] > 0 && data[i] < app.min {
				app.min = data[i]
			}
			sum += hist.rtt
			if data[i] > 0 {
				avgMeasures++
			}
		}
	}
	app.chart.Data = data

	if histLen > 0 {
		txt := ""

		if app.width > 60 {
			//duration := time.Since(app.started)
			//txt += fmt.Sprintf("ran %v, ", duration-(duration%time.Second))
			txt += fmt.Sprintf("%.0f-%.0f ms ", app.min, app.max)
		}

		avg := sum / time.Duration(avgMeasures)
		last := history[len(history)-1]
		txt += fmt.Sprintf("(avg %.0f) ", avg.Seconds()*1000)

		if last.rtt != 0. {
			txt += fmt.Sprintf("now %.0f ms", last.rtt.Seconds()*1000)
		} else {
			txt += "now n/a ms"
		}
		app.footer.Text = txt
		app.footer.X = termui.TermWidth() - len(app.footer.Text) - 2
		app.footer.Width = len(app.footer.Text)
	}

	termui.Render(app.chart, app.footer)
}
