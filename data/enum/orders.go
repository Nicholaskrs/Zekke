package enum

type OrderStatus string

const (
	// TODO NK: Add docs
	OrderStatusCreated   OrderStatus = "Created"
	OrderStatusProcessed OrderStatus = "Processed"
	OrderStatusRejected  OrderStatus = "Rejected"
	OrderStatusDelivered OrderStatus = "Delivered"
)

func (s OrderStatus) String() string {
	return string(s)
}
