package thumb

import (
	"context"
	"errors"
	"fmt"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
)

// Generator generates a thumbnail for a given reader.
type Generator interface {
	// Generate generates a thumbnail for a given reader. Src is the original file path, only provided
	// for local policy files.
	Generate(ctx context.Context, file io.Reader, src, url, name string, options map[string]string) (*Result, error)

	// Priority of execution order, smaller value means higher priority.
	Priority() int

	// EnableFlag returns the setting name to enable this generator.
	EnableFlag() string
}

type Result struct {
	Path     string
	Continue bool
	Cleanup  []func()
}

type (
	GeneratorType string
	GeneratorList []Generator
)

var (
	Generators = GeneratorList{}

	ErrPassThrough  = errors.New("pass through")
	ErrNotAvailable = fmt.Errorf("thumbnail not available: %w", ErrPassThrough)
)

func (g GeneratorList) Len() int {
	return len(g)
}

func (g GeneratorList) Less(i, j int) bool {
	return g[i].Priority() < g[j].Priority()
}

func (g GeneratorList) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

// RegisterGenerator registers a thumbnail generator.
func RegisterGenerator(generator Generator) {
	Generators = append(Generators, generator)
	sort.Sort(Generators)
}

func (p GeneratorList) Generate(ctx context.Context, fileReader io.Reader, src, url string, file *model.File, options map[string]string) (*Result, error) {
	inputFile, inputSrc, inputUrl, inputName := fileReader, src, url, file.Name
	for _, generator := range p {
		if model.IsTrueVal(options[generator.EnableFlag()]) {
			if generator.EnableFlag() == "thumb_ffmpeg_enabled" &&
				file.Size > uint64(model.GetIntSetting("thumb_max_src_size", 31457280)) {
				util.Log().Debug("Generator %s failed to generate thumbnail for %s since file too large, passing through to next generator.", reflect.TypeOf(generator).String(), inputName)
				continue
			}

			res, err := generator.Generate(ctx, inputFile, inputSrc, inputUrl, inputName, options)
			if errors.Is(err, ErrPassThrough) {
				util.Log().Debug("Failed to generate thumbnail using %s for %s: %s, passing through to next generator.", reflect.TypeOf(generator).String(), inputName, err)
				continue
			}

			if res != nil && res.Continue {
				util.Log().Debug("Generator %s for %s returned continue, passing through to next generator.", reflect.TypeOf(generator).String(), inputName)

				// defer cleanup funcs
				for _, cleanup := range res.Cleanup {
					defer cleanup()
				}

				// prepare file reader for next generator
				intermediate, err := os.Open(res.Path)
				if err != nil {
					return nil, fmt.Errorf("failed to open intermediate thumb file: %w", err)
				}

				defer intermediate.Close()
				inputFile = intermediate
				inputSrc = res.Path
				inputName = filepath.Base(res.Path)
				continue
			}

			return res, err
		}
	}
	return nil, ErrNotAvailable
}

func (p GeneratorList) Priority() int {
	return 0
}

func (p GeneratorList) EnableFlag() string {
	return ""
}

func thumbSize(options map[string]string) (uint, uint) {
	w, h := uint(400), uint(300)
	if wParsed, err := strconv.Atoi(options["thumb_width"]); err == nil {
		w = uint(wParsed)
	}

	if hParsed, err := strconv.Atoi(options["thumb_height"]); err == nil {
		h = uint(hParsed)
	}

	return w, h
}
