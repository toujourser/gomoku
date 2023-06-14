package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/toujourser/gomoku/internal/entity"
	"github.com/toujourser/gomoku/pkg/redis"
)

func SetRoom(ctx context.Context, room *entity.Room) error {
	client := redis.RedisClient

	str, err := json.Marshal(room)
	if err != nil {
		return fmt.Errorf("error marshal room: %v", err)
	}

	_, err = client.HSet(ctx, "room", room.Id, str).Result()
	if err != nil {
		return fmt.Errorf("error add player: %v", err)
	}

	return nil
}

func DelRoom(ctx context.Context, id string) error {
	client := redis.RedisClient

	_, err := client.HDel(ctx, "room", id).Result()
	if err != nil {
		return fmt.Errorf("error delete room %v: %v", id, err)
	}

	return nil
}

func GetRoom(ctx context.Context, id string) (*entity.Room, error) {
	client := redis.RedisClient

	b, err := client.HGet(ctx, "room", id).Result()
	if err != nil {
		err = fmt.Errorf("redis: room with id '%v' not found", id)
		return nil, err
	}
	r := &entity.Room{}
	err = json.Unmarshal([]byte(b), r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func GetRooms(ctx context.Context) (*[]entity.Room, error) {
	client := redis.RedisClient

	bs, err := client.HVals(ctx, "room").Result()
	if err != nil {
		return nil, err
	}
	rooms := make([]entity.Room, 0, 100)
	for _, b := range bs {
		r := &entity.Room{}
		err = json.Unmarshal([]byte(b), r)
		if err != nil {
			return nil, fmt.Errorf("unmarshal: %v", err)
		}
		rooms = append(rooms, *r)
	}
	return &rooms, nil
}
