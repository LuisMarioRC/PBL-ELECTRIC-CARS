package main

import (
    "fmt"
    "github.com/LuisMarioRC/PBL-ELECTRIC-CARS/cmd/models"
    "net"
    "strings"
    "sync"
)

var mu sync.Mutex

func handleConnection(conn net.Conn) {
    defer conn.Close()
    buffer := make([]byte, 1024)

    for {
        n, err := conn.Read(buffer)
        if err != nil {
            return
        }

        mensagem := string(buffer[:n])
        fmt.Printf("Recebido: %s\n", mensagem)

        mu.Lock()
        if !strings.Contains(mensagem, "|") {
            if mensagem == "Ponto de recarga disponível" {
                // conn.Write([]byte("Erro: Não é permitido registrar novos pontos de recarga"))
            } else if strings.HasPrefix(mensagem, "Carro recarregado|") {
                // Carro terminou de recarregar
                parts := strings.Split(mensagem, "|")
                if len(parts) < 2 {
                    conn.Write([]byte("Formato inválido da mensagem"))
                    mu.Unlock()
                    continue
                }
                ponto := parts[1]
                if _, exists := models.PontosDisponiveis[ponto]; exists {
                    // Libera o ponto e atende o próximo carro na fila
                    models.PontosDisponiveis[ponto] = true
                    if models.FilaEspera[ponto] > 0 {
                        models.FilaEspera[ponto]--
                        conn.Write([]byte(fmt.Sprintf("Ponto %s liberado. Próximo carro na fila será atendido.", ponto)))
                    } else {
                        conn.Write([]byte(fmt.Sprintf("Ponto %s liberado. Nenhum carro na fila.", ponto)))
                    }
                } else {
                    conn.Write([]byte(fmt.Sprintf("Erro: Ponto %s não encontrado.", ponto)))
                }
            } else {
                conn.Write([]byte("Formato inválido da mensagem"))
            }
            mu.Unlock()
            continue
        }

        parts := strings.Split(mensagem, "|")
        if len(parts) < 2 {
            conn.Write([]byte("Formato inválido da mensagem"))
            mu.Unlock()
            continue
        }
        comando := parts[0]
        carroID := parts[1]

		if comando == "Carro precisa recarga" {
			if carroID == "" {
				conn.Write([]byte("Formato inválido da mensagem: ID do carro está vazio"))
				mu.Unlock()
				continue
			}
		
			// Usa Dijkstra para encontrar o ponto mais próximo
			destino, fila, distancia := models.Dijkstra("Ponto-1") // Substitua "Ponto-1" pelo ponto de partida do carro, se aplicável
		
			if destino != "" {
				if fila == 0 {
					// Ponto disponível: atribuir ao carro
					models.PontosDisponiveis[destino] = false
					conn.Write([]byte(fmt.Sprintf("Carro %s: Dirija-se ao ponto: %s (distância: %.2f)", carroID, destino, distancia)))
				} else {
					// Ponto ocupado: adicionar o carro à fila de espera
					models.FilaEspera[destino]++
					posicaoFila := models.FilaEspera[destino]
					conn.Write([]byte(fmt.Sprintf("Carro %s: Todos os pontos estão ocupados. Você está na fila do %s: posição %d", carroID, destino, posicaoFila)))
				}
			} else {
				conn.Write([]byte(fmt.Sprintf("Carro %s: Nenhum ponto disponível", carroID)))
			}
		}
        mu.Unlock()
    }
}

func main() {
    // Inicializa o grafo com os pontos fixos e conexões
    models.InicializarGrafo()

    // Inicia o servidor
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        panic(err)
    }
    defer listener.Close()

    fmt.Println("Servidor Nuvem iniciado na porta 8080")

    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        go handleConnection(conn)
    }
}