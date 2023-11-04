package websocket

import (
	"context"
	"encoding/json"
	"github.com/olahol/melody"
	uuid "github.com/satori/go.uuid"
	"github.com/toujourser/gomoku/internal/constants"
	"github.com/toujourser/gomoku/internal/dto"
	"github.com/toujourser/gomoku/internal/entity"
	"github.com/toujourser/gomoku/internal/lock"
	"github.com/toujourser/gomoku/internal/service"
	"github.com/toujourser/gomoku/pkg/logger"
	"sync"
)

//var (
//	idSessionMap sync.Map
//	m            *melody.Melody
//	lock         sync.Mutex
//)

type MelodySocket struct {
	M            *melody.Melody
	idSessionMap sync.Map
	lock         sync.Mutex
}

func (ms *MelodySocket) Receive(s *melody.Session, msgByte []byte) {
	msg := &dto.Message{}
	if err := json.Unmarshal(msgByte, msg); err != nil {
		Send(s, dto.NewErrMsg(err))
	}
	ctx := context.Background()
	switch msg.Code {
	case constants.HallChat:
		ms.HallChat(ctx, s, msg)
	case constants.GetHallDialog:
		ms.GetHallDialog(ctx, s, msg)
	case constants.GetRooms:
		ms.GetRooms(ctx, s, msg)
	case constants.CreateRoom:
		ms.CreateRoom(ctx, s, msg)
	case constants.EnterRoom:
		ms.EnterRoom(ctx, s, msg)
	case constants.LeaveRoom:
		ms.LeaveRoom(ctx, s, msg)
	case constants.RoomChat:
		ms.RoomChat(ctx, s, msg)
	case constants.GetPlayer:
		ms.GetPlayer(ctx, s, msg)
	case constants.GetPlayers:
		ms.GetPlayers(ctx, s, msg)
	case constants.PlayerRename:
		ms.PlayerRename(ctx, s, msg)
	case constants.SetPlayerStatus:
		ms.SetPlayerStatus(ctx, s, msg)
	case constants.SetReady:
		ms.SetReady(ctx, s, msg)
	case constants.MakeStep:
		ms.MakeStep(ctx, s, msg)
	case constants.RetractStep:
		ms.RetractStep(ctx, s, msg)
	case constants.Surrender:
		ms.Surrender(ctx, s, msg)
	case constants.AskDraw:
		ms.AskDraw(ctx, s, msg)
	}
}

func Send(s *melody.Session, msg *dto.Message) {
	msgByte, _ := json.Marshal(msg)
	logger.Debugf("[send msg]: %v", string(msgByte))
	if err := s.Write(msgByte); err != nil {
		logger.Error(err)
	}
}

func SendSuccess(s *melody.Session) {
	Send(s, &dto.Message{
		Code: constants.Success,
		Data: "OK",
	})
}

func SendErr(s *melody.Session, err error) {
	Send(s, dto.NewErrMsg(err))
}

func (ms *MelodySocket) Send2PId(pid string, msg *dto.Message) {
	sObj, ok := ms.idSessionMap.Load(pid)
	if !ok {
		logger.WithField("pid", pid).Error("can not load the value of %v from idSessionMap")
		return
	}
	s, ok := sObj.(*melody.Session)
	if !ok {
		logger.Error("sObj is not type of *melody.Session")
		return
	}
	Send(s, msg)
}

func (ms *MelodySocket) Send2Room(r *entity.Room, msg *dto.Message) {
	if r.Host.Id != "" {
		ms.Send2PId(r.Host.Id, msg)
	}
	if r.Challenger.Id != "" {
		ms.Send2PId(r.Challenger.Id, msg)
	}
	for _, spectator := range r.Spectators {
		ms.Send2PId(spectator.Id, msg)
	}
}

func (ms *MelodySocket) Connect(s *melody.Session) {
	ctx := context.Background()
	id := uuid.NewV4().String()
	player, err := service.NewPlayerConnect(ctx, id)
	if err != nil {
		return
	}
	ms.idSessionMap.Store(id, s)
	s.Set("id", id)
	Send(s, &dto.Message{
		Code: constants.GetPlayer,
		Data: player,
	})
}

func (ms *MelodySocket) Disconnect(s *melody.Session) {
	ctx := context.Background()
	idObject, ok := s.Get("id")
	if !ok {
		logger.Error("session with no 'id' key")
		return
	}
	id := idObject.(string)

	ms.lock.Lock()
	defer ms.lock.Unlock()

	ms.idSessionMap.Delete(id)
	rooms, err := service.PlayerDisconnect(ctx, id)
	if err != nil {
		SendErr(s, err)
	}
	for _, room := range *rooms {
		ms.SendLeaveRoom(ctx, s, id, room.Id)
	}
}

func (ms *MelodySocket) SendGameOver(room *entity.Room, gameOverDTO *dto.GameOverDTO) {
	msg := &dto.Message{
		Code: constants.GameOver,
		Data: *gameOverDTO,
	}

	ms.Send2Room(room, msg)
}

func (ms *MelodySocket) SendLeaveRoom(ctx context.Context, s *melody.Session, pid string, rid string) {
	if !lock.RoomLock.IsExist(rid) {
		return
	}
	room, gameOverDTO, err := service.LeaveRoom(ctx, pid, rid)
	if err != nil {
		SendErr(s, err)
		return
	}

	msg := &dto.Message{
		Code: constants.LeaveRoom,
	}

	if room.Host.Id != "" {
		msg.Data = room
	} else {
		msg.Code = constants.DelRoom
		msg.Data = rid
		if !s.IsClosed() {
			Send(s, msg)
		}

		players, err := service.GetPlayers(ctx)
		if err != nil {
			SendErr(s, err)
		}
		ms.Broadcast(&dto.Message{
			Code: constants.GetPlayers,
			Data: players,
		})
	}
	ms.Send2Room(room, msg)

	if gameOverDTO != nil {
		ms.SendGameOver(room, gameOverDTO)
	}
}

func GetPId(s *melody.Session) (pid string, ok bool) {
	pidObj, ok := s.Get("id")
	if !ok {
		return "", false
	}
	pid, ok = pidObj.(string)
	return
}

func (ms *MelodySocket) Broadcast(msg *dto.Message) {
	msgByte, _ := json.Marshal(*msg)
	if err := ms.M.Broadcast(msgByte); err != nil {
		logger.Error(err)
	}
}
