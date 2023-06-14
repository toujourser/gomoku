package service

import (
	"context"
	"github.com/toujourser/gomoku/internal/entity"
	"github.com/toujourser/gomoku/internal/redis"
	"github.com/toujourser/gomoku/pkg/logger"
)

func NewPlayerConnect(ctx context.Context, id string) (*entity.Player, error) {
	p := &entity.Player{
		Id:     id,
		Name:   "unnamed",
		Status: "leisure",
	}
	err := redis.SetPlayer(ctx, p)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	logger.WithField("pid", id).Debug("player connects")
	return p, nil
}

func GetPlayer(ctx context.Context, id string) (*entity.Player, error) {
	return redis.GetPlayer(ctx, id)
}

func GetPlayers(ctx context.Context) (*[]entity.Player, error) {
	return redis.GetPlayers(ctx)
}

func PlayerDisconnect(ctx context.Context, id string) (*[]entity.Room, error) {
	err := redis.DelPlayer(ctx, id)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	logger.WithField("pid", id).Debug("player disconnects")
	return redis.GetRooms(ctx)
}

func PlayerRename(ctx context.Context, id string, newName string) error {
	p, err := redis.GetPlayer(ctx, id)
	if err != nil {
		logger.Error(err)
		return err
	}
	p.Name = newName
	err = redis.SetPlayer(ctx, p)
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func SetPlayerStatus(ctx context.Context, id string, status string) error {
	p, err := redis.GetPlayer(ctx, id)
	if err != nil {
		logger.Error(err)
		return err
	}
	p.Status = status
	err = redis.SetPlayer(ctx, p)
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}
