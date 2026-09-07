package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/Nicokamo/blog_aggregator/internal/config"
	"github.com/Nicokamo/blog_aggregator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}
	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dbQueries := database.New(db)
	s := &state{
		db:         dbQueries,
		CfgPointer: &cfg,
	}
	c := &commands{
		MapCommandToHandler: make(map[string]func(*state, command) error),
	}
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", handlerReset)
	c.register("users", handlerUsers)
	c.register("agg", handlerAgg)
	c.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	c.register("feeds", handlerFeeds)
	c.register("follow", middlewareLoggedIn(handlerFollow))
	c.register("following", middlewareLoggedIn(handlerFollowing))
	c.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	c.register("browse", middlewareLoggedIn(handlerBrowse))
	if len(os.Args) < 2 {
		fmt.Println("error: not enough arguments provided")
		os.Exit(1)
	}
	commandName := os.Args[1]
	commandSlice := os.Args[2:]
	cmd := command{
		Name:         commandName,
		CommandSlice: commandSlice,
	}
	err = c.run(s, cmd)
	if err != nil {
		fmt.Println("error occured:", err)
		os.Exit(1)
	}
}
