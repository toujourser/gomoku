package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/toujourser/gomoku/internal/entity"
	"github.com/toujourser/gomoku/pkg/redis"
)

func AddDialogMsg(ctx context.Context, msg *entity.DialogMsg) error {
	client := redis.RedisClient
	str, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshal msg: %v", err)
	}

	err = client.RPush(ctx, "dialog", str).Err()
	if err != nil {
		return fmt.Errorf("error add dialog msg: %v", err)
	}

	length, err := client.LLen(ctx, "dialog").Result()
	if err != nil {
		return fmt.Errorf("error get dialog length: %v", err)
	}

	if length > 10 {
		_, err = client.LPop(ctx, "dialog").Result()
		if err != nil {
			return fmt.Errorf("error left pop dialog: %v", err)
		}
	}

	return nil
}

func GetDialog(ctx context.Context) (*[]entity.DialogMsg, error) {
	client := redis.RedisClient

	bs, err := client.LRange(ctx, "dialog", 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("error get dialog: %v", err)
	}
	dialog := make([]entity.DialogMsg, 0, 10)
	for _, b := range bs {
		msg := &entity.DialogMsg{}
		if err = json.Unmarshal([]byte(b), msg); err != nil {
			return nil, fmt.Errorf("unmarshal: %v", err)
		}
		dialog = append(dialog, *msg)
	}
	return &dialog, nil
}
