package thumb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var (
	ErrUnknownGenerator = errors.New("unknown generator type")
	ErrUnknownOutput    = errors.New("unknown output from generator")
)

// TestGenerator tests thumb generator by getting lib version
func TestGenerator(ctx context.Context, name, executable string) (string, error) {
	switch name {
	case "vips":
		return testVipsGenerator(ctx, executable)
	case "ffmpeg":
		return testFfmpegGenerator(ctx, executable)
	case "libreOffice":
		return testLibreOfficeGenerator(ctx, executable)
	default:
		return "", ErrUnknownGenerator
	}
}

func testVipsGenerator(ctx context.Context, executable string) (string, error) {
	cmd := exec.CommandContext(ctx, executable, "--version")
	var output bytes.Buffer
	cmd.Stdout = &output
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to invoke vips executable: %w", err)
	}

	if !strings.Contains(output.String(), "vips") {
		return "", ErrUnknownOutput
	}

	return output.String(), nil
}

func testFfmpegGenerator(ctx context.Context, executable string) (string, error) {
	cmd := exec.CommandContext(ctx, executable, "-version")
	var output bytes.Buffer
	cmd.Stdout = &output
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to invoke ffmpeg executable: %w", err)
	}

	if !strings.Contains(output.String(), "ffmpeg") {
		return "", ErrUnknownOutput
	}

	return output.String(), nil
}

func testLibreOfficeGenerator(ctx context.Context, executable string) (string, error) {
	cmd := exec.CommandContext(ctx, executable, "--version")
	var output bytes.Buffer
	cmd.Stdout = &output
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to invoke libreoffice executable: %w", err)
	}

	if !strings.Contains(output.String(), "LibreOffice") {
		return "", ErrUnknownOutput
	}

	return output.String(), nil
}
