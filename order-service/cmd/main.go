package main

import (
	"database/sql"
	"log"
	"order-service/internal/repository"
	orderHTTP "order-service/internal/transport/http"
	"order-service/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=Study.ollie dbname=orders_db sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := repository.NewPostgresOrderRepository(db)
	uc := usecase.NewOrderUseCase(repo, "http://localhost:8081")
	handler := orderHTTP.NewOrderHandler(uc)

	r := gin.Default()
	r.POST("/orders", handler.CreateOrder)
	r.GET("/orders/:id", handler.GetOrder)
	r.PATCH("/orders/:id/cancel", handler.CancelOrder)

	r.Run(":8080")
}
