package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
	"golang.org/x/sync/errgroup"
)

type Requisicao struct {
	ID      string  `json:"id"`
	Preco   float64 `json:"preco"`
	Imposto float64 `json:"imposto"`
}

type Resposta struct {
	UUID         string  `json:"uuid"`
	PrecoTotal   float64 `json:"preco_total"`
	ValorImposto float64 `json:"valor_imposto"`
	RecebidoEm   string  `json:"recebido_em"`
	ProcessadoEm string  `json:"processado_em"`
}

var (
	idCounter uint64
)

func generateID() string {
	newID := atomic.AddUint64(&idCounter, 1)
	return fmt.Sprintf("%d", newID)
}

func main() {
	http.HandleFunc("/process", RecebeRequisicao)

	// Iniciar servidor HTTP
	porta := 8080
	fmt.Printf("Servidor ouvindo na porta :%d...\n", porta)
	err := http.ListenAndServe(fmt.Sprintf(":%d", porta), nil)
	if err != nil {
		fmt.Println("Erro ao iniciar o servidor HTTP:", err)
	}
}
func RecebeRequisicao(w http.ResponseWriter, r *http.Request) {
	var req []Requisicao
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Erro ao decodificar o corpo da requisição", http.StatusBadRequest)
		return
	}

	// Criar fila para armazenar as requisições
	filaRequisicoes := make(chan Requisicao, len(req))
	for _, r := range req {
		filaRequisicoes <- r
	}
	close(filaRequisicoes)

	// Processar requisições
	respostas := processaRequisicoes(filaRequisicoes)

	// Separar as respostas em batches
	tamanhoBatch := 100
	for i := 0; i < len(respostas); i += tamanhoBatch {
		fim := i + tamanhoBatch
		if fim > len(respostas) {
			fim = len(respostas)
		}
		batch := respostas[i:fim]

		// Envia batch como resposta
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(batch)
		if err != nil {
			http.Error(w, "Erro ao codificar a resposta", http.StatusInternalServerError)
			return
		}
	}
}

func processaRequisicoes(filaRequisicoes <-chan Requisicao) []Resposta {
	var respostas []Resposta

	//Aguardar todas as goroutines
	var eg errgroup.Group

	// Número de goroutines para processar as requisições em paralelo
	numGoroutines := 5

	//Processa cada requisição
	processaRequisicao := func(req Requisicao) Resposta {
		reqCopy := req
		reqCopy.ID = generateID()

		recebidoEm := time.Now()
		inicio := time.Now()
		processaRequisicao(&reqCopy)
		fim := time.Now()
		tempoProcessamento := fim.Sub(inicio)

		resp := Resposta{
			UUID:          reqCopy.ID,
			PrecoTotal:    reqCopy.Preco * (1 + reqCopy.Imposto),
			ValorImposto:  reqCopy.Preco * reqCopy.Imposto,
			RecebidoEm:    recebidoEm.Format(time.RFC3339Nano),
			ProcessadoEm:  fim.Format(time.RFC3339Nano),
		}

		fmt.Printf("Requisição ID %s processada em %.2f ms\n", reqCopy.ID, tempoProcessamento.Seconds()*1000)
		return resp
	}

	// Inicie as goroutines para processar as requisições
	for i := 0; i < numGoroutines; i++ {
		eg.Go(func() error {
			for req := range filaRequisicoes {
				resposta := processaRequisicao(req)
				respostas = append(respostas, resposta)
			}
			return nil
		})
	}

	//Aguardar todas as goroutines terminarem
	if err := eg.Wait(); err != nil {
		fmt.Println("Erro ao aguardar a conclusão das goroutines:", err)
	}
	return respostas
}

func processaRequisicao(req *Requisicao) {
	fmt.Printf("Processando requisição ID: %s\n", req.ID)
}
