package service

import (
	"context"
	"fmt"
	"github.com/toujourser/gomoku/internal/entity"
	"github.com/toujourser/gomoku/internal/redis"
	"github.com/toujourser/gomoku/pkg/logger"
	"time"
)

func NewPlayerConnect(ctx context.Context, id string) (*entity.Player, error) {
	p := &entity.Player{
		Id:            id,
		Name:          "unnamed",
		Status:        "leisure",
		LoginTime:     time.Now().Format(time.DateTime),
		MatchesPlayed: 0,
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
	flag := false
	ps, _ := redis.GetPlayers(ctx)
	for _, op := range *ps {
		if op.Name == newName {
			//return fmt.Errorf("player %v already exists", newName)
			flag = true
			newName += "2"
		}
	}
	p.Name = newName
	err = redis.SetPlayer(ctx, p)
	if err != nil {
		logger.Error(err)
		return err
	}
	if flag {
		return fmt.Errorf("该用户名已存在，系统自动重置名称为 %v", newName)
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
