package progress

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

type Tracker struct {
	label      string
	total      int64
	current    atomic.Int64
	startedAt  time.Time
	lastBytes  int64
	lastTime   time.Time
	speedEMA   float64
	done       chan struct{}
	finishOnce atomic.Bool
}

func NewTracker(label string, total int64) *Tracker {
	now := time.Now()

	return &Tracker{
		label:     label,
		total:     total,
		startedAt: now,
		lastTime:  now,
		done:      make(chan struct{}),
	}
}

func (t *Tracker) Write(p []byte) (int, error) {
	n := len(p)
	t.current.Add(int64(n))
	return n, nil
}

func (t *Tracker) Start() {
	ticker := time.NewTicker(300 * time.Millisecond)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				t.printLine(false)
			case <-t.done:
				t.printLine(true)
				fmt.Println()
				return
			}
		}
	}()
}

func (t *Tracker) Finish() {
	if t.finishOnce.CompareAndSwap(false, true) {
		close(t.done)
	}
}

func (t *Tracker) printLine(final bool) {
	now := time.Now()
	current := t.current.Load()

	deltaBytes := current - t.lastBytes
	deltaTime := now.Sub(t.lastTime).Seconds()
	if deltaTime <= 0 {
		deltaTime = 0.001
	}

	instantSpeed := float64(deltaBytes) / deltaTime

	if t.speedEMA == 0 {
		t.speedEMA = instantSpeed
	} else {
		// Hafif smoothing: sayı çok zıplamasın ama canlı kalsın
		t.speedEMA = (t.speedEMA * 0.65) + (instantSpeed * 0.35)
	}

	t.lastBytes = current
	t.lastTime = now

	speed := t.speedEMA
	if final && t.total > 0 && current >= t.total {
		// Final satırda hız 0'a düşmesin, son ölçülen hız kalsın
		if speed <= 0 {
			elapsed := now.Sub(t.startedAt).Seconds()
			if elapsed > 0 {
				speed = float64(current) / elapsed
			}
		}
	}

	percent := 0.0
	if t.total > 0 {
		percent = (float64(current) / float64(t.total)) * 100
		if percent > 100 {
			percent = 100
		}
	}

	bar := renderBar(percent, 28)

	etaText := "-"
	if speed > 0 && t.total > 0 && current < t.total {
		remainingBytes := float64(t.total - current)
		etaSeconds := remainingBytes / speed
		etaText = (time.Duration(etaSeconds * float64(time.Second))).Round(time.Second).String()
	}

	line := fmt.Sprintf(
		"\r%s %s %6.2f%%  %s / %s  %s/s  ETA: %s",
		t.label,
		bar,
		percent,
		humanBytes(current),
		humanBytes(t.total),
		humanBytes(int64(speed)),
		etaText,
	)

	if final {
		line = fmt.Sprintf(
			"\r%s %s %6.2f%%  %s / %s  %s/s  ETA: 0s",
			t.label,
			bar,
			percent,
			humanBytes(current),
			humanBytes(t.total),
			humanBytes(int64(speed)),
		)
	}

	fmt.Print(line)
}

func renderBar(percent float64, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int((percent / 100.0) * float64(width))
	if filled > width {
		filled = width
	}

	return "[" + strings.Repeat("=", filled) + strings.Repeat(" ", width-filled) + "]"
}

func humanBytes(n int64) string {
	const unit = 1024

	if n < unit {
		return fmt.Sprintf("%d B", n)
	}

	div, exp := int64(unit), 0
	for value := n / unit; value >= unit; value /= unit {
		div *= unit
		exp++
	}

	suffixes := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.2f %s", float64(n)/float64(div), suffixes[exp])
}
