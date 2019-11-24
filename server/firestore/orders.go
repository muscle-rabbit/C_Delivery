package firestore

import (
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

// FetchOrderDocs は 全ての order document を取得するメソッドです。
func (client *Client) FetchOrders() ([]OrderDocument, error) {
	var orderDocumens []OrderDocument

	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't create oreder in fetchUserOrderID: %v", err)
	}

	iter := c.Collection("orders").Documents(ctx)
	for {
		var order Order
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		err = doc.DataTo(&order)
		if err != nil {
			return nil, err
		}
		orderDocumens = append(orderDocumens, OrderDocument{ID: doc.Ref.ID, Order: order})
	}

	return orderDocumens, err
}

// CreateOrder は order Document を作成し、その document の ID を返すメソッドです。
func (client *Client) CreateOrder(userID string) (string, error) {
	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return "", fmt.Errorf("couldn't create client in createOrder: %v", err)
	}

	user, err := client.FetchUserByDocID(userID)
	if err != nil {
		return "", err
	}

	ref, _, err := c.Collection("orders").Add(ctx, Order{User: user, CreatedAt: time.Now(), InTrade: true, InProgress: true})
	if err != nil {
		return "", fmt.Errorf("couldn't create document in createOrder: %v", err)
	}

	return ref.ID, nil
}

// UpdateOrderInChat は Message API から order document を更新するためのメソッドです。
func (client *Client) UpdateOrderInChat(userID string, orderDoc OrderDocument, willUpdated string) error {

	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("couldn't create client in updateOrderInChat: %v", err)
	}

	// var willUpdated string

	// switch prevStep {
	// case reservateDate:
	// 	willUpdated = "date"
	// case reservateTime:
	// 	if isTimeMessage(order.Time) {
	// 		willUpdated = "time"
	// 	} else {
	// 		willUpdated = "products"
	// 	}
	// case setLocation:
	// 	willUpdated = "location"
	// case setMenu:
	// 	willUpdated = "total_price"
	// }

	_, err = c.Collection("orders").Doc(orderDoc.ID).Set(ctx, orderDoc.Order, firestore.Merge([]string{willUpdated}))
	if err != nil {
		return fmt.Errorf("couldn't update order in updateOrderInChat: %v", err)
	}

	return nil
}

// CompleteOrderInChat は Message API から order document を完了するためのメソッドです。
func (client *Client) CompleteOrderInChat(orderID string, userID string, totalPrice int) error {
	// userSession := app.sessionStore.sessions[userID]

	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("couldn't create client in completeOrderInChat: %v", err)
	}

	chatID, err := client.createChatRoom(orderID, userID)
	if err != nil {
		return err
	}

	_, err = c.Collection("orders").Doc(orderID).Set(ctx, map[string]interface{}{
		"chat_id": chatID, "in_progress": false, "total_price": totalPrice}, firestore.MergeAll)
	return err
}

// DeleteOrder は order document を削除するためのメソッドです。
func (client *Client) DeleteOrder(orderID string) error {
	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("couldn't create client in addUser: %v", err)
	}
	fmt.Println("indelete")

	ref := c.Collection("orders").Doc(orderID)
	_, err = ref.Delete(ctx)
	fmt.Println(ref.ID + " is beign deleted...")
	return err
}

// FinishTrade は order document の 「取引中」ステータスを false にし取引を完了させるメソッドです。
func (client *Client) FinishTrade(orderID string) error {
	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("couldn't create client in finshTrade: %v", err)
	}

	_, err = c.Collection("orders").Doc(orderID).Set(ctx, Order{InTrade: false}, firestore.Merge([]string{"in_trade"}))

	return err
}

// UnfinishTrade は order document の 「取引中」ステータスを true にし取引を完了させるメソッドです。
func (client *Client) UnfinishTrade(orderID string) error {
	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("couldn't create client in finshTrade: %v", err)
	}

	_, err = c.Collection("orders").Doc(orderID).Set(ctx, Order{InTrade: true}, firestore.Merge([]string{"in_trade"}))

	return err
}

// FetchUserOrder は 単一の order document を取得するメソッドです。
func (client *Client) FetchUserOrder(orderID string) (OrderDocument, error) {

	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return OrderDocument{}, fmt.Errorf("couldn't create oreder in fetchUserOrderID: %v", err)
	}

	var order Order
	doc, err := c.Collection("orders").Doc(orderID).Get(ctx)
	doc.DataTo(&order)
	return OrderDocument{ID: doc.Ref.ID, Order: order}, nil
}

// ReserveProducts は 商品を予約するためのメソッドです。
func (client *Client) ReserveProducts(userID string, products Products) (bool, error) {
	// productM, err := parseMessageToProductText(message, app.service.menu)
	// if err != nil {
	// 	return false, err
	// }
	// userSession := app.sessionStore.sessions[userID]

	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return false, fmt.Errorf("couldn't create client in reserveProduct: %v", err)
	}

	stockDocs, err := client.FetchStocks()
	if err != nil {
		return false, err
	}

	for id, product := range products {
		if product.Reserved {
			continue
		}
		stock, err := stockDocs.SearchDocByProductID(id)
		if err != nil {
			return false, err
		}
		if stock.Stock >= product.Stock {
			stock.Stock -= product.Stock
			docID, err := stockDocs.SearchDocIDByProductID(id)
			if err != nil {
				return false, err
			}
			if _, err := c.Collection("stocks").Doc(docID).Set(ctx, stock); err != nil {
				return false, err
			}
			// この関数の外で行う。
			// if err := userSession.products.setProduct(productM); err != nil {
			// 	return false, err
			// }
			// userSession.products[id].Reserved = true
			return false, nil
		}
		return true, nil
	}
	return false, err
}

func (client *Client) CancelOrder(userID string, orderID string, products Products) error {
	// userSession := app.sessionStore.sessions[userID]
	if err := client.DeleteOrder(orderID); err != nil {
		return err
	}
	if err := client.RestoreStocks(userID, products); err != nil {
		return err
	}
	// if err := app.sessionStore.deleteUserSession(userID); err != nil {
	// 	return err
	// }
	return nil
}

func (menu Menu) makeMesssageText(p Products) (string, error) {
	var menuText string
	for id, product := range p {
		doc, err := menu.searcProductByID(id)
		if err != nil {
			return "", err
		}
		menuText += "・" + doc.ProductItem.Name + " x " + strconv.Itoa(product.Stock) + "\n"
	}
	return menuText, nil
}
