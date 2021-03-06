package graping

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gizak/termui"
	fastping "github.com/tatsushid/go-fastping"
)

var (
	historySize            = 200
	currentlyShowingPoints = 10
)

// App ...
type App struct {
	started       time.Time
	chart         *termui.LineChart
	footer        *termui.Par
	width, height int
	max, min      float64
}

type response struct {
	addr *net.IPAddr
	rtt  time.Duration
}

type savedResult struct {
	host string
	rtt  time.Duration
	ts   time.Time
	dead bool
}

// NewApp ...
func NewApp() *App {
	app := &App{
		started: time.Now(),
		max:     0.,
		min:     9999.,
	}

	app.chart = termui.NewLineChart()
	app.chart.AxesColor = termui.ColorWhite
	app.chart.LineColor = termui.ColorGreen
	app.footer = termui.NewPar("")
	app.footer.Height = 1
	app.footer.Border = false

	return app
}

// Loop ...
func (app *App) Loop() {

	host := app.getHost()
	app.chart.BorderLabel = "ping " + host

	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	history := []savedResult{}

	app.updateUIPositions()
	app.repaintScreen(history)

	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
	})

	termui.Handle("/sys/wnd/resize", func(termui.Event) {
		app.updateUIPositions()
		app.repaintScreen(history)
	})

	p := fastping.NewPinger()

	netProto := "ip4:icmp"
	if strings.ContainsAny(host, ":") {
		netProto = "ip6:ipv6-icmp"
	}
	ra, err := net.ResolveIPAddr(netProto, host)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	results := make(map[string]*response)
	results[ra.String()] = nil
	p.AddIPAddr(ra)

	onRecv, onIdle := make(chan *response), make(chan bool)
	p.OnRecv = func(addr *net.IPAddr, t time.Duration) {
		onRecv <- &response{addr: addr, rtt: t}
	}
	p.OnIdle = func() {
		onIdle <- true
	}

	p.MaxRTT = time.Second
	p.RunLoop()

	c := make(chan os.Signal, 1)

	go func() {
	loop:
		for {
			select {
			case <-c:
				fmt.Println("get interrupted")
				break loop
			case r := <-onRecv:
				if _, ok := results[r.addr.String()]; ok {
					results[r.addr.String()] = r
				}
			case <-onIdle:
				matches := 0
				for host, r := range results {
					res := savedResult{
						host: host,
					}
					if r == nil {
						res.dead = true
						fmt.Printf("%s : unreachable %v\n", host, time.Now())
					} else {
						res.rtt = r.rtt
						res.ts = time.Now()
						//					fmt.Printf("%s : %v %v\n", host, r.rtt, time.Now())
					}
					history = append(history, res)
					matches++
				}
				if len(history) > historySize {
					history = history[len(history)-historySize:]
				}
				if matches > 0 {
					app.repaintScreen(history)
				}
			case <-p.Done():
				err = p.Err()
				break loop
			}
		}
		signal.Stop(c)
		p.Stop()

		termui.StopLoop()

		if err != nil {
			// NOTE: due to termui screen clear, this is not always visible
			log.Println("Ping failed:", err)
		}
	}()

	termui.Loop()
}

func (app *App) getHost() string {
	if len(os.Args) < 2 {
		return "google.com"
	}
	return os.Args[1]
}
