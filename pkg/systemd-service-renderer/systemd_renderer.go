package systemd_renderer

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/coreos/go-systemd/v22/dbus"
	renderer "github.com/metal-stack/os-installer/pkg/template-renderer"
	"github.com/spf13/afero"
)

type (
	Config struct {
		ServiceName    string
		Log            *slog.Logger
		TemplateString string
		Data           any
		// Validate allows the validation of the rendered template on a given temp file path, optional
		Validate func(path string) error
		Fs       afero.Fs
	}

	systemdRenderer struct {
		log         *slog.Logger
		r           *renderer.Renderer
		serviceName string
	}
)

// New returns a new system service renderer
func New(c *Config) (*systemdRenderer, error) {
	if c == nil {
		return nil, fmt.Errorf("systemd service renderer config is nil")
	}

	r, err := renderer.New(&renderer.Config{
		Log:            c.Log.With("service-name", c.ServiceName),
		TemplateString: c.TemplateString,
		Data:           c.Data,
		Validate:       c.Validate,
		Fs:             c.Fs,
	})
	if err != nil {
		return nil, err
	}

	return &systemdRenderer{
		log:         c.Log.WithGroup("systemd-service-renderer").With("service-name", c.ServiceName),
		serviceName: c.ServiceName,
		r:           r,
	}, nil
}

// Render renders the given template to the given destination and reloads the unit if requested.
// Returns true when the template has changed.
func (r *systemdRenderer) Render(ctx context.Context, destFile string, reload bool) (changed bool, err error) {
	r.log.Info("rendering systemd service template file")

	changed, err = r.r.Render(ctx, destFile)
	if err != nil {
		return changed, err
	}

	if !reload {
		return changed, nil
	}

	if err := r.reload(ctx); err != nil {
		return true, err
	}

	return true, err
}

func (r *systemdRenderer) reload(ctx context.Context) error {
	const done = "done"

	r.log.Info("reloading systemd service unit")

	dbc, err := dbus.NewWithContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to connect to dbus: %w", err)
	}
	defer dbc.Close()

	c := make(chan string)

	if _, err = dbc.ReloadUnitContext(ctx, r.serviceName, "replace", c); err != nil {
		return err
	}

	job := <-c

	if job != done {
		return fmt.Errorf("reloading failed: %s", job)
	}

	return nil
}
