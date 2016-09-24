package graping

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gizak/termui"
	fastping "github.com/tatsushid/go-fastping"
)

var (
	lc0                    *termui.LineChart
	avgPar                 *termui.Par
	historySize            = 200
	currentlyShowingPoints = 10
)

// App ...
type App struct {
	started       time.Time
	width, height int
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
	return &App{started: time.Now()}
}

// Loop ...
func (app *App) Loop() {

	host := ""
	if len(os.Args) < 2 {
		host = "google.com"
	} else {
		host = os.Args[1]
	}
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	avgPar = termui.NewPar("")
	avgPar.Height = 1
	avgPar.Border = false

	lc0 = termui.NewLineChart()
	lc0.Mode = "dot"
	lc0.BorderLabel = "ping " + host
	lc0.AxesColor = termui.ColorWhite
	lc0.LineColor = termui.ColorYellow

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
				if err = p.Err(); err != nil {
					fmt.Println("Ping failed:", err)
				}
				break loop
			}
		}
		signal.Stop(c)
		p.Stop()
	}()

	termui.Loop()
}
