package constants

const (
	Fail = iota
	Success
	HallChat
	GetHallDialog
	GetRooms
	CreateRoom
	EnterRoom
	LeaveRoom
	DelRoom
	RoomChat
	GetPlayer
	GetPlayers
	PlayerRename
	SetPlayerStatus
	SetReady
	MakeStep
	RetractStep // 悔棋
	Surrender   // 投降
	AskDraw     // 求和
	GameOver
)
