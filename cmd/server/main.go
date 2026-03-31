package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	_ "github.com/modami/be-payment-service/docs/swagger"
	"github.com/modami/be-payment-service/config"
	"github.com/modami/be-payment-service/module/core/repository/postgres"
	"github.com/modami/be-payment-service/module/core/storage"
	"github.com/modami/be-payment-service/module/core/usecases"
	"github.com/modami/be-payment-service/module/payment_gateway/gateway_handler"
	httpRouter "github.com/modami/be-payment-service/module/payment_gateway/router"
	"github.com/modami/be-payment-service/module/payment_gateway_adapter"
	"github.com/modami/be-payment-service/module/payment_gateway_adapter/bank_transfer"
	"github.com/modami/be-payment-service/module/payment_gateway_adapter/momo"
	"github.com/modami/be-payment-service/module/payment_gateway_adapter/vnpay"
	"github.com/modami/be-payment-service/module/payment_gateway_adapter/zalopay"
	"github.com/modami/be-payment-service/pkg/cache"
	"github.com/modami/be-payment-service/pkg/logger"
)

// @title Modami Payment Service
// @version 1.0
// @description Payment service API (Gin)
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @BasePath /api/v1
func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger.Init(cfg.App.Env)
	defer logger.Sync()

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := storage.NewPostgres(cfg.Database.DSN(), cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	uow := storage.NewUnitOfWork(db)

	redisClient := cache.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err := redisClient.Ping(context.Background()); err != nil {
		// allow boot without redis (rate-limit/idempotency become best-effort)
		redisClient = nil
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	// Repositories
	outboxRepo := postgres.NewOutboxRepo(db)
	paymentRepo := postgres.NewPaymentTransactionRepo(db)
	subRepo := postgres.NewSubscriptionRepo(db)
	subEvtRepo := postgres.NewSubscriptionEventRepo(db)
	pkgRepo := postgres.NewPackageRepo(db)
	unlockRepo := postgres.NewContactUnlockRepo(db)
	invoiceRepo := postgres.NewInvoiceRepo(db)
	creditWalletRepo := postgres.NewCreditWalletRepo(db)
	creditTxRepo := postgres.NewCreditTransactionRepo(db)

	// Gateways
	selector := payment_gateway_adapter.NewGatewaySelector(
		vnpay.New(vnpay.Config{
			TMNCode:    cfg.VNPay.TMNCode,
			HashSecret: cfg.VNPay.HashSecret,
			PaymentURL: cfg.VNPay.PaymentURL,
			ReturnURL:  cfg.VNPay.ReturnURL,
			IPNURL:     cfg.VNPay.IPNURL,
			QueryURL:   cfg.VNPay.QueryURL,
		}),
		momo.New(momo.Config{
			PartnerCode: cfg.MoMo.PartnerCode,
			AccessKey:   cfg.MoMo.AccessKey,
			SecretKey:   cfg.MoMo.SecretKey,
			APIEndpoint: cfg.MoMo.APIEndpoint,
			ReturnURL:   cfg.MoMo.ReturnURL,
			NotifyURL:   cfg.MoMo.NotifyURL,
		}),
		zalopay.New(zalopay.Config{
			AppID:       cfg.ZaloPay.AppID,
			Key1:        cfg.ZaloPay.Key1,
			Key2:        cfg.ZaloPay.Key2,
			CreateURL:   cfg.ZaloPay.CreateURL,
			QueryURL:    cfg.ZaloPay.QueryURL,
			RefundURL:   cfg.ZaloPay.RefundURL,
			CallbackURL: cfg.ZaloPay.CallbackURL,
			ReturnURL:   cfg.ZaloPay.ReturnURL,
		}),
		bank_transfer.New(bank_transfer.Config{
			BankName:      cfg.BankTransfer.BankName,
			AccountNumber: cfg.BankTransfer.AccountNumber,
			AccountName:   cfg.BankTransfer.AccountName,
			Branch:        cfg.BankTransfer.Branch,
		}),
	)

	// Usecases (wire circular deps via setters)
	creditUC := usecases.NewCreditUsecase(creditWalletRepo, creditTxRepo, outboxRepo, uow)
	paymentUC := usecases.NewPaymentUsecase(paymentRepo, outboxRepo, selector, uow)
	subUC := usecases.NewSubscriptionUsecase(subRepo, subEvtRepo, pkgRepo, outboxRepo, creditUC, uow)
	unlockUC := usecases.NewUnlockUsecase(unlockRepo, creditUC, outboxRepo, uow)
	invoiceUC := usecases.NewInvoiceUsecase(invoiceRepo, uow)

	paymentUC.SetSubscriptionUsecase(subUC)
	paymentUC.SetCreditUsecase(creditUC)
	subUC.SetPaymentUsecase(paymentUC)

	// Handlers
	paymentH := gateway_handler.NewPaymentHandler(paymentUC)
	webhookH := gateway_handler.NewWebhookHandler(paymentUC)
	subscriptionH := gateway_handler.NewSubscriptionHandler(subUC)
	unlockH := gateway_handler.NewUnlockHandler(unlockUC)
	invoiceH := gateway_handler.NewInvoiceHandler(invoiceUC)
	creditH := gateway_handler.NewCreditHandler(creditUC)

	// Router
	r := httpRouter.New(httpRouter.Deps{
		JWTSecret: cfg.JWT.Secret,
		Redis:     redisClient,

		Payment:      paymentH,
		Webhook:      webhookH,
		Subscription: subscriptionH,
		Invoice:      invoiceH,
		Unlock:       unlockH,
		Credit:       creditH,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

