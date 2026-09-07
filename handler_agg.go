package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/Nicokamo/blog_aggregator/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func scrapeFeeds(s *state) error {
	ctx := context.Background()
	// Get the next Feed based on when it was last updated
	NextFeed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		return err
	}
	currentTime := time.Now()
	arg := database.MarkFeedFetchedParams{
		ID: NextFeed.ID,
		LastFetchedAt: sql.NullTime{
			Time:  currentTime,
			Valid: true,
		},
	}
	// Mark the post as fetched
	err = s.db.MarkFeedFetched(ctx, arg)
	if err != nil {
		return err
	}
	// Get the rssFeed list for the given feed url
	rssFeed, err := fetchFeed(ctx, NextFeed.Url)
	if err != nil {
		return err
	}
	for _, item := range rssFeed.Channel.Item {
		var publishTime time.Time
		var sqlPublishTime sql.NullTime
		publishTime, err := time.Parse(time.RFC1123, item.PubDate)
		if err != nil {
			sqlPublishTime = sql.NullTime{
				Valid: false,
			}
		} else {
			sqlPublishTime = sql.NullTime{
				Time: publishTime,
				Valid: true,
			}
		}
		newPostArg := database.CreatePostParams{
			ID:        uuid.New(),
			CreatedAt: currentTime,
			UpdatedAt: currentTime,
			Title:     item.Title,
			Url:       NextFeed.Url,
			Description: sql.NullString{
				String: item.Description,
				Valid:  true,
			},
			PublishedAt: sqlPublishTime,
			FeedID:      NextFeed.ID,
		}
		_, err = s.db.CreatePost(ctx, newPostArg)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Code != "23505" {
					return pqErr
				}
				return nil
			} else {
				return err
			}
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.CommandSlice) < 1 {
		return errors.New("No time_between_reqs provided")
	}
	timeBetweenReqs, err := time.ParseDuration(cmd.CommandSlice[0])
	if err != nil {
		return err
	}
	ticker := time.NewTicker(timeBetweenReqs)
	fmt.Println("Collecting feeds every ", cmd.CommandSlice[0])
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}

}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var rssFeed RSSFeed
	err = xml.Unmarshal(body, &rssFeed)
	if err != nil {
		return nil, err
	}
	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
	for index, item := range rssFeed.Channel.Item {
		rssFeed.Channel.Item[index].Title = html.UnescapeString(item.Title)
		rssFeed.Channel.Item[index].Description = html.UnescapeString(item.Description)
	}
	return &rssFeed, nil
}
