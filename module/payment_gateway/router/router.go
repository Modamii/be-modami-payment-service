package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/modami/be-payment-service/module/payment_gateway/gateway_handler"
	"github.com/modami/be-payment-service/module/payment_gateway/middleware"
	"github.com/modami/be-payment-service/pkg/cache"
)

type Deps struct {
	JWTSecret string
	Redis     *cache.Client

	Payment      *gateway_handler.PaymentHandler
	Webhook      *gateway_handler.WebhookHandler
	Subscription *gateway_handler.SubscriptionHandler
	Invoice      *gateway_handler.InvoiceHandler
	Unlock       *gateway_handler.UnlockHandler
	Credit       *gateway_handler.CreditHandler
}

func New(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestIDMiddleware())

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")

	// Public endpoints
	api.POST("/webhooks/vnpay", d.Webhook.VNPayIPN)
	api.POST("/webhooks/momo", d.Webhook.MoMoIPN)
	api.POST("/webhooks/zalopay", d.Webhook.ZaloPayCallback)

	api.GET("/payments/return/vnpay", d.Webhook.VNPayReturn)
	api.GET("/payments/return/momo", d.Webhook.MoMoReturn)
	api.GET("/payments/return/zalopay", d.Webhook.ZaloPayReturn)

	api.GET("/packages", d.Subscription.ListPackages)
	api.GET("/packages/:code", d.Subscription.GetPackage)

	// Authed endpoints
	authed := api.Group("")
	authed.Use(middleware.AuthMiddleware(d.JWTSecret))
	if d.Redis != nil {
		authed.Use(middleware.IdempotencyMiddleware(d.Redis))
		authed.Use(middleware.RateLimitMiddleware(d.Redis, "api", 200, time.Minute))
	}

	authed.POST("/payments", d.Payment.CreatePayment)
	authed.GET("/payments/:id", d.Payment.GetPayment)
	authed.GET("/payments/:id/status", d.Payment.GetPaymentStatus)
	authed.GET("/payments/history", d.Payment.GetHistory)

	authed.POST("/subscriptions", d.Subscription.Subscribe)
	authed.GET("/subscriptions/current", d.Subscription.GetCurrent)
	authed.PUT("/subscriptions/current/cancel", d.Subscription.Cancel)
	authed.PUT("/subscriptions/current/renew", d.Subscription.Renew)
	authed.POST("/subscriptions/upgrade", d.Subscription.Upgrade)
	authed.GET("/subscriptions/history", d.Subscription.GetHistory)

	authed.GET("/invoices", d.Invoice.ListInvoices)
	authed.GET("/invoices/:id", d.Invoice.GetInvoice)
	authed.GET("/invoices/:id/download", d.Invoice.DownloadInvoice)

	authed.POST("/unlocks", d.Unlock.Unlock)
	authed.GET("/unlocks", d.Unlock.ListUnlocks)
	authed.GET("/unlocks/check/:product_id", d.Unlock.CheckUnlock)

	authed.GET("/credits/balance", d.Credit.GetBalance)
	authed.GET("/credits/transactions", d.Credit.GetTransactions)
	authed.POST("/credits/purchase", d.Credit.Purchase)

	admin := authed.Group("/admin")
	admin.Use(middleware.AdminMiddleware())
	admin.POST("/credits/adjust", d.Credit.AdminAdjust)

	return r
}

