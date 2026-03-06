package renderer

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/google/uuid"
	"github.com/spf13/afero"
)

// Config provides a config for the template renderer
type (
	Config struct {
		Log            *slog.Logger
		ServiceName    string
		TemplateString string
		Data           any
		// Validate allows the validation of the rendered template on a given temp file path, optional
		Validate func(path string) error
		Fs       afero.Fs
	}

	renderer struct {
		fs          afero.Afero
		log         *slog.Logger
		serviceName string
		tpl         *template.Template
		data        any
		validateFn  func(path string) error
	}
)

// New returns a new template renderer
func New(c *Config) (*renderer, error) {
	tpl, err := template.New("tpl").Funcs(sprig.FuncMap()).Parse(c.TemplateString)
	if err != nil {
		return nil, err
	}

	fs := afero.NewOsFs()
	if c.Fs != nil {
		fs = c.Fs
	}

	return &renderer{
		log:         c.Log.WithGroup("template-renderer"),
		serviceName: c.ServiceName,
		tpl:         tpl,
		data:        c.Data,
		validateFn:  c.Validate,
		fs: afero.Afero{
			Fs: fs,
		},
	}, nil
}

// Render renders the given template to the given destination and reloads the unit if requested.
// Returns true when the template has changed.
func (r *renderer) Render(ctx context.Context, destFile string, reload bool) (changed bool, err error) {
	r.log.Info("rendering template file", "service-name", r.serviceName, "destination", destFile)

	stagingFile := fmt.Sprintf("%s-%s", destFile, uuid.New().String())

	f, err := r.fs.Create(stagingFile)
	if err != nil {
		return false, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			r.log.Error("unable to close file", "error", err)
		}

		if removeErr := r.fs.Remove(stagingFile); removeErr != nil && !os.IsNotExist(removeErr) {
			r.log.Error("unable to remove staging file", "error", removeErr)
			err = removeErr
		}
	}()

	w := bufio.NewWriter(f)

	if err = r.tpl.Execute(w, r.data); err != nil {
		return false, err
	}

	if err = w.Flush(); err != nil {
		return false, err
	}

	if r.validateFn != nil {
		if err := r.validateFn(f.Name()); err != nil {
			return false, err
		}

		r.log.Debug("validated template successfully")
	}

	if equal := r.compare(f.Name(), destFile); equal {
		return false, nil
	}

	if err = r.fs.Rename(f.Name(), destFile); err != nil {
		return false, err
	}

	if !reload {
		return true, nil
	}

	if err := r.reload(ctx); err != nil {
		return true, err
	}

	return true, err
}

func (r *renderer) compare(source, target string) bool {
	sourceChecksum, err := r.checksum(source)
	if err != nil {
		return false
	}

	targetChecksum, err := r.checksum(target)
	if err != nil {
		return false
	}

	return bytes.Equal(sourceChecksum, targetChecksum)
}

func (r *renderer) reload(ctx context.Context) error {
	const done = "done"

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

func (r *renderer) checksum(file string) ([]byte, error) {
	f, err := r.fs.Open(file)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = f.Close()
	}()

	h := sha256.New()

	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
