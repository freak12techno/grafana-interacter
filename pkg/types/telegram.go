package types

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jarcoal/httpmock"
)

type TelegramResponse struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func TelegramResponseHasText(text string) httpmock.Matcher {
	return httpmock.NewMatcher(text,
		func(req *http.Request) bool {
			response := TelegramResponse{}
			err := json.NewDecoder(req.Body).Decode(&response)
			if err != nil {
				return false
			}

			if response.Text != text {
				panic(fmt.Sprintf("expected %q but got %q", response.Text, text))
			}

			return true
		})
}
