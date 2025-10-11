package halloween

import (
	"net/http"
	"strconv"

	GetConfiguration "github.com/ArteShow/Game-Manager/pkg/getConfig"
)

// Start http Server
func StartTournamentHttp() error {
	port, err := GetConfiguration.GetTournamentPort()
	if err != nil {
		return err
	}
	strport := strconv.Itoa(port)

	return http.ListenAndServe(":"+strport, nil)
}
