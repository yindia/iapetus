package bash

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/yindia/iapetus"
	"go.uber.org/zap"
)

func init() {
	iapetus.RegisterBackend("bash", &BashBackend{})
}

type BashBackend struct{}

func (b *BashBackend) ValidateTask(t *iapetus.Task) error {
	return nil
}

func (b *BashBackend) RunTask(t *iapetus.Task) error {
	t.EnsureDefaults()
	ctx, cancel := context.WithTimeout(context.Background(), t.Timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, t.Command, t.Args...)

	// Merge environment variables: os.Environ + t.Env + t.EnvMap (EnvMap takes precedence)
	envMap := map[string]string{}
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	for k, v := range t.EnvMap {
		envMap[k] = v
	}
	finalEnv := make([]string, 0, len(envMap))
	for k, v := range envMap {
		finalEnv = append(finalEnv, k+"="+v)
	}
	cmd.Env = finalEnv

	if t.WorkingDir != "" {
		cmd.Dir = t.WorkingDir
	}
	t.Logger().Debug("Command", zap.String("cmd", t.Command+" "+strings.Join(t.Args, " ")))
	output, err := cmd.CombinedOutput()
	t.Actual.Output = string(output)
	t.Actual.ExitCode = iapetus.GetExitCode(err)
	if err != nil {
		t.Actual.Error = err.Error()
		if ctx.Err() == context.DeadlineExceeded {
			t.Logger().Error("Task timed out", zap.String("task", t.Name), zap.Duration("timeout", t.Timeout))
			return fmt.Errorf("task %s timed out after %v", t.Name, t.Timeout)
		}
		t.Logger().Error("Error executing task", zap.String("task", t.Name), zap.Error(err))
	}
	// Use RunAssertions to aggregate all assertion errors
	err = iapetus.RunAssertions(t)
	if err != nil {
		t.Logger().Error("Assertion(s) failed", zap.String("task", t.Name), zap.Error(err))
		return err
	}
	return nil
}
