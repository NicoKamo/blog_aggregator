package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/Nicokamo/blog_aggregator/internal/database"
)

func handlerBrowse(s *state, cmd command, user database.User) error {
	var limit int32
	var helper int
	var err error
	if len(cmd.CommandSlice) < 1 {
		limit = 2
	} else {
		helper, err = strconv.Atoi(cmd.CommandSlice[0])
		if err != nil {
			return errors.New("Limit provided could not be parsed.")
		} else {
			limit = int32(helper)
		}
	}
	ctx := context.Background()
	posts, err := s.db.GetPostsForUser(ctx, database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return err
	}
	fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("%s from %s\n", post.PublishedAt.Time.Format("Mon Jan 2"), post.FeedName)
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description.String)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}
	return nil
}

