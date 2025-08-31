package models

type Ports struct {
	ApplicationPort int `json:"app_port"`
}

type Login struct {
	Username  string `json:"username"`
	Passwword string `json:"password"`
}

type DatabasePaths struct {
	UserDatabasePath string `json:"user_database_path"`
}
