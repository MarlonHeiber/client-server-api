package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type CotacaoResponse struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatalf("Erro ao criar requisição: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Erro ao fazer requisição: %v", err)
	}
	defer res.Body.Close()

	var cotacao CotacaoResponse

	err = json.NewDecoder(res.Body).Decode(&cotacao)
	if err != nil {
		log.Fatalf("Erro ao decodificar resposta: %v", err)
	}

	err = criaArquivoESalva(fmt.Sprintf("Dólar: %s", cotacao.Bid))
	if err != nil {
		log.Fatalf("Erro ao decodificar resposta: %v", err)
	}

}

func criaArquivoESalva(content string) error {
	f, err := os.Create("cotacao.txt")
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(content))
	if err != nil {
		return err
	}

	return nil
}
