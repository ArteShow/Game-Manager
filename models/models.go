package models

type Ports struct {
	ApplicationPort int `json:"app_port"`
	InternalPort    int `json:"internal_port"`
}

type Login struct {
	Username  string `json:"username"`
	Passwword string `json:"password"`
	UserID    int64  `json:"user_ID"`
}

type DatabasePaths struct {
	UserDatabasePath    string `json:"user_database_path"`
	JWTDatabsePath      string `json:"jwt_database_path"`
	ProfilsDatabasePath string `json:"proflis_database_path"`
}

type ProfileData struct {
	UserID      int64  `json:"userID"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ProfileID   int64  `json:"profile_id"`
	UsedCount   int64  `json:"used_count"`
}

type Game struct {
	Name      string `json:"name"`
	UserID    int64  `json:"userID"`
	ProfileID int64  `json:"profile_id"`
	GameID    int64  `json:"game_id"`
}

type Room struct {
	Users    map[int64]bool
	RoomName string
}

type HubCache struct {
	Rooms map[int64]*Room
	Join  chan JoinRequest
	Leave chan LeaveRequest
	Start chan StartRoomRequest
}

type JoinRequest struct {
	UserID int64
	RoomID int64
}

type LeaveRequest struct {
	UserID int64
	RoomID int64
}

type StartRoomRequest struct {
	RoomID   int64
	RoomName string
}
