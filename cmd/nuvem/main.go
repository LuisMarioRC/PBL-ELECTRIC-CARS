package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/LuisMarioRC/PBL-ELECTRIC-CARS/cmd/models"
)

var mu sync.Mutex
var conexoesCarros = make(map[string]net.Conn)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		mensagem, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		mensagem = strings.TrimSpace(mensagem)
		parts := strings.Split(mensagem, "|")
		if len(parts) < 2 {
			conn.Write([]byte("Formato inválido da mensagem\n"))
			return
		}

		comando := parts[0]
		id := parts[1]

		mu.Lock()
		fmt.Printf("☁️ Recebido comando: %s, ID: %s\n", comando, id)

		switch comando {
		case "Registrar ponto":
			models.Mutex.Lock()
			defer models.Mutex.Unlock()

			if _, exists := models.GraphInstance.Nodes[id]; !exists {
				conn.Write([]byte(fmt.Sprintf("Erro: Ponto %s não está registrado\n", id)))
				return
			}

			models.PontoManagerInstance.SetDisponivel(id, true)
			fmt.Printf("Ponto %s registrado\n", id)
			conn.Write([]byte(fmt.Sprintf("Ponto %s registrado e conectado\n", id)))

		case "Carro recarregado":
			models.Mutex.Lock()
			defer models.Mutex.Unlock()

			if _, exists := models.GraphInstance.Nodes[id]; !exists {
				conn.Write([]byte(fmt.Sprintf("Erro: Ponto %s não registrado\n", id)))
				return
			}

			// Libera o ponto
			models.PontoManagerInstance.SetDisponivel(id, true)
			fmt.Printf("Ponto %s liberado\n", id)

			// Verifica se há carros na fila
			go func(pontoID string) {
				models.Mutex.Lock()
				defer models.Mutex.Unlock()
			
				if fila, ok := models.PontoManagerInstance.Filas[pontoID]; ok {
					if carroID, existe := fila.Remover(); existe {
						// Marca o ponto como ocupado
						models.PontoManagerInstance.SetDisponivel(pontoID, false)
			
						// Atualiza o estado do carro
						if carro, ok := models.CarrosEstado[carroID]; ok {
							carro.NaFila = false
							carro.PontoNaFila = ""
							carro.Localizacao = pontoID
							fmt.Printf("Ponto %s: Carro %s iniciando recarga\n", pontoID, carroID)
			
							// Envia o comando diretamente para o carro
							if connCarro, exists := conexoesCarros[carroID]; exists {
								connCarro.Write([]byte(fmt.Sprintf("Inicie recarga|%s\n", pontoID)))
							} else {
								fmt.Printf("Erro: Conexão com o carro %s não encontrada\n", carroID)
							}
						}
					}
				}
			}(id)

		case "Verificar fila":
			if len(parts) < 3 {
				conn.Write([]byte("Formato inválido\n"))
				return
			}

			pontoID := id
			carroID := parts[2]

			models.Mutex.Lock()
			defer models.Mutex.Unlock()

			fila, ok := models.PontoManagerInstance.Filas[pontoID]  // Alterado de filas para Filas
			if !ok {
				conn.Write([]byte("Ponto inválido\n"))
				return
			}

			// Verifica se o carro é o próximo da fila
			if fila.Tamanho() > 0 && fila.Carros[0] == carroID {
				disponivel, _ := models.PontoManagerInstance.GetDisponivel(pontoID)
				if disponivel {
					models.PontoManagerInstance.SetDisponivel(pontoID, false)
					fila.Remover()
					
					if carro, existe := models.CarrosEstado[carroID]; existe {
						carro.NaFila = false
						carro.PontoNaFila = ""
						carro.Localizacao = pontoID
					}
					
					conn.Write([]byte(fmt.Sprintf("Inicie recarga|%s\n", pontoID)))
					return
				}
			}

			conn.Write([]byte(fmt.Sprintf("Carro %s: Aguarde na fila\n", carroID)))

		case "Carro precisa recarga":
			carro, exists := models.CarrosEstado[id]

			if !exists {
				carro = &models.Carro{ID: id}
				models.CarrosEstado[id] = carro
			}
		
			// Registra a conexão do carro
			conexoesCarros[id] = conn

			pontoMaisProximo, _, distancia := models.Dijkstra("1")
			if pontoMaisProximo == "" {
				conn.Write([]byte("Nenhum ponto disponível\n"))
				mu.Unlock()
				return
			}

			disponivel, _ := models.PontoManagerInstance.GetDisponivel(pontoMaisProximo)
			if disponivel {
				// Se o ponto estiver disponível, envia o comando para o carro se dirigir ao ponto
				models.PontoManagerInstance.SetDisponivel(pontoMaisProximo, false)
				conn.Write([]byte(fmt.Sprintf("Carro %s: Dirija-se ao ponto %s (distância: %.2f)\n", id, pontoMaisProximo, distancia)))
		
				carro.Localizacao = pontoMaisProximo
				carro.NaFila = false
				carro.PontoNaFila = ""
			} else {
				// Se não houver ponto disponível, adiciona o carro à fila
				posicao := models.PontoManagerInstance.Filas[pontoMaisProximo].Adicionar(id)
				conn.Write([]byte(fmt.Sprintf("Carro %s: Adicionado à fila do ponto %s (posição: %d)\n", id, pontoMaisProximo, posicao)))
		
				carro.NaFila = true
				carro.PontoNaFila = pontoMaisProximo
			}
		default:
			conn.Write([]byte("Comando desconhecido\n"))
		}
		mu.Unlock()
	}
}

func main() {
	models.InicializarGrafo()

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Servidor Nuvem iniciado na porta 8080")
	fmt.Println("Aguardando conexões...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		conn.SetDeadline(time.Now().Add(5 * time.Minute))
		go handleConnection(conn)
	}
}