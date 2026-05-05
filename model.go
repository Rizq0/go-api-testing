package main

import "database/sql"

type Product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

func getProductsFromDB(db *sql.DB) ([]Product, error) {
	rows, err := db.Query("SELECT id, name, quantity, price FROM products")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Quantity, &p.Price); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (p *Product) getProductByIDFromDB(db *sql.DB) error {
	row := db.QueryRow("SELECT id, name, quantity, price FROM products WHERE id = $1", p.ID)
	if err := row.Scan(&p.ID, &p.Name, &p.Quantity, &p.Price); err != nil {
		return err
	}
	return nil
}

func (p *Product) createProductInDB(db *sql.DB) error {
	err := db.QueryRow("INSERT INTO products (name, price, quantity) VALUES ($1, $2, $3) RETURNING id", p.Name, p.Price, p.Quantity).Scan(&p.ID)
	if err != nil {
		return err
	}
	return nil
}

func (p *Product) updateProductInDB(db *sql.DB) error {
	result, err := db.Exec("UPDATE products SET name = $1, price = $2, quantity = $3 WHERE id = $4", p.Name, p.Price, p.Quantity, p.ID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (p *Product) deleteProductFromDB(db *sql.DB) error {
	result, err := db.Exec("DELETE FROM products WHERE id = $1", p.ID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}