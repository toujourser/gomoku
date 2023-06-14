package service

import (
	"context"
	"github.com/toujourser/gomoku/internal/entity"
	"github.com/toujourser/gomoku/internal/redis"
	"github.com/toujourser/gomoku/pkg/logger"
)

func HallChat(ctx context.Context, msg *entity.DialogMsg) error {
	err := redis.AddDialogMsg(ctx, msg)
	if err != nil {
		logger.Error(err)
	}

	return err
}

func GetHallDialog(ctx context.Context) (*[]entity.DialogMsg, error) {
	dialog, err := redis.GetDialog(ctx)
	if err != nil {
		logger.Error(err)
	}
	return dialog, err
}
