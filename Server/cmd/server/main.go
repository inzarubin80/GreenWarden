package main

import (
	"context"
	"fmt"
	"os"
	"github.com/inzarubin80/Server/internal/app"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load("../../.env")
	if err != nil {
		fmt.Println("err godotenv")
	}

	ctx := context.Background()
	// default HTTP port
	options := app.Options{
		Addr:  "0.0.0.0:8090",
	}

	

	conf := app.NewConfig(options)

	databaseUrl := os.Getenv("DATABASE_URL")

	cfg, err := pgxpool.ParseConfig(databaseUrl)
	if err != nil {
		panic(err.Error())

	}
	dbConn, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		panic(err.Error())
	}

	server, err := app.NewApp(ctx, conf, dbConn)
	if err != nil {
		panic(err.Error())
	}

	err = server.ListenAndServe()
	if err != nil {
		fmt.Println(err.Error())
	}

}

