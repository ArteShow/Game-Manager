package tournament

import (
	"net/http"
	"strconv"

	getconfig "github.com/ArteShow/Game-Manager/pkg/getConfig"
)

// Start http Server
func StartTournamentHttp() error {
	port, err := getconfig.GetTournamentPort()
	if err != nil {
		return err
	}
	strport := strconv.Itoa(port)

	return http.ListenAndServe(":"+strport, nil)
}
