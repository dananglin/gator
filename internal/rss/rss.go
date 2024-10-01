package rss

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

type Feed struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx context.Context, url string) (*Feed, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("received an error creating the HTTP request: %w", err)
	}

	request.Header.Set("User-Agent", "Gator/0.0.0")

	client := http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error getting the response from the server: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return nil, fmt.Errorf(
			"received a bad status from %s: (%d) %s",
			url,
			response.StatusCode,
			response.Status,
		)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to read the response from the server: %w",
			err,
		)
	}

	var feed Feed

	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf(
			"unable to decode the XML data: %w",
			err,
		)
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for _, item := range feed.Channel.Items {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
	}

	return &feed, nil
}
