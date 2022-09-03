package neurons

import (
	"crypto/md5"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

var possibleStatuses = []string{"online", "away"}

var versionID string

//  from mattermost-server/model/version.go
var versions = []string{
	"6.6.1",
	"6.6.0",
	"6.5.0",
	"6.4.0",
	"6.3.0",
	"6.2.0",
	"6.1.0",
	"6.0.0",
	// "5.39.0",
	// "5.38.0",
	// "5.37.0",
	// "5.36.0",
	// "5.35.0",
	// "5.34.0",
	// "5.33.0",
	// the other old versions are removed
}

// modified from mattermost-server/model/status.go
type status struct {
	// field names must begin with capital letters to be exported for JSON marshalling
	UserId         string `json:"user_id"`
	Status         string `json:"status"`
	Manual         bool   `json:"manual"`
	LastActivityAt int64  `json:"last_activity_at"`
	// ActiveChannel  string `json:"active_channel,omitempty" db:"-"`
	DNDEndTime int64 `json:"dnd_end_time"`
	// PrevStatus     string `json:"-"`
}

// only called once even if imported by multiple
func init() {
	initUtilRand()
	versionID = generateVersionID()
}

func createNRandomStatuses(numberOfStatuses int) []status {
	statuses := make([]status, numberOfStatuses)
	for i := range statuses {
		statuses[i] = status{
			generateID(),
			chooseRandom(possibleStatuses),
			generateRandomBool(),
			generateRandomTimestamp(time.Millisecond),
			0,
		}
	}
	return statuses
}

// TODO:
// func createNStatuses(data []byte, numberOfStatuses int) []status {
// 	encoded := EncryptEncodeZ32(data)
// 	encodedChunks := createNChunks(encoded, numberOfStatuses)
// 	statuses := make([]status, numberOfStatuses)
// 	for i := range statuses {
// 		statuses[i] = status{
// 		}
// 	}
// 	return statuses
// }

func createStatuses(data []byte) []status {
	encoded := EncryptEncodeZ32(data)
	encodedChunks := createChunks(encoded, idLength) // chunks will come out fixed at length 26 except for maybe the last one
	statuses := make([]status, len(encodedChunks))
	for i := range statuses {
		statuses[i] = status{
			encodedChunks[i],
			chooseRandom(possibleStatuses),
			generateRandomBool(),
			generateRandomTimestamp(time.Millisecond),
			0,
		}
	}
	return statuses
}

func decodeCookie(req *http.Request) (string, error) {
	cookies := req.Cookies()
	if len(cookies) == 3 {
		if cookies[0].Name == "MMAUTHTOKEN" &&
			cookies[1].Name == "MMUSERID" &&
			cookies[2].Name == "MMCSRF" {
			var encodedUid strings.Builder
			for _, cookie := range cookies {
				encodedUid.WriteString(cookie.Value)
			}
			uid, err := Decode32(encodedUid.String())
			if err != nil {
				return "", err
			}
			return string(uid), nil
		}
	}
	return "", errors.New("Incorrect cookie format")
}

// adapted from mattermost/pillar/utils/utils.go and mattermost-server/model/utils.go
func generateID() string {
	UUID, err := generateUUID()
	if err != nil {
		log.Println(err)
		return ""
	}
	encoded := EncodeZ32(UUID)
	// we can safely truncate it this way because we know unicode runes won't be involved
	return encoded[:idLength]
}

// adapted from mattermost-server/web/handlers.go and mattermost-server/app/config.go -> regenerateClientConfig()
// this versionID should be randomly generated only once per running instance
func generateVersionID() string {
	// clientConfigHash is generated as follows:
	// clientConfig := config.GenerateClientConfig(ch.cfgSvc.Config(), ch.srv.TelemetryId(), ch.srv.License())
	// clientConfigJSON, _ := json.Marshal(clientConfig)
	// ch.clientConfigHash.Store(fmt.Sprintf("%x", md5.Sum(clientConfigJSON)))
	clientConfigHash := fmt.Sprintf("%x", md5.Sum(getRandBytes(192)))

	// model.BuildNumber seems to come from the Makefile, and it can be "dev" or "xxx-rc" sometimes it seems
	return fmt.Sprintf("%v.%v.%v.%v", versions[0], versions[0] /* model.BuildNumber */, clientConfigHash, generateRandomBool())
}

func generateLastModifiedDate() string {
	// RFC1123 requires that the date format uses GMT so we just use http.TimeFormat here
	return time.Unix(generateRandomTimestamp(time.Second), 0).Format(http.TimeFormat)
}
