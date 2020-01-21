package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"path/filepath"
	"time"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
)

func main() {
	app := kingpin.New(filepath.Base(os.Args[0]), "A tool to interact with ChaosMesh")
	app.Version("v0.0.1")
	app.HelpFlag.Short('h')

	nwCmd := app.Command("network", "")
	daemonAddr := nwCmd.Flag("addr", "").Required().String()
	containerID := nwCmd.Flag("id", "").Required().String()

	delayCmd := nwCmd.Command("delay", "delay")
	latency := delayCmd.Arg("latency", "latency").Required().String()
	jitter := delayCmd.Arg("jitter", "jitter").Required().String()
	//fsCmd := app.Command("fs", "")

	parsedCmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	switch parsedCmd {
	case nwCmd.FullCommand():
		c, err := newNetworkDaemonClient(ctx, *daemonAddr)
		delayTime, err := time.ParseDuration(*latency)
		if err != nil {
			logrus.Fatalf("failed to parse latency %v", err)
			os.Exit(1)
		}
		jitterTime, err := time.ParseDuration(*jitter)
		if err != nil {
			logrus.Fatalf("failed to parse jitter %v", err)
			os.Exit(1)
		}

		if _, err := c.SetNetem(ctx, &pb.NetemRequest{
			ContainerId: *containerID,
			Netem: &pb.Netem{
				Time: uint32(delayTime.Nanoseconds() / 1e3),
				Jitter: uint32(jitterTime.Nanoseconds() / 1e3),
			},
		}); err != nil {
			logrus.Fatalf("failed to set netem %v", err)
			os.Exit(1)
		}
	}
}

func newNetworkDaemonClient(ctx context.Context, addr string) (pb.ChaosDaemonClient, error) {
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		logrus.Fatalf("could not connect to network chaos daemon %s: %v", addr, err)
		return nil, err
	}
	return pb.NewChaosDaemonClient(conn), nil
}