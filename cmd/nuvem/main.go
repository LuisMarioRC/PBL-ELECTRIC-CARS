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
				ponto := conn.RemoteAddr().String()
				models.PontosDisponiveis[ponto] = true
				models.FilaEspera[ponto] = 0
				conn.Write([]byte("Ponto registrado na nuvem"))
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

		if comando == "Ponto de recarga disponível" {
			ponto := conn.RemoteAddr().String()
			models.PontosDisponiveis[ponto] = true
			models.FilaEspera[ponto] = 0
			conn.Write([]byte("Ponto registrado na nuvem"))
		} else if comando == "Carro precisa recarga" {
			if carroID == "" {
				conn.Write([]byte("Formato inválido da mensagem: ID do carro está vazio"))
				mu.Unlock()
				continue
			}

			destino, fila := models.Dijkstra(conn.RemoteAddr().String())

			if destino != "" {
				models.PontosDisponiveis[destino] = false
				models.FilaEspera[destino]++
				conn.Write([]byte(fmt.Sprintf("Carro %s: Dirija-se ao ponto: %s | Carros na fila: %d", carroID, destino, fila)))
			} else {
				conn.Write([]byte(fmt.Sprintf("Carro %s: Nenhum ponto disponível", carroID)))
			}
		}
		mu.Unlock()
	}
}

func main() {
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
