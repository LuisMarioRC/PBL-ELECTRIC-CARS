package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	for {
		conn, err := net.Dial("tcp", "nuvem:8080")
		if err != nil {
			fmt.Println("Erro ao conectar na nuvem:", err)
			os.Exit(1)
		}

		message := "Ponto de recarga disponível"
		conn.Write([]byte(message))

		buffer := make([]byte, 1024)
		n, _ := conn.Read(buffer)
		fmt.Println("Resposta da nuvem:", string(buffer[:n]))

		conn.Close()

		// Aguarda 10 segundos antes de enviar a próxima atualização
		time.Sleep(10 * time.Second)
	}
}
