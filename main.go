// Package main runs a small sample program against the movie database.
package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spf13/viper"

	db "github.com/mitch-jensen/mymovies/dbstore"
)

func run() error {
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.SetConfigFile(".env")

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	ctx := context.Background()

	connectionString := postgresConnectionString()

	conn, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}

	defer func() {
		closeErr := conn.Close(ctx)
		if closeErr != nil {
			log.Printf("close postgres connection: %v", closeErr)
		}
	}()

	queries := db.New(conn)

	// list all movies
	movies, err := queries.ListMovies(ctx)
	if err != nil {
		return fmt.Errorf("list movies: %w", err)
	}

	log.Println(movies)

	// create an movie
	const phibesReleaseYear = 1971

	insertedMovie, err := queries.CreateMovie(ctx, db.CreateMovieParams{
		Title:       "The Abominable Dr. Phibes",
		ReleaseYear: phibesReleaseYear,
		RuntimeMin:  pgtype.Int4{Int32: 0, Valid: false},
	})
	if err != nil {
		return fmt.Errorf("create movie: %w", err)
	}

	log.Println(insertedMovie)

	// get the movie we just inserted
	fetchedMovie, err := queries.GetMovie(ctx, insertedMovie.ID)
	if err != nil {
		return fmt.Errorf("get inserted movie: %w", err)
	}

	// prints true
	log.Println(insertedMovie == fetchedMovie)

	return nil
}

func postgresConnectionString() string {
	hostPort := net.JoinHostPort(viper.GetString("POSTGRES_ADDRESS"), viper.GetString("POSTGRES_PORT"))

	return fmt.Sprintf(
		"postgresql://%s:%s@%s/%s",
		viper.GetString("POSTGRES_USER"),
		viper.GetString("POSTGRES_PASSWORD"),
		hostPort,
		viper.GetString("POSTGRES_DB"),
	)
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
