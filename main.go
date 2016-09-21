package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gizak/termui"
	fastping "github.com/tatsushid/go-fastping"
)

type response struct {
	addr *net.IPAddr
	rtt  time.Duration
}

var (
	lc0                    *termui.LineChart
	avgPar                 *termui.Par
	historySize            = 200
	currentlyShowingPoints = 10
)

func updateUIPositions() {
	height := termui.TermHeight()
	width := termui.TermWidth()
	lc0.Width = width
	lc0.Height = height
	avgPar.X = width - len(avgPar.Text) - 1
	avgPar.Y = height - 1
	currentlyShowingPoints = width - 8
}

func main() {
	flag.Parse()

	host := ""
	if len(os.Args) < 2 {
		fmt.Println("host undefined, using google.com")
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
	lc0.BorderLabel = "ping " + host
	lc0.AxesColor = termui.ColorWhite
	lc0.LineColor = termui.ColorYellow

	history := []savedResult{}

	updateUIPositions()
	repaintScreen(history)

	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
	})

	termui.Handle("/sys/wnd/resize", func(termui.Event) {
		updateUIPositions()
		repaintScreen(history)
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
	//signal.Notify(c, syscall.SIGTERM)

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
					repaintScreen(history)
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

type savedResult struct {
	host string
	rtt  time.Duration
	ts   time.Time
	dead bool
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
