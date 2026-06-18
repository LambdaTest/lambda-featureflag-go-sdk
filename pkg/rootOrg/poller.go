package rootOrg

import "time"

type poller struct {
	shutdown chan struct{}
}

func newPoller() *poller {
	return &poller{
		shutdown: make(chan struct{}),
	}
}

func (p *poller) Poll(interval time.Duration, function func()) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-p.shutdown:
				ticker.Stop()
				return
			case <-ticker.C:
				go function()
			}
		}
	}()
}
