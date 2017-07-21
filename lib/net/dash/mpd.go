package dash

import (
	"context"
	"encoding/xml"
	"sync"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/net/dash/mpd"
)

func readMPD(ctx context.Context, manifest string) (*mpd.MPD, error) {
	b, err := files.Read(ctx, manifest)
	if err != nil {
		return nil, err
	}

	m := new(mpd.MPD)
	if err := xml.Unmarshal(b, m); err != nil {
		return nil, err
	}

	return m, nil
}

type cachedMPD struct {
	sync.RWMutex

	expTime time.Duration
	expired <-chan struct{}
	err     error

	manifest string
	*mpd.MPD
}

func newMPD(manifest string, minimumUpdatePeriod time.Duration) *cachedMPD {
	expire := make(chan struct{})
	close(expire)

	return &cachedMPD{
		expTime: minimumUpdatePeriod,
		expired: expire,

		manifest: manifest,
		MPD:      new(mpd.MPD),
	}
}

func (m *cachedMPD) refresh(ctx context.Context) (*mpd.MPD, error) {
	m.Lock()
	defer m.Unlock()

	// we weren’t locked before, so we have to check again
	select {
	case <-m.expired:
		// still expired, continuing on to refresh

	case <-ctx.Done():
		return nil, ctx.Err()

	// someone else refreshed the channel while we were waiting…
	// so, we get the return values gratis
	default:
		return m.MPD, m.err
	}

	m.MPD, m.err = readMPD(ctx, m.manifest)
	if m.err != nil {
		return nil, m.err
	}

	expire := make(chan struct{})
	m.expired = expire

	go func() {
		// using closure, no one else can close this channel
		defer close(expire)

		select {
		case <-time.After(m.expTime):
		case <-ctx.Done():
		}
	}()

	return m.MPD, m.err
}

func (m *cachedMPD) get(ctx context.Context) (*mpd.MPD, error) {
	select {
	case <-m.expired:
		return m.refresh(ctx)

	case <-ctx.Done():
		return nil, ctx.Err()
	}

	m.RLock()
	defer m.RUnlock()

	return m.MPD, m.err
}
