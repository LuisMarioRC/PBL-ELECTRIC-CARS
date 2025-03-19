package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "nuvem:8080")
	if err != nil {
		fmt.Println("Erro ao conectar na nuvem:", err)
		os.Exit(1)
	}
	defer conn.Close()

	message := "Ponto de recarga disponível"
	conn.Write([]byte(message))

	buffer := make([]byte, 1024)
	n, _ := conn.Read(buffer)
	fmt.Println("Resposta da nuvem:", string(buffer[:n]))
}
