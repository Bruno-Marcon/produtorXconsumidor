package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

type Requisicao struct {
	ID      string  `json:"id"`
	Preco   float64 `json:"preco"`
	Imposto float64 `json:"imposto"`
}

type Resposta struct {
	UUID        string  `json:"uuid"`
	PrecoTotal  float64 `json:"preco_total"`
	ValorImposto float64 `json:"valor_imposto"`
}

type RespostaComTempo struct {
	Resposta
	TempoTotalExecucao float64 `json:"tempo_total_execucao"`
}

func main() {
	registros := gerarExemplosDeRegistros(1000)
	enderecoServidor := "http://localhost:8080/process"
	respostasComTempo, err := enviarRegistrosParaServidor(enderecoServidor, registros)
	if err != nil {
		fmt.Println("Erro ao enviar registros para o servidor:", err)
		return
	}

	err = exportarRespostasParaJSON("respostas.json", respostasComTempo)
	if err != nil {
		fmt.Println("Erro ao exportar respostas para JSON:", err)
		return
	}

	fmt.Println("Respostas exportadas com sucesso para respostas.json.")
}

func gerarExemplosDeRegistros(quantidade int) []Requisicao {
	registros := make([]Requisicao, quantidade)
	for i := 0; i < quantidade; i++ {
		registro := Requisicao{
			ID:     fmt.Sprintf("ID-%d", i+1),
			Preco:  gerarValorAleatorio(10.0, 1000.0),
			Imposto: gerarValorAleatorio(0.05, 2),
		}
		registros[i] = registro
	}
	return registros
}

func gerarValorAleatorio(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func enviarRegistrosParaServidor(enderecoServidor string, registros []Requisicao) ([]RespostaComTempo, error) {
	var respostasComTempo []RespostaComTempo

	// Defina o tamanho do batch
	tamanhoBatch := 100

	// Itere sobre os registros em batches
	for i := 0; i < len(registros); i += tamanhoBatch {
		// Calcule os índices para o batch atual
		fim := i + tamanhoBatch
		if fim > len(registros) {
			fim = len(registros)
		}

		// Extraia o batch atual
		batch := registros[i:fim]

		// Envie o batch para o servidor
		for _, req := range batch {
			recebidoEm := time.Now()
			resposta, err := enviarRequisicaoParaServidor(enderecoServidor, req)
			if err != nil {
				fmt.Printf("Erro ao enviar requisição para o servidor (ID: %s): %v\n", req.ID, err)
				continue
			}
			fim := time.Now()
			tempoTotalExecucao := fim.Sub(recebidoEm).Seconds() * 1000 // Convertendo para milissegundos

			respostaComTempo := RespostaComTempo{
				Resposta:           resposta,
				TempoTotalExecucao: tempoTotalExecucao,
			}

			respostasComTempo = append(respostasComTempo, respostaComTempo)
		}
	}

	return respostasComTempo, nil
}

func enviarRequisicaoParaServidor(enderecoServidor string, req Requisicao) (Resposta, error) {
	dadosJson, err := json.Marshal([]Requisicao{req})
	if err != nil {
		return Resposta{}, fmt.Errorf("Erro ao codificar JSON: %v", err)
	}

	resp, err := http.Post(enderecoServidor, "application/json", bytes.NewBuffer(dadosJson))
	if err != nil {
		return Resposta{}, fmt.Errorf("Erro ao enviar requisição POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Resposta{}, fmt.Errorf("O servidor retornou um status não-OK: %s", resp.Status)
	}

	var respostas []Resposta
	if err := json.NewDecoder(resp.Body).Decode(&respostas); err != nil {
		return Resposta{}, fmt.Errorf("Erro ao decodificar respostas JSON: %v", err)
	}

	if len(respostas) != 1 {
		return Resposta{}, fmt.Errorf("Resposta do servidor com tamanho inesperado")
	}

	return respostas[0], nil
}

func exportarRespostasParaJSON(nomeArquivo string, respostas []RespostaComTempo) error {
	// Criar um buffer de bytes para armazenar os dados JSON formatados
	var buffer bytes.Buffer

	// Iterar sobre cada resposta e escrever no buffer
	for _, resposta := range respostas {
		linha, err := json.Marshal(resposta)
		if err != nil {
			return fmt.Errorf("Erro ao codificar resposta JSON: %v", err)
		}
		// Adicionar uma nova linha após cada resposta
		buffer.Write(append(linha, '\n'))
	}

	// Escrever o conteúdo do buffer no arquivo
	err := ioutil.WriteFile(nomeArquivo, buffer.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("Erro ao escrever arquivo JSON: %v", err)
	}

	return nil
}