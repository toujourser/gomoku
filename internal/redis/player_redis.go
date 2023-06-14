package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/toujourser/gomoku/internal/entity"
	"github.com/toujourser/gomoku/pkg/redis"
)

func SetPlayer(ctx context.Context, player *entity.Player) error {
	client := redis.RedisClient

	str, err := json.Marshal(player)
	if err != nil {
		return fmt.Errorf("error marshal player: %v", err)
	}

	err = client.HSet(ctx, "player", player.Id, str).Err()
	if err != nil {
		return fmt.Errorf("error add player: %v", err)
	}

	return nil
}

func GetPlayer(ctx context.Context, id string) (*entity.Player, error) {
	client := redis.RedisClient
	b, err := client.Do(ctx, "HGET", "player", id).Result()

	p := &entity.Player{}
	err = json.Unmarshal([]byte(b.(string)), p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func GetPlayers(ctx context.Context) (*[]entity.Player, error) {
	client := redis.RedisClient

	bs, err := client.HVals(ctx, "player").Result()
	if err != nil {
		return nil, err
	}
	players := make([]entity.Player, 0, 100)
	for _, b := range bs {
		p := &entity.Player{}
		err = json.Unmarshal([]byte(b), p)
		if err != nil {
			return nil, fmt.Errorf("unmarshal: %v", err)
		}
		players = append(players, *p)
	}
	return &players, nil
}

func DelPlayer(ctx context.Context, id string) error {
	client := redis.RedisClient

	_, err := client.HDel(ctx, "player", id).Result()
	if err != nil {
		return fmt.Errorf("error delete player %v: %v", id, err)
	}

	return nil
}
