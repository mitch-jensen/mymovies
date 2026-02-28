package main

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/jackc/pgx/v5"

	"github.com/spf13/viper"

	db "github.com/mitch-jensen/mymovies/dbstore"
)

func run() error {
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	ctx := context.Background()

	connectionString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", viper.GetString("POSTGRES_USER"), viper.GetString("POSTGRES_PASSWORD"), viper.GetString("POSTGRES_ADDRESS"), viper.GetString("POSTGRES_PORT"), viper.GetString("POSTGRES_DB"))

	conn, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	queries := db.New(conn)

	// list all movies
	movies, err := queries.ListMovies(ctx)
	if err != nil {
		return err
	}
	log.Println(movies)

	// create an movie
	insertedMovie, err := queries.CreateMovie(ctx, db.CreateMovieParams{
		Title:       "The Abominable Dr. Phibes",
		ReleaseYear: 1971,
	})
	if err != nil {
		return err
	}
	log.Println(insertedMovie)

	// get the movie we just inserted
	fetchedMovie, err := queries.GetMovie(ctx, insertedMovie.ID)
	if err != nil {
		return err
	}

	// prints true
	log.Println(reflect.DeepEqual(insertedMovie, fetchedMovie))
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
