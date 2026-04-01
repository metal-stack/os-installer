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
	"github.com/google/uuid"
	"github.com/spf13/afero"
)

// Config provides a config for the template Renderer
type (
	Config struct {
		Log            *slog.Logger
		TemplateString string
		Data           any
		// Validate allows the validation of the rendered template on a given temp file path, optional
		Validate func(path string) error
		// An optional prefix when creating tmp files
		TmpFilePrefix string
		Fs            afero.Fs
	}

	Renderer struct {
		fs         afero.Afero
		log        *slog.Logger
		tpl        *template.Template
		prefix     string
		data       any
		validateFn func(path string) error
	}
)

// New returns a new template renderer
func New(c *Config) (*Renderer, error) {
	if c == nil {
		return nil, fmt.Errorf("renderer config is nil")
	}

	tpl, err := template.New("tpl").Funcs(sprig.FuncMap()).Parse(c.TemplateString)
	if err != nil {
		return nil, err
	}

	fs := afero.NewOsFs()
	if c.Fs != nil {
		fs = c.Fs
	}

	return &Renderer{
		log:        c.Log.WithGroup("template-renderer"),
		tpl:        tpl,
		data:       c.Data,
		validateFn: c.Validate,
		fs: afero.Afero{
			Fs: fs,
		},
		prefix: c.TmpFilePrefix,
	}, nil
}

// Render renders the given template to the given destination.
// Returns true when the template has changed.
func (r *Renderer) Render(ctx context.Context, destFile string) (changed bool, err error) {
	r.log.Info("rendering template file", "destination", destFile, "data", r.data)

	stagingFile := fmt.Sprintf("%s%s-%s", r.prefix, destFile, uuid.New().String())

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

	return true, err
}

func (r *Renderer) compare(source, target string) bool {
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

func (r *Renderer) checksum(file string) ([]byte, error) {
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
