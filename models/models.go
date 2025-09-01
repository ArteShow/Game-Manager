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
}

type Game struct {
	Name      string `json:"name"`
	UserID    int64  `json:"userID"`
	ProfileID int64  `json:"profile_id"`
}
