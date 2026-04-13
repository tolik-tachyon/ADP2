package main

import (
	"database/sql"
	"log"
	"payment-service/internal/repository"
	paymentHTTP "payment-service/internal/transport/http"
	"payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=Study.ollie dbname=payments_db sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := repository.NewPostgresPaymentRepository(db)
	uc := usecase.NewPaymentUseCase(repo)
	handler := paymentHTTP.NewPaymentHandler(uc)

	r := gin.Default()
	r.POST("/payments", handler.CreatePayment)
	r.GET("/payments/:order_id", handler.GetPayment)

	r.Run(":8081")
}
