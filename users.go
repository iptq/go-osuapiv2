package osuapiv2

import "fmt"

func (api *Api) GetUser(userId string) (user User, err error) {
	url := fmt.Sprintf("/users/%s", userId)
	err = api.Request("GET", url, &user)
	if err != nil {
		return
	}

	return
}

func (api *Api) GetUserEvents(userId int, limit int, offset int) (events []Event, err error) {
	url := fmt.Sprintf(
		"/users/%d/recent_activity?limit=%d&offset=%d",
		userId,
		limit,
		offset,
	)
	err = api.Request("GET", url, &events)
	if err != nil {
		return
	}

	return
}
