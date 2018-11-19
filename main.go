package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"time"

	httpDelivery "github.com/bxcodec/go-clean-arch/v2/article/delivery/http"
	articleRepository "github.com/bxcodec/go-clean-arch/v2/article/repository"
	articleUsecase "github.com/bxcodec/go-clean-arch/v2/article/usecase"
	authorRepository "github.com/bxcodec/go-clean-arch/v2/author/repository"
	"github.com/bxcodec/go-clean-arch/v2/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigFile(`config.json`)
	err := viper.ReadInConfig()

	if err != nil {
		panic(err)
	}

	if viper.GetBool(`debug`) {
		fmt.Println("Service RUN in DEBUG mode")
	}
}

func main() {
	dsn := buildDataSourceName()
	dbConn, err := sql.Open(`mysql`, dsn)
	if err != nil && viper.GetBool("debug") {
		fmt.Println(err)
	}
	err = dbConn.Ping()
	if err != nil {
		log.Fatalf("unable to ping the MySQL db: %+v", err)
	}

	defer func() {
		if err := dbConn.Close(); err != nil {
			log.Fatalf("unable to close MySQL: %+v", err)
		}

	}()
	e := echo.New()
	middL := middleware.InitMiddleware()
	e.Use(middL.CORS)
	authorRepo := authorRepository.NewMysqlAuthorRepository(dbConn)
	articleRepo := articleRepository.NewMysqlArticleRepository(dbConn)

	timeoutContext := time.Duration(viper.GetInt("context.timeout")) * time.Second
	au := articleUsecase.NewArticleUsecase(articleRepo, authorRepo, timeoutContext)
	httpDelivery.NewArticleHttpHandler(e, au)

	if err := e.Start(viper.GetString("server.address")); err != nil {
		log.Fatalf("echo start: %+v", err)
	}
}

func buildDataSourceName() string {
	dbHost := viper.GetString(`database.host`)
	dbPort := viper.GetString(`database.port`)
	dbUser := viper.GetString(`database.user`)
	dbPass := viper.GetString(`database.pass`)
	dbName := viper.GetString(`database.name`)
	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)

	val := url.Values{}
	val.Add("parseTime", "1")
	val.Add("loc", "Asia/Jakarta")
	dsn := fmt.Sprintf("%s?%s", connection, val.Encode())
	return dsn
}
