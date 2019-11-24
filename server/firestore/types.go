package firestore

import (
	"time"

	firebase "firebase.google.com/go"
)

// Client は firestore のクライアント構造体です。
type Client struct {
	*firebase.App
}

// Products は Message API と共有する製品情報です。
type Products map[string]*Product

// Product は ユーザのセッションで保持する製品情報をまとめた構造体です。
type Product struct {
	Name     string `firestore:"name,omitempty" json:"name"`
	Stock    int    `firestore:"stock,omitempty" json:"stock"`
	Reserved bool   `firestore:"reserved,omitempty" json:"reserved"`
}

type ProductDocument struct {
	ID          string
	ProductItem ProductItem
}

type ProductItem struct {
	Name       string `firestore:"name,omitempty"`
	Price      int    `firestore:"price,omitempty"`
	PictureURL string `firestore:"picture_url,omitempty"`
}

// User は firestore の ユーザー情報をまとめた構造体です。firestore の users document の属性値と一致します。
type User struct {
	UserID      string    `firestore:"user_id,omitempty" json:"user_id"`
	DisplayName string    `firestore:"display_name,omitempty" json:"display_name"`
	CreatedAt   time.Time `firestore:"created_at,omitempty" json:"-"`
}

// StockDocuments は firestore の stock document を指す構造体です。
type StockDocuments map[string]Stock

type Order struct {
	User       User      `firestore:"user,omitempty" json:"user"`
	Date       string    `firestore:"date,omitempty" json:"date"`
	Time       string    `firestore:"time,omitempty" json:"time"`
	Location   string    `firestore:"location,omitempty" json:"location"`
	Products   Products  `firestore:"products,omitempty" json:"products"`
	TotalPrice int       `firestore:"total_price,omitempty" json:"total_price"`
	InTrade    bool      `firestore:"in_trade,omitempty" json:"in_trade"`
	InProgress bool      `firestore:"in_progress,omitempty" json:"in_progress"`
	CreatedAt  time.Time `firestore:"created_at,omitempty" json:"created_at"`
	UpdatedAt  time.Time `firestore:"updated_at,omitempty" json:"updated_at"`
	ChatID     string    `firestore:"chat_id,omitempty" json:"chat_id"`
}

type OrderDocument struct {
	ID    string `firestore:"document_id,omitempty" json:"document_id"`
	Order Order  `firestore:"order,omitempty" json:"order"`
}

type Stock struct {
	ProductID string `firestore:"product_id,omitempty"`
	Stock     int    `firestore:"stock,omitempty"`
}

type Location struct {
	Name string `firestore:"name,omitempty"`
}

type ChatDocument struct {
	Messages []Message `firestore:"messages,omitempty" json:"messages"`
	OrderID  string    `firestore:"order_id,omitempty" json:"order_id"`
}

type Message struct {
	Content   string    `firestore:"content,omitempty" json:"content"`
	CreatedAt time.Time `firestore:"created_at,omitempty" json:"created_at"`
	UserID    string    `firestore:"user_id,omitempty" json:"user_id"`
}

type Menu []ProductDocument
