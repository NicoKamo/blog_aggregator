package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Nicokamo/blog_aggregator/internal/database"
	"github.com/google/uuid"
)

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.CommandSlice) < 2 {
		return errors.New("Provide both name and url of the to be added feed")
	}
	name := cmd.CommandSlice[0]
	url := cmd.CommandSlice[1]
	ctx := context.Background()
	newID := uuid.New()
	current_time := time.Now()
	arg := database.CreateFeedParams{
		ID:        newID,
		CreatedAt: current_time,
		UpdatedAt: current_time,
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}
	feed, err := s.db.CreateFeed(ctx, arg)
	if err != nil {
		return err
	}
	newArg := database.CreateFeedFollowParams{
		ID:        newID,
		CreatedAt: current_time,
		UpdatedAt: current_time,
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	_, err = s.db.CreateFeedFollow(ctx, newArg)
	if err != nil {
		return err
	}
	printFeed(feed, user)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	ctx := context.Background()
	feeds, err := s.db.AllFeeds(ctx)
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		fmt.Println("Name: ", feed.Name)
		fmt.Println("URL: ", feed.Url)
		userName, err := s.db.GetUserID(ctx, feed.UserID)
		if err != nil {
			return err
		}
		fmt.Println("Username: ", userName)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.CommandSlice) == 0 {
		return errors.New("No url provided")
	}
	ctx := context.Background()
	url := cmd.CommandSlice[0]
	feed, err := s.db.GetFeedFromUrl(ctx, url)
	if err != nil {
		return err
	}
	var newFeedFollow database.CreateFeedFollowRow
	currentTime := time.Now()
	newID := uuid.New()
	arg := database.CreateFeedFollowParams{
		ID:        newID,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	newFeedFollow, err = s.db.CreateFeedFollow(ctx, arg)
	if err != nil {
		return err
	}
	fmt.Println(newFeedFollow, user.Name)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	ctx := context.Background()
	feedFollowers, err := s.db.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		return err
	}
	for _, feedFollower := range feedFollowers {
		fmt.Println("Feed name: ", feedFollower.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.CommandSlice) < 1 {
		return errors.New("no Url provided")
	}
	url := cmd.CommandSlice[0]
	ctx := context.Background()
	feed, err := s.db.GetFeedFromUrl(ctx, url)
	if err != nil {
		return err
	}
	arg := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	err = s.db.DeleteFeedFollow(ctx, arg)
	if err != nil {
		return err
	}
	fmt.Printf("User %v unfollowed from the feed %s\n", user.Name, feed.Name)
	return nil
}


func printFeed(feed database.Feed, user database.User) {
	fmt.Printf("* ID:            %s\n", feed.ID)
	fmt.Printf("* Created:       %v\n", feed.CreatedAt)
	fmt.Printf("* Updated:       %v\n", feed.UpdatedAt)
	fmt.Printf("* Name:          %s\n", feed.Name)
	fmt.Printf("* URL:           %s\n", feed.Url)
	fmt.Printf("* User:          %s\n", user.Name)
}
