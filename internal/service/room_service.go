package service

import (
	"context"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/toujourser/gomoku/internal/dto"
	"github.com/toujourser/gomoku/internal/entity"
	"github.com/toujourser/gomoku/internal/lock"
	"github.com/toujourser/gomoku/internal/redis"
	"github.com/toujourser/gomoku/pkg/logger"
)

func GetRooms(ctx context.Context) (*[]entity.Room, error) {
	lock.RoomLock.RLockAll()
	defer lock.RoomLock.RUnlockAll()
	return redis.GetRooms(ctx)
}

func isInRoom(pid string, room *entity.Room) (inRoom bool, role string, index int) {
	if room.Host.Id == pid {
		inRoom = true
		role = "host"
		return
	} else if room.Challenger.Id == pid {
		inRoom = true
		role = "challenger"
		return
	} else if len(room.Spectators) > 0 {
		for i, p := range room.Spectators {
			if p.Id == pid {
				inRoom = true
				role = "spectator"
				index = i
				return
			}
		}
	}
	return
}

func EnterRoom(ctx context.Context, pid string, rid string, role string) (*entity.Room, error) {
	p, err := redis.GetPlayer(ctx, pid)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	lock.RoomLock.Lock(rid)
	defer lock.RoomLock.Unlock(rid)

	r, err := redis.GetRoom(ctx, rid)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if r.Host.Id == "" {
		err = fmt.Errorf("error: Room %v has no host", rid)
		logger.Error(err)
		return nil, err
	}

	if inRoom, _, _ := isInRoom(pid, r); inRoom {
		err = fmt.Errorf("你已经在该房间内了哦")
		logger.Error(err)
		return nil, err
	}

	if role == "challenger" {
		// todo: check if player is already in room
		// 玩家角色状态？
		r.Challenger = entity.PlayerDetails{
			Player: *p,
			Role:   role,
			Color:  1 - r.Host.Color,
			Ready:  false,
		}
	} else if role == "spectator" {
		if p.Status == "leisure" {
			p.Status = "spectating"
			if err := SetPlayerStatus(ctx, pid, "spectating"); err != nil {
				logger.Error(err)
				return nil, err
			}
		}
		r.Spectators = append(r.Spectators, *p)
	} else {
		err = fmt.Errorf("error: The role /'%v/' can't enter room", role)
		logger.Error(err)
		return nil, err
	}

	if err = redis.SetRoom(ctx, r); err != nil {
		return nil, err
	}

	//New player enters room can't see the dialog before
	r.Dialog = make([]entity.DialogMsg, 0)
	return r, nil
}

func CreateRoom(ctx context.Context, pid string, color int8) (*entity.Room, error) {
	p, err := redis.GetPlayer(ctx, pid)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	h := entity.PlayerDetails{
		Player: *p,
		Role:   "host",
		Color:  color,
		Ready:  false,
	}

	id := uuid.NewV4().String()
	lock.RoomLock.Add(id)
	lock.RoomLock.Lock(id)
	defer lock.RoomLock.Unlock(id)

	r := &entity.Room{
		Id:      id,
		Dialog:  make([]entity.DialogMsg, 0),
		Steps:   make([]entity.Chess, 0),
		Started: false,
		Host:    h,
		Challenger: entity.PlayerDetails{
			Color: 1 - color,
		},
		Spectators: make([]entity.Player, 0),
	}

	if err = redis.SetRoom(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

func LeaveRoom(ctx context.Context, pid string, rid string) (*entity.Room, *dto.GameOverDTO, error) {
	lock.RoomLock.Lock(rid)

	r, err := redis.GetRoom(ctx, rid)
	if err != nil {
		logger.Error(err)
		lock.RoomLock.Unlock(rid)
		return nil, nil, err
	}

	inRoom, role, i := isInRoom(pid, r)
	if !inRoom {
		lock.RoomLock.Unlock(rid)
		return r, nil, nil
	}

	var gameOverDTO *dto.GameOverDTO
	if role == "host" {
		if r.Challenger.Id == "" {
			if err = redis.DelRoom(ctx, rid); err != nil {
				logger.Error(err)
				lock.RoomLock.Unlock(rid)
				return nil, nil, err
			}
			r.Host = entity.PlayerDetails{}
			for _, player := range r.Spectators {
				if err := SetPlayerStatus(ctx, player.Id, "leisure"); err != nil {
					logger.Error(err)
					lock.RoomLock.Unlock(rid)
					return nil, nil, err
				}
			}
			lock.RoomLock.Unlock(rid)
			lock.RoomLock.Delete(rid)
			return r, nil, nil
		} else {
			if r.Started {
				r.Started = false
				gameOverDTO = &dto.GameOverDTO{
					RId:    r.Id,
					Winner: r.Challenger,
					Loser:  r.Host,
					Cause:  "escape",
				}
			}
			r.Host = r.Challenger
		}
		r.Challenger = entity.PlayerDetails{}
	} else if role == "challenger" {
		if r.Started {
			r.Started = false
			gameOverDTO = &dto.GameOverDTO{
				RId:    r.Id,
				Winner: r.Host,
				Loser:  r.Challenger,
				Cause:  "escape",
			}
		}
		r.Challenger = entity.PlayerDetails{}
	} else if role == "spectator" {
		r.Spectators = append(r.Spectators[:i], r.Spectators[i+1:]...)
	}

	if err = redis.SetRoom(ctx, r); err != nil {
		logger.Error(err)
		lock.RoomLock.Unlock(rid)
		return nil, nil, err
	}

	lock.RoomLock.Unlock(rid)
	return r, gameOverDTO, nil
}

func RoomChat(ctx context.Context, rid string, msg *entity.DialogMsg) (*entity.Room, error) {
	lock.RoomLock.Lock(rid)
	defer lock.RoomLock.Unlock(rid)

	r, err := redis.GetRoom(ctx, rid)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if len(r.Dialog) >= 10 {
		r.Dialog = r.Dialog[1:]
	}
	r.Dialog = append(r.Dialog, *msg)

	if err = redis.SetRoom(ctx, r); err != nil {
		logger.Error(err)
		return nil, err
	}

	return r, nil
}
