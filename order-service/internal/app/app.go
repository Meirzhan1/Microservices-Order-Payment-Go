package app

import (
	"database/sql"
	"fmt"
	"order-service/internal/repository"
	httptransport "order-service/internal/transport/http"
	"order-service/internal/transport/http/middleware"
	"order-service/internal/transport/httpclient"
	"order-service/internal/usecase"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Config struct {
	Port                   string
	DBDSN                  string
	PaymentServiceGRPCAddr string
}

func NewRouter(db *sql.DB, paymentServiceGRPCAddr string, redisAddr string) (*gin.Engine, error) {
	r := gin.Default()

	redisCache, err := repository.NewRedisCache(redisAddr)
	if err != nil {
		return nil, err
	}

	r.Use(middleware.RateLimiter(redisCache.Client(), 10, time.Minute))

	ttlMin, _ := strconv.Atoi(os.Getenv("CACHE_TTL_MINUTES"))
	if ttlMin == 0 {
		ttlMin = 5
	}

	paymentClient, err := httpclient.NewGRPCPaymentClient(paymentServiceGRPCAddr, 2*time.Second)
	if err != nil {
		return nil, err
	}
	repo := repository.NewPostgresOrderRepository(db)
	uc := usecase.NewOrderUseCase(repo, paymentClient, redisCache, time.Duration(ttlMin)*time.Minute)
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
