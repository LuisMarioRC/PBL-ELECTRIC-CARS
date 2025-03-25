package main

import (
	"fmt"
	"net"
	"sync"
	"github.com/LuisMarioRC/PBL-ELECTRIC-CARS/cmd/models" 
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
		if mensagem == "Ponto de recarga disponível" {
			ponto := conn.RemoteAddr().String()
			models.PontosDisponiveis[ponto] = true  // Usando o prefixo do pacote models
			models.FilaEspera[ponto] = 0            // Resetando a fila quando um ponto fica disponível
			conn.Write([]byte("Ponto registrado na nuvem"))
		} else if mensagem == "Carro precisa recarga" {
			destino, fila := models.Dijkstra(conn.RemoteAddr().String())  // Chamando Dijkstra com o prefixo 'models'
			if destino != "" {
				models.PontosDisponiveis[destino] = false  // Marca o ponto como não disponível
				models.FilaEspera[destino]++               // Incrementa a fila de espera
				conn.Write([]byte(fmt.Sprintf("Dirija-se ao ponto: %s | Carros na fila: %d", destino, fila)))
			} else {
				conn.Write([]byte("Nenhum ponto disponível no momento"))
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
