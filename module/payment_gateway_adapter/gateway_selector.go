package payment_gateway_adapter

import apperrors "github.com/modami/be-payment-service/pkg/errors"

// GatewaySelector holds all registered gateways and selects one by method name.
type GatewaySelector struct {
	gateways map[string]PaymentGateway
}

// NewGatewaySelector creates a GatewaySelector from the given gateways.
func NewGatewaySelector(gateways ...PaymentGateway) *GatewaySelector {
	m := make(map[string]PaymentGateway, len(gateways))
	for _, g := range gateways {
		m[g.Name()] = g
	}
	return &GatewaySelector{gateways: m}
}

// Select returns the gateway for the given payment method, or ErrUnsupportedGateway.
func (s *GatewaySelector) Select(method string) (PaymentGateway, error) {
	gw, ok := s.gateways[method]
	if !ok {
		return nil, apperrors.ErrUnsupportedGateway
	}
	return gw, nil
}
