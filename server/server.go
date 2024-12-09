package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type CotacaoMoeda struct {
	Usdbrl struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

type Cotacao struct {
	ID    string
	Moeda string
	Valor string
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", handler)
	log.Println("Server listening on port 8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Conexão com SQLite
	db, err := sql.Open("sqlite3", "./goexpert.db")
	if err != nil {
		http.Error(w, "Erro no banco de dados", http.StatusInternalServerError)
		log.Printf("Erro ao conectar ao banco: %v", err)
		return
	}
	defer db.Close()

	createTable(db)

	cotacaoDolar, err := BuscaCotacaoDolar()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Erro ao buscar cotação: %v", err)
	}

	response := map[string]string{"bid": cotacaoDolar.Usdbrl.Bid}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Erro ao codificar resposta: %v", err)
		return
	}

	cotacao := NovaCotacao("USDBRL", cotacaoDolar.Usdbrl.Bid)
	err = insertCotacao(db, cotacao)
	if err != nil {
		log.Printf("Erro ao codificar resposta: %v", err)
		return
	}

}

func BuscaCotacaoDolar() (CotacaoMoeda, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return CotacaoMoeda{}, fmt.Errorf("erro ao criar requisição: %w", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return CotacaoMoeda{}, fmt.Errorf("erro ao realizar requisição: %w", err)
	}
	defer res.Body.Close()

	response, err := io.ReadAll(res.Body)
	if err != nil {
		return CotacaoMoeda{}, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	var data CotacaoMoeda
	err = json.Unmarshal(response, &data)
	if err != nil {
		return CotacaoMoeda{}, err
	}
	return data, nil

}

func createTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS cotacao (
		id TEXT PRIMARY KEY,
		moeda TEXT NOT NULL,
		valor REAL NOT NULL
	);`
	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}
}

func NovaCotacao(moeda string, valor string) *Cotacao {
	return &Cotacao{
		ID:    uuid.New().String(),
		Moeda: moeda,
		Valor: valor,
	}
}

func insertCotacao(db *sql.DB, cotacao *Cotacao) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	stmt, err := db.PrepareContext(ctx, "INSERT INTO cotacao(id, moeda, valor) VALUES(?, ?, ?)")
	if err != nil {
		return fmt.Errorf("erro ao preparar statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, cotacao.ID, cotacao.Moeda, cotacao.Valor)
	return err
}
