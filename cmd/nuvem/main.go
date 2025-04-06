package main

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/LuisMarioRC/PBL-ELECTRIC-CARS/cmd/models"
)

var mu sync.Mutex

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)

	n, err := conn.Read(buffer)
	if err != nil {
		return
	}

	mensagem := string(buffer[:n])
	parts := strings.Split(mensagem, "|")
	if len(parts) < 2 {
		conn.Write([]byte("Formato inválido da mensagem"))
		return
	}

	comando := parts[0]
	id := parts[1]

	mu.Lock()
	defer mu.Unlock()

	fmt.Printf("🤖 Recebido comando: %s, ID: %s\n", comando, id)

	switch comando {
	case "Registrar ponto":
		// Verifica se o ponto existe no grafo
		if _, exists := models.GraphInstance.Nodes[id]; !exists {
			conn.Write([]byte(fmt.Sprintf("Erro: Ponto %s não está registrado no grafo da nuvem.", id)))
			return
		}

		// Verifica se o ponto já existe no sistema
		if _, exists := models.PontosDisponiveis[id]; !exists {
			models.PontosDisponiveis[id] = true
			fmt.Printf("Ponto %s registrado com sucesso.\n", id)
			conn.Write([]byte(fmt.Sprintf("Ponto %s registrado com sucesso e conectado ao sistema.", id)))
		} else {
			// Atualiza o status do ponto para disponível
			models.PontosDisponiveis[id] = true
			conn.Write([]byte(fmt.Sprintf("Ponto %s já estava registrado. Atualizado como disponível e conectado.", id)))
		}

	case "Carro recarregado":
		if _, exists := models.GraphInstance.Nodes[id]; !exists {
			conn.Write([]byte(fmt.Sprintf("Erro: Ponto %s não está registrado no grafo da nuvem.", id)))
			return
		}
		if _, exists := models.PontosDisponiveis[id]; exists {
			models.PontosDisponiveis[id] = true
			fmt.Printf("Carro terminou de recarregar no ponto %s\n", id)

			// Verifica se há carros na fila para este ponto
			if models.FilaEspera[id] > 0 {
				models.FilaEspera[id]--
				models.PontosDisponiveis[id] = false
				fmt.Printf("Ponto %s: Próximo carro na fila será atendido.\n", id)
				conn.Write([]byte(fmt.Sprintf("Ponto %s: Próximo carro pode ser atendido", id)))
			} else {
				fmt.Printf("Ponto %s liberado. Nenhum carro na fila.\n", id)
				conn.Write([]byte(fmt.Sprintf("Ponto %s liberado.", id)))
			}
		} else {
			conn.Write([]byte(fmt.Sprintf("Erro: Ponto %s não encontrado.", id)))
		}

	case "Carro precisa recarga":
		pontoMaisProximo, _, distancia := models.Dijkstra("Ponto-1") // Localização atual poderia ser dinâmica

		if pontoMaisProximo != "" {
			if models.PontosDisponiveis[pontoMaisProximo] {
				models.PontosDisponiveis[pontoMaisProximo] = false
				fmt.Printf("Carro %s iniciou recarga no ponto %s (distância: %.2f)\n", id, pontoMaisProximo, distancia)
				conn.Write([]byte(fmt.Sprintf("Carro %s: Dirija-se ao ponto: %s (distância: %.2f)", id, pontoMaisProximo, distancia)))
			} else {
				// Adiciona o carro à fila do ponto mais próximo
				models.FilaEspera[pontoMaisProximo]++
				posicaoFila := models.FilaEspera[pontoMaisProximo]
				fmt.Printf("Carro %s adicionado à fila do ponto %s. Posição na fila: %d\n", id, pontoMaisProximo, posicaoFila)
				conn.Write([]byte(fmt.Sprintf("Carro %s: Ponto ocupado. Você está na fila do %s: posição %d", id, pontoMaisProximo, posicaoFila)))
			}
		} else {
			// Caso nenhum ponto esteja disponível ou no grafo
			fmt.Printf("Carro %s: Nenhum ponto disponível no momento.\n", id)
			conn.Write([]byte(fmt.Sprintf("Carro %s: Nenhum ponto disponível no momento.", id)))
		}

	default:
		conn.Write([]byte("Comando desconhecido"))
	}
}

func main() {
	// Inicializa o grafo com os pontos e suas conexões
	models.InicializarGrafo()

	// Inicia o servidor na porta 8080
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Servidor Nuvem iniciado na porta 8080")
	fmt.Println("Aguardando conexões dos pontos de recarga...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}
