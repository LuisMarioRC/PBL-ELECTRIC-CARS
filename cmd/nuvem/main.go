package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/LuisMarioRC/PBL-ELECTRIC-CARS/cmd/models"
)

var mu sync.Mutex


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
		fmt.Printf("🤖 Recebido comando: %s, ID: %s\n", comando, id)

		switch comando {
		case "Registrar ponto":
			models.Mutex.Lock()
			defer models.Mutex.Unlock()

			if _, exists := models.GraphInstance.Nodes[id]; !exists {
				conn.Write([]byte(fmt.Sprintf("Erro: Ponto %s não está registrado no grafo da nuvem.\n", id)))
				return
			}

			if _, exists := models.PontosDisponiveis[id]; !exists {
				models.PontosDisponiveis[id] = true
				fmt.Printf("Ponto %s registrado com sucesso.\n", id)
				conn.Write([]byte(fmt.Sprintf("Ponto %s registrado com sucesso e conectado ao sistema.\n", id)))
			} else {
				models.PontosDisponiveis[id] = true
				conn.Write([]byte(fmt.Sprintf("Ponto %s já estava registrado. Atualizado como disponível e conectado.\n", id)))
			}

		case "Carro recarregado":
			models.Mutex.Lock()

			if _, exists := models.GraphInstance.Nodes[id]; !exists {
				models.Mutex.Unlock()
				conn.Write([]byte(fmt.Sprintf("Erro: Ponto %s não está registrado no grafo da nuvem.\n", id)))
				return
			}

			if len(models.FilaEspera[id]) > 0 {
				proximoCarro := models.FilaEspera[id][0]
				models.FilaEspera[id] = models.FilaEspera[id][1:]
			
				models.PontosDisponiveis[id] = false
			
				var carro *models.Carro
				if c, existe := models.CarrosEstado[proximoCarro]; existe {
					carro = c
					carro.NaFila = false
					carro.PontoNaFila = ""
					carro.Localizacao = id
				} else {
					carro = &models.Carro{
						ID:          proximoCarro,
						NaFila:      false,
						PontoNaFila: "",
						Localizacao: id,
					}
					models.CarrosEstado[proximoCarro] = carro
				}
			
				// Envia mensagem para o carro informando que ele está carregando
				fmt.Printf("Ponto %s: Próximo carro %s entrou para carregar.\n", id, proximoCarro)
				conn.Write([]byte(fmt.Sprintf("Carro %s: É sua vez de carregar no ponto %s!\n", proximoCarro, id)))
			
				models.Mutex.Unlock()
			
				// Simula o tempo de recarga
				tempoRecarga := 10 + rand.Intn(10) // Tempo de recarga entre 10 e 20 segundos
				fmt.Printf("Ponto %s: Carro %s está recarregando por %d segundos...\n", id, proximoCarro, tempoRecarga)
				time.Sleep(time.Duration(tempoRecarga) * time.Second)
			
				// Atualiza o nível da bateria do carro
				carro.Bateria = 100
				fmt.Printf("Ponto %s: Carro %s terminou a recarga. Bateria: %.2f%%\n", id, proximoCarro, carro.Bateria)
			
				// Libera o ponto após a recarga
				models.Mutex.Lock()
				models.PontosDisponiveis[id] = true

				//reativando o carro pra voltar ao clico de solicitacao do carregament
				carro.SolicitarRecarga()
				models.Mutex.Unlock()
			
				fmt.Printf("Ponto %s: Liberado após recarga do carro %s.\n", id, proximoCarro)
			} else {
				models.PontosDisponiveis[id] = true
				models.Mutex.Unlock()
				fmt.Printf("Ponto %s liberado. Nenhum carro na fila.\n", id)
			}

		case "Verificar fila":
			if len(parts) < 3 {
				conn.Write([]byte("Formato inválido: Verificar fila|IDPonto|IDCarro\n"))
				return
			}

			pontoID := id       // ID do ponto
			carroID := parts[2] // ID do carro que está verificando seu status

			models.Mutex.Lock()
			defer models.Mutex.Unlock()

			// Inicializa a fila do ponto, se necessário
			if _, ok := models.FilaEspera[pontoID]; !ok {
				models.FilaEspera[pontoID] = []string{}
			}

			// Verifica se o carro está na fila
			encontrou := false
			posicao := 0
			for idx, carroNaFila := range models.FilaEspera[pontoID] {
				if carroNaFila == carroID {
					encontrou = true
					posicao = idx + 1 // Posição começa em 1
					break
				}
			}

			if encontrou {
				models.Mutex.Lock()
				defer models.Mutex.Unlock()

				if posicao == 1 && models.PontosDisponiveis[pontoID] {
					models.PontosDisponiveis[pontoID] = false
					models.FilaEspera[pontoID] = models.FilaEspera[pontoID][1:]

					// Atualiza o estado do carro
					if carro, existe := models.CarrosEstado[carroID]; existe {
						carro.NaFila = false
						carro.PontoNaFila = ""
						carro.Localizacao = pontoID
					}

					fmt.Printf("Ponto %s: Carro %s é o próximo da fila e o ponto está disponível!\n", pontoID, carroID)
					conn.Write([]byte(fmt.Sprintf("Carro %s: É sua vez de carregar no ponto %s!\n", carroID, pontoID)))
					return
				} else {
					fmt.Printf("Ponto %s: Carro %s está na posição %d da fila.\n", pontoID, carroID, posicao)
					conn.Write([]byte(fmt.Sprintf("Carro %s: Você está na posição %d da fila do ponto %s. Aguarde sua vez.\n", carroID, posicao, pontoID)))
				}
			} else {
				// Verifica se o carro tem um registro no sistema indicando que está na fila
				if carro, existe := models.CarrosEstado[carroID]; existe && carro.NaFila && carro.PontoNaFila == pontoID {
					if disp, existe := models.PontosDisponiveis[pontoID]; existe && disp {
						// O ponto está disponível, o carro pode carregar
						models.PontosDisponiveis[pontoID] = false
						fmt.Printf("Ponto %s: Carro %s está na fila e será recarregado agora.\n", pontoID, carroID)
						conn.Write([]byte(fmt.Sprintf("Carro %s: É sua vez de carregar no ponto %s!\n", carroID, pontoID)))

						// Atualiza o estado do carro
						carro.NaFila = false
						carro.PontoNaFila = ""
						carro.Localizacao = pontoID
					} else {
						fmt.Printf("Ponto %s: Carro %s está na fila, mas o ponto não está disponível.\n", pontoID, carroID)
						conn.Write([]byte(fmt.Sprintf("Carro %s: Aguarde, o ponto %s ainda não está disponível.\n", carroID, pontoID)))
					}
				} else {
					// Adiciona o carro à fila
					models.FilaEspera[pontoID] = append(models.FilaEspera[pontoID], carroID)
					posicao = len(models.FilaEspera[pontoID])

					fmt.Printf("Ponto %s: Carro %s adicionado à fila. Posição: %d\n", pontoID, carroID, posicao)
					conn.Write([]byte(fmt.Sprintf("Carro %s: Você foi adicionado à fila do ponto %s. Posição: %d\n", carroID, pontoID, posicao)))

					// Atualiza o estado do carro
					if _, existe := models.CarrosEstado[carroID]; !existe {
						models.CarrosEstado[carroID] = &models.Carro{ID: carroID}
					}
					carro := models.CarrosEstado[carroID]
					carro.Localizacao = pontoID
					carro.NaFila = true
					carro.PontoNaFila = pontoID
				}
			}

		case "Carro precisa recarga":
			// CORREÇÃO: Primeiro verifica se o carro já está em alguma fila
			carro, exists := models.CarrosEstado[id]
			if exists && carro.Localizacao != "" && carro.NaFila {
				if fila, ok := models.FilaEspera[carro.Localizacao]; ok {
					for idx, carroNaFila := range fila {
						if carroNaFila == id {
							posicao := idx + 1 // Posição começa em 1
							conn.Write([]byte(fmt.Sprintf("Carro %s: Você já está na fila do ponto %s. Posição: %d. Aguarde sua vez.\n", id, carro.Localizacao, posicao)))
							mu.Unlock()
							return
						}
					}
				}
			}

			pontoMaisProximo, _, distancia := models.Dijkstra("1") // isso pode ser parametrizado futuramente

			if models.PontosDisponiveis[pontoMaisProximo] {
				// Ponto mais próximo está disponível
				models.PontosDisponiveis[pontoMaisProximo] = false
				fmt.Printf("Carro %s iniciou recarga no ponto %s (distância: %.2f)\n", id, pontoMaisProximo, distancia)
				conn.Write([]byte(fmt.Sprintf("Carro %s: Dirija-se ao ponto: %s (distância: %.2f)\n", id, pontoMaisProximo, distancia)))

				if !exists {
					carro = &models.Carro{ID: id}
					models.CarrosEstado[id] = carro
				}
				carro.Localizacao = pontoMaisProximo
				carro.NaFila = false
				carro.PontoNaFila = ""
			} else {
				pontoMenorFila := ""
				menorTamanhoFila := math.MaxInt // Inicializa com o maior valor possível

				// Verifica todos os pontos disponíveis no grafo
				for pontoID := range models.GraphInstance.Nodes {
					if _, existe := models.PontosDisponiveis[pontoID]; existe {
						// Se o ponto existe e tem fila menor
						if fila, ok := models.FilaEspera[pontoID]; ok && len(fila) < menorTamanhoFila {
							pontoMenorFila = pontoID
							menorTamanhoFila = len(fila)
						} else if !ok {
							// Se o ponto não tem fila ainda, inicializa com fila vazia
							models.FilaEspera[pontoID] = []string{}
							pontoMenorFila = pontoID
							menorTamanhoFila = 0
						}
					}
				}

				// Verifica se encontrou um ponto válido
				if pontoMenorFila == "" {
					conn.Write([]byte("Erro: Nenhum ponto disponível encontrado.\n"))
					mu.Unlock()
					return
				}

				// Adiciona o carro à fila do ponto com menor espera
				models.Mutex.Lock()
				models.FilaEspera[pontoMenorFila] = append(models.FilaEspera[pontoMenorFila], id) // Adiciona o ID do carro à fila
				models.Mutex.Unlock()

				posicaoFila := len(models.FilaEspera[pontoMenorFila]) // Posição do carro na fila
				fmt.Printf("Carro %s adicionado à fila do ponto %s. Posição na fila: %d\n", id, pontoMenorFila, posicaoFila)
				conn.Write([]byte(fmt.Sprintf("Carro %s: Ponto ocupado. Você está na fila do ponto %s: posição %d\n", id, pontoMenorFila, posicaoFila)))

				if !exists {
					carro = &models.Carro{ID: id}
					models.CarrosEstado[id] = carro
				}
				carro.Localizacao = pontoMenorFila

				// CORREÇÃO: Sinaliza que o carro está em uma fila específica
				carro.NaFila = true
				carro.PontoNaFila = pontoMenorFila
			}

		default:
			conn.Write([]byte("Comando desconhecido\n"))
		}
		mu.Unlock()
	}
}

func main() {
	models.InicializarGrafo()

	// Inicializar as estruturas necessárias
	if models.FilaEspera == nil {
		models.FilaEspera = make(map[string][]string)
	}
	if models.PontosDisponiveis == nil {
		models.PontosDisponiveis = make(map[string]bool)
	}

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
		// Configura um timeout para leitura/escrita
		conn.SetDeadline(time.Now().Add(5 * time.Minute)) // 5 minutos de timeout
		go handleConnection(conn)
	}
}