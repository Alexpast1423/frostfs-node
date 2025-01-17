package main

import (
	"context"

	profilerconfig "github.com/TrueCloudLab/frostfs-node/cmd/frostfs-node/config/profiler"
	httputil "github.com/TrueCloudLab/frostfs-node/pkg/util/http"
	"go.uber.org/zap"
)

func initProfiler(c *cfg) {
	if !profilerconfig.Enabled(c.appCfg) {
		c.log.Info("pprof is disabled")
		return
	}

	var prm httputil.Prm

	prm.Address = profilerconfig.Address(c.appCfg)
	prm.Handler = httputil.Handler()

	srv := httputil.New(prm,
		httputil.WithShutdownTimeout(
			profilerconfig.ShutdownTimeout(c.appCfg),
		),
	)

	c.workers = append(c.workers, newWorkerFromFunc(func(context.Context) {
		runAndLog(c, "profiler", false, func(c *cfg) {
			fatalOnErr(srv.Serve())
		})
	}))

	c.closers = append(c.closers, func() {
		c.log.Debug("shutting down profiling service")

		err := srv.Shutdown()
		if err != nil {
			c.log.Debug("could not shutdown pprof server",
				zap.String("error", err.Error()),
			)
		}

		c.log.Debug("profiling service has been stopped")
	})
}
