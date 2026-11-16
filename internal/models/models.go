package models

import "time"

const (
	RoleGuest   = "00000000-0000-0000-0000-000000000001"
	RoleClient  = "00000000-0000-0000-0000-000000000002"
	RoleManager = "00000000-0000-0000-0000-000000000003"
	RoleAdmin   = "00000000-0000-0000-0000-000000000004"
)

const (
	StatusAccepted   = "Принят"
	StatusInProgress = "В работе"
	StatusAssembled  = "Собран"
	StatusReady      = "Готов к выдаче"
	StatusIssued     = "Получен"
	StatusCancelled  = "Отменён"
)

var AllStatuses = []string{StatusAccepted, StatusInProgress, StatusAssembled, StatusReady, StatusIssued, StatusCancelled}
var AllowedToChange = []string{StatusAccepted, StatusInProgress, StatusAssembled, StatusReady, StatusIssued}
var ValidGenders = []string{"M", "W", "U"}
var ValidSeasons = []string{"Зима", "Весна", "Лето", "Осень", "Демисезон"}

type Product struct {
	ID          string  `db:"id" json:"id"`
	Name        string  `db:"name" json:"name"`
	Type        string  `db:"type" json:"type"`
	Gender      string  `db:"gender" json:"gender"`
	Size        string  `db:"size" json:"size"`
	Color       string  `db:"color" json:"color"`
	Description string  `db:"description" json:"description"`
	Brand       string  `db:"brand" json:"brand"`
	Season      string  `db:"season" json:"season"`
	Price       float64 `db:"price" json:"price"`
	Quantity    int     `db:"quantity" json:"quantity"`
	MainPhotoID *string `db:"main_photo_id" json:"main_photo_id,omitempty"`
}

type Photo struct {
	ID        string `db:"id" json:"id"`
	Filename  string `db:"filename" json:"filename"`
	ProductID string `db:"product_id" json:"product_id"`
}

type Role struct {
	ID   string `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type User struct {
	ID               string    `db:"id" json:"id"`
	Name             string    `db:"name" json:"name"`
	Email            string    `db:"email" json:"email"`
	Phone            string    `db:"phone" json:"phone"`
	PasswordHash     string    `db:"password_hash" json:"-"`
	RoleID           string    `db:"role_id" json:"role_id"`
	RegistrationDate time.Time `db:"registration_date" json:"registration_date"`
	Balance          float64   `db:"balance" json:"balance"`
}

type Cart struct {
	ID        string `db:"id" json:"id"`
	UserID    string `db:"user_id" json:"user_id"`
	ProductID string `db:"product_id" json:"product_id"`
	Count     int    `db:"count" json:"count"`
}

type Order struct {
	ID           string    `db:"id" json:"id"`
	UserID       string    `db:"user_id" json:"user_id"`
	DateTime     time.Time `db:"datetime" json:"datetime"`
	Status       string    `db:"status" json:"status"`
	Refunded     bool      `db:"refunded" json:"refunded"`
	CancelReason string    `db:"cancel_reason" json:"cancel_reason"`
}

type OrderItem struct {
	ID        string  `db:"id" json:"id"`
	OrderID   string  `db:"order_id" json:"order_id"`
	ProductID string  `db:"product_id" json:"product_id"`
	Price     float64 `db:"price" json:"price"`
	Quantity  int     `db:"quantity" json:"quantity"`
}

type OrderListItem struct {
	ID           string    `db:"id" json:"id"`
	UserID       string    `db:"user_id" json:"user_id"`
	UserName     string    `db:"user_name" json:"user_name"`
	DateTime     time.Time `db:"datetime" json:"datetime"`
	Status       string    `db:"status" json:"status"`
	ItemsCount   int       `db:"items_count" json:"items_count"`
	Total        float64   `db:"total" json:"total"`
	Refunded     bool      `db:"refunded" json:"refunded"`
	CancelReason string    `db:"cancel_reason" json:"cancel_reason"`
}

type OrderItemDetail struct {
	ID          string  `db:"id" json:"id"`
	ProductID   string  `db:"product_id" json:"product_id"`
	ProductName string  `db:"product_name" json:"product_name"`
	Price       float64 `db:"price" json:"price"`
	Quantity    int     `db:"quantity" json:"quantity"`
	PhotoID     *string `db:"photo_id" json:"photo_id,omitempty"`
}

type OrderDetail struct {
	ID           string            `db:"id" json:"id"`
	UserID       string            `db:"user_id" json:"user_id"`
	UserName     string            `db:"user_name" json:"user_name"`
	DateTime     time.Time         `db:"datetime" json:"datetime"`
	Status       string            `db:"status" json:"status"`
	Refunded     bool              `db:"refunded" json:"refunded"`
	CancelReason string            `db:"cancel_reason" json:"cancel_reason"`
	Total        float64           `json:"total"`
	Items        []OrderItemDetail `json:"items"`
}
