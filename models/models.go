package models

import (
	"sync"
)

type Ports struct {
	ApplicationPort int `json:"app_port"`
	InternalPort    int `json:"internal_port"`
	HubPort         int `json:"hub_port"`
	Tournament_port int `json:"tournament_port"`
}

type TasksPath struct {
	TasksPath string `json:"task_path"`
}

type StaticFolderPath struct {
	Static string `json:"static_folder_path"`
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
	Users     []int64
	RoomName  string `json:"room_name"`
	CreatorID int64  `json:"creator_id"`
	Mu        sync.Mutex
}

type HubCache struct {
	Rooms        map[int64]*Room
	Join         chan JoinRequest
	Leave        chan LeaveRequest
	Start        chan StartRoomRequest
	Broadcast    chan RoomUpdate
	UserProfiles map[int64]ProfileData
	Mu           sync.Mutex
}

type JoinRequest struct {
	UserID int64
	RoomID int64
}

type UserRequest struct {
	UserID int64 `json:"userID"`
}

type UserResponse struct {
	Username string `json:"username"`
}

type LeaveRequest struct {
	UserID int64
	RoomID int64
}

type StartRoomRequest struct {
	RoomID   int64
	RoomName string
}

type RoomUpdate struct {
	Action string `json:"action"`
	RoomID int64  `json:"roomID"`
	Name   string `json:"roomName"`
}

type RoomMessage struct {
	UserID  int64  `json:"userID"`
	Message string `json:"message"`
}

