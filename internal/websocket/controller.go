package websocket

import (
	"context"
	"fmt"
	"github.com/olahol/melody"
	"github.com/toujourser/gomoku/internal/constants"
	"github.com/toujourser/gomoku/internal/dto"
	"github.com/toujourser/gomoku/internal/entity"
	"github.com/toujourser/gomoku/internal/service"
	"time"
)

func (ms *MelodySocket) HallChat(ctx context.Context, s *melody.Session, msg *dto.Message) {
	id, _ := s.Get("id")
	content, ok := msg.Data.(string)
	if !ok {
		err := fmt.Errorf("interface conversion: data is not string")
		SendErr(s, err)
		return
	}

	p, err := service.GetPlayer(ctx, id.(string))
	if err != nil {
		SendErr(s, err)
		return
	}

	dialogMsg := &entity.DialogMsg{
		Time:    time.Now().Format("2006-01-02 15:04:05"),
		From:    p.Name,
		Content: content,
	}
	if err := service.HallChat(ctx, dialogMsg); err != nil {
		SendErr(s, err)
		return
	}
	msg = &dto.Message{
		Code: constants.HallChat,
		Data: *dialogMsg,
	}
	ms.Broadcast(msg)
}

func (ms *MelodySocket) GetHallDialog(ctx context.Context, s *melody.Session, msg *dto.Message) {
	dialog, err := service.GetHallDialog(ctx)
	if err != nil {
		SendErr(s, err)
		return
	}
	msg.Data = dialog
	Send(s, msg)
}

func (ms *MelodySocket) GetRooms(ctx context.Context, s *melody.Session, msg *dto.Message) {
	rooms, err := service.GetRooms(ctx)
	if err != nil {
		SendErr(s, err)
		return
	}
	msg.Data = rooms
	Send(s, msg)
}

func (ms *MelodySocket) CreateRoom(ctx context.Context, s *melody.Session, msg *dto.Message) {
	pid, _ := GetPId(s)
	color, ok := msg.Data.(float64)
	if !ok {
		err := fmt.Errorf("interface conversion: data is not a number")
		SendErr(s, err)
		return
	}

	room, err := service.CreateRoom(ctx, pid, int8(color))
	if err != nil {
		SendErr(s, err)
		return
	}

	msg.Data = room
	Send(s, msg)

	rooms, err := service.GetRooms(ctx)
	if err != nil {
		SendErr(s, err)
		return
	}
	ms.Broadcast(&dto.Message{
		Code: constants.GetRooms,
		Data: rooms,
	})
}

func (ms *MelodySocket) EnterRoom(ctx context.Context, s *melody.Session, msg *dto.Message) {
	pid, _ := GetPId(s)

	data := msg.Data.(map[string]interface{})
	rid := data["rid"].(string)
	role := data["role"].(string)

	room, err := service.EnterRoom(ctx, pid, rid, role)
	if err != nil {
		SendErr(s, err)
		return
	}

	msg.Data = room
	ms.Send2Room(room, msg)
}

func (ms *MelodySocket) LeaveRoom(ctx context.Context, s *melody.Session, msg *dto.Message) {
	pid, _ := GetPId(s)

	rid, ok := msg.Data.(string)
	if !ok {
		err := fmt.Errorf("error: data is not string")
		SendErr(s, err)
		return
	}

	ms.SendLeaveRoom(ctx, s, pid, rid)
}

func (ms *MelodySocket) RoomChat(ctx context.Context, s *melody.Session, msg *dto.Message) {
	data := msg.Data.(map[string]interface{})
	from := data["from"].(string)
	content := data["content"].(string)
	rid := data["rid"].(string)

	dialogMsg := &entity.DialogMsg{
		Time:    time.Now().Format("2006-01-02 15:04:05"),
		From:    from,
		Content: content,
	}

	room, err := service.RoomChat(ctx, rid, dialogMsg)
	if err != nil {
		SendErr(s, err)
		return
	}

	roomChatDTO := struct {
		RoomId string `json:"rid"`
		entity.DialogMsg
	}{
		rid,
		*dialogMsg,
	}

	msg.Data = roomChatDTO
	ms.Send2Room(room, msg)
}

func (ms *MelodySocket) GetPlayer(ctx context.Context, s *melody.Session, msg *dto.Message) {
	pid, _ := GetPId(s)
	player, err := service.GetPlayer(ctx, pid)
	if err != nil {
		SendErr(s, err)
		return
	}

	msg.Data = player
	Send(s, msg)
}

func (ms *MelodySocket) GetPlayers(ctx context.Context, s *melody.Session, msg *dto.Message) {
	players, err := service.GetPlayers(ctx)
	if err != nil {
		SendErr(s, err)
		return
	}

	msg.Data = players
	Send(s, msg)
}

func (ms *MelodySocket) PlayerRename(ctx context.Context, s *melody.Session, msg *dto.Message) {
	pid, _ := GetPId(s)
	name, ok := msg.Data.(string)
	if !ok {
		err := fmt.Errorf("interface conversion: data is not string")
		SendErr(s, err)
		return
	}

	if err := service.PlayerRename(ctx, pid, name); err != nil {
		SendErr(s, err)
		return
	}

	msg.Code = constants.GetPlayers
	players, err := service.GetPlayers(ctx)
	if err != nil {
		SendErr(s, err)
		return
	}

	msg.Data = players
	ms.Broadcast(msg)
}

func (ms *MelodySocket) SetPlayerStatus(ctx context.Context, s *melody.Session, msg *dto.Message) {
	pid, _ := GetPId(s)
	status, ok := msg.Data.(string)
	if !ok {
		err := fmt.Errorf("interface conversion: data is not string")
		SendErr(s, err)
		return
	}

	if err := service.SetPlayerStatus(ctx, pid, status); err != nil {
		SendErr(s, err)
		return
	}

	players, err := service.GetPlayers(ctx)
	if err != nil {
		SendErr(s, err)
		return
	}
	msg.Code = constants.GetPlayers
	msg.Data = players
	ms.Broadcast(msg)
}

func (ms *MelodySocket) SetReady(ctx context.Context, s *melody.Session, msg *dto.Message) {
	pid, _ := GetPId(s)
	data := msg.Data.(map[string]interface{})
	rid := data["rid"].(string)
	ready := data["ready"].(bool)

	room, err := service.SetReady(ctx, rid, pid, ready)
	if err != nil {
		SendErr(s, err)
		return
	}

	msg.Data = room
	ms.Send2Room(room, msg)
}

func (ms *MelodySocket) MakeStep(ctx context.Context, s *melody.Session, msg *dto.Message) {
	data := msg.Data.(map[string]interface{})
	rid := data["rid"].(string)
	i := data["i"].(float64)
	j := data["j"].(float64)
	c := entity.Chess{
		I: int8(i),
		J: int8(j),
	}
	over, gameOverDTO, room, err := service.MakeStep(ctx, rid, c)
	if err != nil {
		SendErr(s, err)
		return
	}
	ms.Send2Room(room, msg)
	if over {
		ms.SendGameOver(room, gameOverDTO)
		ms.Send2Room(room, &dto.Message{
			Code: constants.SetReady,
			Data: *room,
		})
	}
}

func (ms *MelodySocket) RetractStep(ctx context.Context, s *melody.Session, msg *dto.Message) {
	pid, _ := GetPId(s)
	data := msg.Data.(map[string]interface{})
	rid := data["rid"].(string)
	consent := int(data["consent"].(float64))
	opponentId, room, count, err := service.RetractStep(ctx, pid, rid, consent)
	if err != nil {
		SendErr(s, err)
	}
	if consent == 2 {
		data["count"] = count
		msg.Data = data
		ms.Send2Room(room, msg)
	} else {
		ms.Send2PId(opponentId, msg)
	}
}

func (ms *MelodySocket) Surrender(ctx context.Context, s *melody.Session, msg *dto.Message) {
	pid, _ := GetPId(s)

	rid, ok := msg.Data.(string)
	if !ok {
		err := fmt.Errorf("error: data is not string")
		SendErr(s, err)
		return
	}

	gameOverDTO, room, err := service.Surrender(ctx, pid, rid)
	if err != nil {
		SendErr(s, err)
		return
	}

	ms.SendGameOver(room, gameOverDTO)
	ms.Send2Room(room, &dto.Message{
		Code: constants.SetReady,
		Data: *room,
	})
}

func (ms *MelodySocket) AskDraw(ctx context.Context, s *melody.Session, msg *dto.Message) {
	pid, _ := GetPId(s)
	data := msg.Data.(map[string]interface{})
	rid := data["rid"].(string)
	consent := int(data["consent"].(float64))
	opponentId, room, err := service.Draw(ctx, pid, rid, consent)
	if err != nil {
		SendErr(s, err)
	}

	if consent == 2 {
		ms.SendGameOver(room, &dto.GameOverDTO{
			RId:   rid,
			Cause: "draw",
		})
		ms.Send2Room(room, &dto.Message{
			Code: constants.SetReady,
			Data: *room,
		})
	} else {
		ms.Send2PId(opponentId, msg)
	}
}
