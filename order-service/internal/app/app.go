package app

import (
	"database/sql"
	"fmt"
	"order-service/internal/repository"
	httptransport "order-service/internal/transport/http"
	"order-service/internal/transport/httpclient"
	"order-service/internal/usecase"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Config struct {
	Port                   string
	DBDSN                  string
	PaymentServiceGRPCAddr string
}

func NewRouter(db *sql.DB, paymentServiceGRPCAddr string) (*gin.Engine, error) {
	r := gin.Default()

	paymentClient, err := httpclient.NewGRPCPaymentClient(paymentServiceGRPCAddr, 2*time.Second)
	if err != nil {
		return nil, err
	}
	repo := repository.NewPostgresOrderRepository(db)
	uc := usecase.NewOrderUseCase(repo, paymentClient)
	handler := httptransport.NewHandler(uc)
	handler.RegisterRoutes(r)

	return r, nil
}

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func Addr(port string) string {
	return fmt.Sprintf(":%s", port)
}
