package webhooksupport

import (
	"context"
	"errors"
	"net/http"
	"time"

	"k8s.io/klog/v2"
)

type ServerConfig struct {
	CertPath string
	KeyPath  string
	Addr     string
}

func Serve(ctx context.Context, cfg ServerConfig, mux *http.ServeMux) error {
	var (
		srv = &http.Server{
			Addr:           cfg.Addr,
			Handler:        mux,
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   5 * time.Second,
			MaxHeaderBytes: 1 << 20, // 1048576

		}
		serveFailed = make(chan error)
	)

	go func() {
		var err error

		if cfg.CertPath == "" || cfg.KeyPath == "" {
			klog.InfoS("Starting INSECURE webhook server over HTTP", "addr", srv.Addr)

			err = srv.ListenAndServe()
		} else {
			klog.InfoS(
				"Starting webhook server over HTTPS",
				"addr",
				srv.Addr,
				"cert_path",
				cfg.CertPath,
				"key_path",
				cfg.KeyPath,
			)

			err = srv.ListenAndServeTLS(cfg.CertPath, cfg.KeyPath)
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			// If the server fails to serve, we need to stop.
			serveFailed <- err
		}
	}()

	select {
	case err := <-serveFailed:
		klog.ErrorS(err, "Server exited reporting an error")
		return err
	case <-ctx.Done():
		klog.Info("Main context exited, gracefully stoping server")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		klog.ErrorS(err, "shutdown reported an error, closing the server")

		_ = srv.Close()
	}

	klog.Info("Webhook server exited")

	return nil
}
