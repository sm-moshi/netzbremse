package collector

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/sm-moshi/netzbremse/internal/config"
	"github.com/sm-moshi/netzbremse/internal/model"
	"github.com/sm-moshi/netzbremse/internal/postgres"
)

func Run(ctx context.Context, cfg config.Measurement) (model.Measurement, error) {
	fields := strings.Fields(cfg.Command)
	if len(fields) == 0 {
		return model.Measurement{}, fmt.Errorf("NETZBREMSE_SPEEDTEST_COMMAND is empty")
	}

	commandCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	cmd := exec.CommandContext(commandCtx, fields[0], fields[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return model.Measurement{}, fmt.Errorf("run %q: %w", cfg.Command, err)
	}

	return postgres.ParseMeasurementPayload(output, time.Now().UTC(), cfg.Endpoint)
}
