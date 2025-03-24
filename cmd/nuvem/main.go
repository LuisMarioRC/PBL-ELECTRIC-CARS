package main

import (
	"fmt"
	"net"
	"sync"
)

var pontosDisponiveis = make(map[string]bool) // Armazena pontos de recarga disponíveis
var mu sync.Mutex                             // Para evitar problemas de concorrência

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
			// Armazena o ponto de recarga disponível
			pontosDisponiveis[conn.RemoteAddr().String()] = true
			conn.Write([]byte("Ponto registrado na nuvem"))
		} else if mensagem == "Carro precisa recarga" {
			// Encontra um ponto disponível e o atribui ao carro
			for ponto, disponivel := range pontosDisponiveis {
				if disponivel {
					pontosDisponiveis[ponto] = false // O ponto agora está ocupado
					conn.Write([]byte("Dirija-se ao ponto: " + ponto))
					mu.Unlock()
					return
				}
			}
			conn.Write([]byte("Nenhum ponto disponível no momento"))
		}
		mu.Unlock()
	}
}
