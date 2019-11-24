package firestore

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

func (client *Client) FetchProduct(productID string) (ProductDocument, error) {
	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return ProductDocument{}, fmt.Errorf("couldn't create client in fetchProduct: %v", err)
	}

	doc, err := c.Collection("products").Doc(productID).Get(ctx)
	if err != nil {
		return ProductDocument{}, fmt.Errorf("couldn't get document in fetchProduct: %v", err)
	}

	var prdItem ProductItem
	if err := doc.DataTo(&prdItem); err != nil {
		return ProductDocument{}, err
	}

	return ProductDocument{ID: doc.Ref.ID, ProductItem: prdItem}, err
}

func (client *Client) FetchProducts() ([]ProductDocument, error) {
	ctx := context.Background()
	c, err := client.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't create client in getMenu: %v", err)
	}

	var docs []ProductDocument

	iter := c.Collection("products").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var prdItem ProductItem
		if err := doc.DataTo(&prdItem); err != nil {
			return nil, err
		}

		docs = append(docs, ProductDocument{ID: doc.Ref.ID, ProductItem: prdItem})

	}

	return docs, nil
}

func (products Products) setProduct(p Products) error {
	for id, product := range p {
		if products[id] != nil {
			products[id].Stock += product.Stock
			return nil
		}
		products[id] = product
		return nil
	}
	return fmt.Errorf("couldn't set prodct in session")
}
