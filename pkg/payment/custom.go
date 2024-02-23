package payment

import (
	"encoding/json"
	"errors"
	"fmt"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gofrs/uuid"
	"github.com/qiniu/go-sdk/v7/sms/bytes"
	"net/http"
	"net/url"
)

// Custom payment client
type Custom struct {
	client     request.Client
	endpoint   string
	authClient auth.Auth
}

const (
	paymentTTL            = 3600 * 24 // 24h
	CallbackSessionPrefix = "custom_payment_callback_"
)

func newCustomClient(endpoint, secret string) *Custom {
	authClient := auth.HMACAuth{
		SecretKey: []byte(secret),
	}
	return &Custom{
		endpoint:   endpoint,
		authClient: auth.General,
		client: request.NewClient(
			request.WithCredential(authClient, paymentTTL),
			request.WithMasterMeta(),
		),
	}
}

// Request body from Cloudreve to create a new payment
type NewCustomOrderRequest struct {
	Name      string `json:"name"`       // Order name
	OrderNo   string `json:"order_no"`   // Order number
	NotifyURL string `json:"notify_url"` // Payment callback url
	Amount    int64  `json:"amount"`     // Order total amount
}

// Create a new payment
func (pay *Custom) Create(order *model.Order, pack *serializer.PackProduct, group *serializer.GroupProducts, user *model.User) (*OrderCreateRes, error) {
	callbackID := uuid.Must(uuid.NewV4())
	gateway, _ := url.Parse(fmt.Sprintf("/api/v3/callback/custom/%s/%s", order.OrderNo, callbackID))
	callback, err := auth.SignURI(pay.authClient, model.GetSiteURL().ResolveReference(gateway).String(), paymentTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to sign callback url: %w", err)
	}

	cache.Set(CallbackSessionPrefix+callbackID.String(), order.OrderNo, paymentTTL)

	body := &NewCustomOrderRequest{
		Name:      order.Name,
		OrderNo:   order.OrderNo,
		NotifyURL: callback.String(),
		Amount:    int64(order.Price * order.Num),
	}
	bodyJson, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to encode body: %w", err)
	}

	res, err := pay.client.Request("POST", pay.endpoint, bytes.NewReader(bodyJson)).
		CheckHTTPResponse(http.StatusOK).DecodeResponse()
	if err != nil {
		return nil, fmt.Errorf("failed to request payment gateway: %w", err)
	}

	if res.Code != 0 {
		return nil, errors.New(res.Error)
	}

	if _, err := order.Create(); err != nil {
		return nil, ErrInsertOrder.WithError(err)
	}

	return &OrderCreateRes{
		Payment: true,
		QRCode:  res.Data.(string),
		ID:      order.OrderNo,
	}, nil
}
