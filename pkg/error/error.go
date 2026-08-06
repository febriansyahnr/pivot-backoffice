package errors

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/proto"
)

func New(errType string, err error) error {
	err = fmt.Errorf("%s | %w", errType, err)
	return err
}

func ExtractError(err error) (string, error) {
	extErr := strings.Split(err.Error(), " | ")
	if len(extErr) >= 2 {
		return extErr[0], errors.New(strings.Join(extErr[1:], " | "))
	}
	return "", err
}

func LogProtoMarshalError(ctx context.Context, log logger.ILogger, err error, msg proto.Message) {
	if err == nil {
		return
	}
	log.Warn(ctx, "Proto Marshal Error: "+err.Error(), logger.Any("message", msg))
}
