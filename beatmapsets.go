package osuapiv2

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
)

func (api *Api) SearchBeatmaps(rankStatus string) (beatmapSearch BeatmapSearch, err error) {
	values := url.Values{}
	values.Set("s", rankStatus)
	query := values.Encode()
	url := "/beatmapsets/search?" + query
	err = api.Request("GET", url, &beatmapSearch)
	if err != nil {
		return
	}

	return
}

func (api *Api) DownloadSingleBeatmap(beatmapId int, path string) (err error) {
	url := fmt.Sprintf("https://osu.ppy.sh/osu/%d", beatmapId)
	resp, err := api.httpClient.Get(url)

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return
	}
	return
}

func (api *Api) GetBeatmapSet(beatmapSetId int) (beatmapSet Beatmapset, err error) {
	url := fmt.Sprintf("/beatmapsets/%d", beatmapSetId)
	err = api.Request("GET", url, &beatmapSet)
	if err != nil {
		return
	}

	return
}

func (api *Api) BeatmapsetDownload(beatmapSetId int) (path string, err error) {
	url := fmt.Sprintf("/beatmapsets/%d/download", beatmapSetId)
	resp, err := api.Request0("GET", url)
	if err != nil {
		return
	}

	file, err := ioutil.TempFile(os.TempDir(), "beatmapsetDownload")
	if err != nil {
		return
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return
	}
	file.Close()

	path = file.Name()
	return
}

type GetBeatmapsetEventsOptions struct {
	User  string
	Types []string
}

func (api *Api) GetBeatmapsetEvents(opts *GetBeatmapsetEventsOptions) (events []BeatmapsetEvent, err error) {
	values := url.Values{}
	values.Set("user", opts.User)
	query := values.Encode()
	for _, t := range opts.Types {
		query += "&types[]=" + t
	}
	url := "/beatmapsets/events?" + query
	fmt.Println("URL IS", url)

	var reply BeatmapsetEvents
	err = api.Request("GET", url, &reply)
	if err != nil {
		return
	}

	events = reply.Events
	return
}
