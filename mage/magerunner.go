package mage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"phaas-localservices-ui/app"
	"strings"
)

type mageRunner struct {
	appSettings *app.Settings
}

var defaultRunner *mageRunner

func Init(
	ctx context.Context,
	appSettings *app.Settings,
) error {
	slog.With(
		slog.String("shellExecutable", appSettings.ShellExecutablePath),
		slog.String("shellInitFile", appSettings.ShellInitFilePath),
	).InfoContext(ctx, "Initializing shell")
	out := bytes.NewBufferString("")
	cmd := exec.Command(appSettings.ShellExecutablePath, "-c", "source", appSettings.ShellInitFilePath)
	cmd.Stdout = out
	cmd.Stderr = out
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to init shell: %w", err)
	}
	slog.With(slog.String("logs", out.String())).InfoContext(ctx, "Shell initialized")
	defaultRunner = &mageRunner{
		appSettings: appSettings,
	}
	return nil
}

var ErrNotInitialized = errors.New("mage package not initialized")

func Exec(ctx context.Context, path string, logTo io.Writer, commands ...string) (*os.Process, error) {
	cmd, err := buildCmd(ctx, path, logTo, commands...)
	if err != nil {
		return nil, fmt.Errorf("unable to build mage command: %w", err)
	}
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start mage command: %w", err)
	}
	return cmd.Process, nil
}

func ExecWait(ctx context.Context, path string, logTo io.Writer, commands ...string) error {
	cmd, err := buildCmd(ctx, path, logTo, commands...)
	if err != nil {
		return fmt.Errorf("unable to build mage command: %w", err)
	}
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start mage command: %w", err)
	}
	return cmd.Run()
}

func buildCmd(ctx context.Context, path string, logTo io.Writer, commands ...string) (*exec.Cmd, error) {
	if defaultRunner == nil {
		return nil, ErrNotInitialized
	}
	cmd := exec.CommandContext(ctx, defaultRunner.appSettings.ShellExecutablePath, "-c", strings.Join(append([]string{"mage"}, commands...), " "))
	cmd.Dir = path
	cmd.Stdout = logTo
	cmd.Stderr = logTo
	cmd.Env = append(cmd.Environ(), "PHAAS_DOCKER_DISABLE_INTERACTIVE=1")
	overrides := defaultRunner.appSettings.GetEnvParamOverrides()
	envParams := make([]string, 0, len(overrides))
	for _, param := range overrides {
		if param.Enabled {
			envParams = append(envParams, fmt.Sprintf("PHAAS_OVERRIDE_%s=%s", strings.ToUpper(param.Key), param.Value))
		}
	}
	cmd.Env = append(cmd.Env, envParams...)
	return cmd, nil
}
