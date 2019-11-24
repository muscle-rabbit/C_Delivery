package firestore

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

func (client *Client) RestoreStocks(userID string, products Products) error {
	// userSession := app.sessionStore.sessions[userID]

	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("couldn't create client in restoreStock: %v", err)
	}

	stockDocs, err := client.FetchStocks()
	if err != nil {
		return err
	}

	for id, product := range products {
		stock, err := stockDocs.SearchDocByProductID(id)
		if err != nil {
			return err
		}
		stock.Stock += product.Stock
		docID, err := stockDocs.SearchDocIDByProductID(id)
		if err != nil {
			return err
		}
		if _, err := c.Collection("stocks").Doc(docID).Set(ctx, stock); err != nil {
			return err
		}
	}

	return nil
}

func (client *Client) FetchStocks() (StockDocuments, error) {
	stockDocs := make(StockDocuments)
	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't create client in fetchStocks: %v", err)
	}

	iter := c.Collection("stocks").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var stock Stock
		if err := doc.DataTo(&stock); err != nil {
			return nil, err
		}
		stockDocs[doc.Ref.ID] = stock
	}
	return stockDocs, nil
}
