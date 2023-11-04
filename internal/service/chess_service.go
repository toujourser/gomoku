package service

import (
	"context"
	"fmt"
	"github.com/toujourser/gomoku/internal/constants"
	"github.com/toujourser/gomoku/internal/dto"
	"github.com/toujourser/gomoku/internal/entity"
	"github.com/toujourser/gomoku/internal/lock"
	"github.com/toujourser/gomoku/internal/redis"
	"github.com/toujourser/gomoku/internal/util"
	"github.com/toujourser/gomoku/pkg/logger"
)

func SetReady(ctx context.Context, rid string, pid string, ready bool) (*entity.Room, error) {
	lock.RoomLock.Lock(rid)
	defer lock.RoomLock.Unlock(rid)

	room, err := redis.GetRoom(ctx, rid)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	inRoom, role, _ := isInRoom(pid, room)

	if !inRoom {
		err = fmt.Errorf("error: Player %v not in room %v", pid, rid)
		logger.Error(err)
		return nil, err
	}

	if role == "host" {
		room.Host.Ready = ready
	} else if role == "challenger" {
		room.Challenger.Ready = ready
	} else {
		err = fmt.Errorf("error: Role %v cannot get ready", role)
		logger.Error(err)
		return nil, err
	}
	room.Started = room.Host.Ready && room.Challenger.Ready
	if room.Started {
		room.Steps = make([]entity.Chess, 0)
	}

	if err = redis.SetRoom(ctx, room); err != nil {
		logger.Error(err)
		return nil, err
	}

	return room, nil
}

func MakeStep(ctx context.Context, rid string, c entity.Chess) (bool, *dto.GameOverDTO, *entity.Room, error) {
	lock.RoomLock.Lock(rid)
	defer lock.RoomLock.Unlock(rid)

	room, err := redis.GetRoom(ctx, rid)
	if err != nil {
		logger.Error(err)
		return false, nil, nil, err
	}
	if room.Started {
		room.Steps = append(room.Steps, c)
	} else {
		err = fmt.Errorf("error: Can not make step while game is not started")
		logger.Error(err)
		return false, nil, nil, err
	}

	over, gameOverDTO, err := CheckFive(room)
	if err != nil {
		logger.Error(err)
		return false, nil, nil, err
	}
	if err = redis.SetRoom(ctx, room); err != nil {
		logger.Error(err)
		return false, nil, nil, err
	}
	return over, gameOverDTO, room, nil
}

func PrepareNewGame(room *entity.Room) {
	room.Host.Ready = false
	room.Challenger.Ready = false
	room.Host.Color = 1 - room.Host.Color
	room.Challenger.Color = 1 - room.Challenger.Color
	room.Started = false
}

// CheckFive 检查是否五子连珠
func CheckFive(room *entity.Room) (bool, *dto.GameOverDTO, error) {
	hasFive, color := util.CheckFiveOfLastStep(&room.Steps)
	if !hasFive {
		return false, nil, nil
	}

	var gameOverDTO *dto.GameOverDTO
	// 如果房主的颜色与连珠的颜色相同，则房主胜利， 反之挑战者胜利
	if room.Host.Color == color {
		gameOverDTO = &dto.GameOverDTO{
			RId:    room.Id,
			Winner: room.Host,
			Loser:  room.Challenger,
			Cause:  "five",
		}
	} else {
		gameOverDTO = &dto.GameOverDTO{
			RId:    room.Id,
			Winner: room.Challenger,
			Loser:  room.Host,
			Cause:  "five",
		}
	}

	// 准备新的游戏
	PrepareNewGame(room)
	return true, gameOverDTO, nil
}

// RetractStep 悔棋
func RetractStep(ctx context.Context, pid string, rid string, consent int) (string, *entity.Room, int, error) {
	lock.RoomLock.Lock(rid)
	defer lock.RoomLock.Unlock(rid)

	// 首先，获取房间对象。如果获取失败或者房间未开始或者没有任何棋步，则返回错误。
	room, err := redis.GetRoom(ctx, rid)
	if err != nil {
		logger.Error(err)
		return "", nil, 0, err
	}
	length := len(room.Steps)
	if !room.Started || length == 0 {
		err = fmt.Errorf("error: room %v is not started or there is no step", rid)
		logger.Error(err)
		return "", nil, 0, err
	}

	// 判断当前玩家是否在房间内并且扮演的是参赛者角色。如果不是，则返回错误。
	inRoom, role, _ := isInRoom(pid, room)
	if !inRoom || role == "spectator" {
		err = fmt.Errorf("error: player %v is not playing in room %v", pid, rid)
		logger.Error(err)
		return "", nil, 0, err
	}

	// 根据玩家角色确定对手 ID 和颜色。
	var opponentId string
	var color int8
	if role == "host" {
		opponentId = room.Challenger.Id
		color = room.Challenger.Color
	} else if role == "challenger" {
		opponentId = room.Host.Id
		color = room.Host.Color
	}
	// 如果同意悔棋并且满足一些条件（例如当前玩家不是白方或者当前棋步颜色与当前玩家颜色相同），
	// 则将棋局步骤数组 room.Steps 中的最后一步或最后两步删除，
	// 并设置返回值 count 为 1 或 2，表示悔棋了一步或两步。
	var count int
	if consent == 2 {
		if length == 1 && color == constants.WHITE {
			err = fmt.Errorf("error: there is no white step so white side can't retract")
			logger.Error(err)
			return "", nil, 0, err
		}
		lastColor := int8((length - 1) % 2)
		if lastColor == color {
			count = 1
			room.Steps = room.Steps[:length-1]
		} else {
			count = 2
			room.Steps = room.Steps[:length-2]
		}
	}

	// 更新房间对象并返回结果。
	if err = redis.SetRoom(ctx, room); err != nil {
		logger.Error(err)
		return "", nil, 0, err
	}
	return opponentId, room, count, err
}

// Surrender 投降
func Surrender(ctx context.Context, pid string, rid string) (*dto.GameOverDTO, *entity.Room, error) {
	lock.RoomLock.Lock(rid)
	defer lock.RoomLock.Unlock(rid)

	room, err := redis.GetRoom(ctx, rid)
	if err != nil {
		logger.Error(err)
		return nil, nil, err
	}

	if !room.Started {
		err = fmt.Errorf("error: room %v is not started", rid)
		logger.Error(err)
		return nil, nil, err
	}

	// 判断当前玩家是否在房间内并且扮演的是参赛者角色
	inRoom, role, _ := isInRoom(pid, room)
	if !inRoom || role == "spectator" {
		err = fmt.Errorf("error: player %v is not playing in room %v", pid, rid)
		logger.Error(err)
		return nil, nil, err
	}

	gameOverDTO := &dto.GameOverDTO{
		RId:   rid,
		Cause: "surrender",
	}

	// 根据玩家角色确定胜利方和失败方，并创建一个 GameOverDTO 结构体对象，将房间 ID 和投降原因设置为相应的值。
	if role == "host" {
		gameOverDTO.Winner = room.Challenger
		gameOverDTO.Loser = room.Host
	} else if role == "challenger" {
		gameOverDTO.Winner = room.Host
		gameOverDTO.Loser = room.Challenger
	}

	// 准备新的游戏，即重置房间对象的状态和数据, 更新房间对象并返回 GameOverDTO 结构体指针和房间对象指针。
	PrepareNewGame(room)
	if err = redis.SetRoom(ctx, room); err != nil {
		logger.Error(err)
		return nil, nil, err
	}

	return gameOverDTO, room, nil
}

// Draw 请求平局的功能.
// pid，玩家 ID；rid，房间 ID；consent，一个整数值表示玩家是否同意平局请求（1：拒绝；2：同意）
func Draw(ctx context.Context, pid string, rid string, consent int) (string, *entity.Room, error) {
	lock.RoomLock.Lock(rid)
	defer lock.RoomLock.Unlock(rid)

	room, err := redis.GetRoom(ctx, rid)
	if err != nil {
		logger.Error(err)
		return "", nil, err
	}
	if !room.Started {
		err = fmt.Errorf("error: room %v is not started", rid)
		logger.Error(err)
		return "", nil, err
	}

	inRoom, role, _ := isInRoom(pid, room)
	if !inRoom || role == "spectator" {
		err = fmt.Errorf("error: player %v is not playing in room %v", pid, rid)
		logger.Error(err)
		return "", nil, err
	}

	// 根据玩家是否同意平局请求，分别进行不同的操作。
	// 若玩家同意，则准备新的游戏，即重置房间对象的状态和数据，并更新房间对象。
	// 若玩家拒绝，则不进行任何操作。
	if consent == 2 {
		PrepareNewGame(room)
		if err = redis.SetRoom(ctx, room); err != nil {
			logger.Error(err)
			return "", nil, err
		}
	}

	// 返回对手 ID、房间对象指针和错误。
	var opponentId string
	if role == "host" {
		opponentId = room.Challenger.Id
	} else if role == "challenger" {
		opponentId = room.Host.Id
	}

	return opponentId, room, nil
}
