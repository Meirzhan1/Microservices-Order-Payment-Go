package app

import (
	"database/sql"
	"fmt"
	"payment-service/internal/repository"
	httptransport "payment-service/internal/transport/http"
	"payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Config struct {
	Port  string
	DBDSN string
}

func NewRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()
	repo := repository.NewPostgresPaymentRepository(db)
	uc := usecase.NewPaymentUseCase(repo)
	handler := httptransport.NewHandler(uc)
	handler.RegisterRoutes(r)
	return r
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
