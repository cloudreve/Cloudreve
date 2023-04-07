package thumb

import (
	"errors"
	"fmt"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"io"
	"sort"
	"strconv"
)

// Generator generates a thumbnail for a given reader.
type Generator interface {
	Generate(file io.Reader, w io.Writer, name string, options map[string]string) error

	// Priority of execution order, smaller value means higher priority.
	Priority() int

	// EnableFlag returns the setting name to enable this generator.
	EnableFlag() string
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

func (p GeneratorList) Generate(file io.Reader, w io.Writer, name string, options map[string]string) error {
	for _, generator := range p {
		if model.IsTrueVal(options[generator.EnableFlag()]) {
			err := generator.Generate(file, w, name, options)
			if errors.Is(err, ErrPassThrough) {
				util.Log().Debug("Failed to generate thumbnail for %s: %s, passing through to next generator.", name, err)
				continue
			}

			return err
		}
	}
	return ErrNotAvailable
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
