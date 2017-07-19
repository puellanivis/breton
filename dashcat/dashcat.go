package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	httpfiles "github.com/puellanivis/breton/lib/files/http"
	_ "github.com/puellanivis/breton/lib/files/plugins"
	"github.com/puellanivis/breton/lib/flag"
	"github.com/puellanivis/breton/lib/log"
	_ "github.com/puellanivis/breton/lib/metrics/http"
	"github.com/puellanivis/breton/lib/net/dash"
	"github.com/puellanivis/breton/lib/util"
)

var mimeTypes []string

var (
	_ = flag.FuncWithArg("mime-type", "which mime-type(s) to stream",
		func(s string) error {
			mimeTypes = append(mimeTypes, s)
			return nil
		}, flag.WithShort('t'))

	play     = flag.Bool("play", false, "start a subprocess to pipe the output to (currently only mpv)")
	metrics  = flag.Bool("metrics", false, "listens on a random port to report metrics", flag.WithDefault(true))
	quiet = flag.Bool("quiet", false, "surpresses unnecessary output", flag.WithShort('q'))
)

var stderr = os.Stderr

func main() {
	defer util.Init("dash-cat", 0, 1)()

	ctx := util.Context()
	ctx = httpfiles.WithUserAgent(ctx, "dashcat/1.0")

	args := flag.Args()
	if len(args) < 1 {
		util.Statusln(flag.Usage)
		return
	}

	if *quiet {
		stderr = nil
	}

	if *metrics {
		go func() {
			l, err := net.Listen("tcp", ":0")
			if err != nil {
				util.Statusln("failed to establish listener", err)
				return
			}

			util.Statusln("listening on:", l.Addr())
			log.Infoln("listening on:", l.Addr())

			go func() {
				<-ctx.Done()
				l.Close()
			}()

			if err := http.Serve(l, nil); err != nil {
				util.Statusln(err)
			}
		}()
	}

	done := make(chan struct{})
	if !*play {
		// close done, because there will be no subprocess
		close(done)
	}
	defer func() {
		// without doing this last, we will not close the
		// subprocess pipe before blocking on done.
		<-done
	}()

	if len(mimeTypes) < 1 {
		mimeTypes = append(mimeTypes, "video/mp4")
	}

	var out io.Writer = os.Stdout

	if *play {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)

		mpv, err := exec.LookPath("mpv")
		if err != nil {
			log.Fatal(err)
		}

		cmd := exec.CommandContext(ctx, mpv, "-")

		pipe, err := cmd.StdinPipe()
		if err != nil {
			log.Fatal(err)
		}
		defer pipe.Close()
		out = pipe

		cmd.Stdout = os.Stdout
		cmd.Stderr = stderr

		if err := cmd.Start(); err != nil {
			log.Error(err)
		}

		go func() {
			defer close(done)
			defer cancel()

			if err := cmd.Wait(); err != nil {
				log.Error(err)
			}

			util.Statusln("subprocess quit")
		}()

	}

	for _, arg := range args {
		if err := maybeMUX(ctx, out, arg); err != nil {
			log.Error(err)
		}
	}
}

func maybeMUX(ctx context.Context, out io.Writer, arg string) error {
	mpd, err := dash.New(ctx, arg)
	if err != nil {
		return err
	}

	if len(mimeTypes) == 1 {
		return stream(ctx, out, mpd, mimeTypes[0])
	}

	ffmpegArgs := []string{
		"-nostdin",
	}

	for i, _ := range mimeTypes {
		ffmpegArgs = append(ffmpegArgs, "-i", fmt.Sprintf("/dev/fd/%d", 3+i))
	}

	ffmpegArgs = append(ffmpegArgs,
		"-c", "copy",
		"-copyts",
		"-movflags", "frag_keyframe+empty_moov",
		"-f", "mp4",
		"-",
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if log.V(5) {
		log.Info("ffmpeg", ffmpegArgs)
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", ffmpegArgs...)
	cmd.Stdout = out
	cmd.Stderr = stderr

	for _, mimeType := range mimeTypes {
		rd, wr, err := os.Pipe()
		if err != nil {
			return err
		}

		cmd.ExtraFiles = append(cmd.ExtraFiles, rd)

		mimeType := mimeType

		go func() {
			defer func() {	
				if err := wr.Close(); err != nil {
					log.Error(err)
				}
			}()

			if err := stream(ctx, wr, mpd, mimeType); err != nil {
				log.Error(err)
				cancel()
			}
		}()
	}

	return cmd.Run()
}

func stream(ctx context.Context, out io.Writer, mpd *dash.Manifest, mimeType string) error {
	s, err := mpd.Stream(out, mimeType, dash.PickHighestBandwidth)
	if err != nil {
		return err
	}

	if err := s.Init(ctx); err != nil {
		return err
	}

	var totalDuration time.Duration

	// we will later divide this duration by 2 below, to keep it the right
	// value to ensure we don’t update too often.
	minDuration := mpd.MinimumUpdatePeriod() * 2

readLoop:
	for {
		duration, err := mpd.Pull(ctx, s)
		totalDuration += duration

		if err != nil {
			if err != io.EOF {
				log.Error(err)
			}

			break
		}

		if duration > 0 {
			if log.V(1) {
				util.Statusln("segments had a duration of:", duration)
			}
		}

		if duration < minDuration {
			duration = minDuration
		}

		select {
		case <-time.After(duration / 2):
		case <-ctx.Done():
			break readLoop
		}
	}

	util.Statusln("total duration:", totalDuration)
	return nil
}
