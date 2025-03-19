package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	// Conecta ao servidor
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Digite uma mensagem para o servidor:")
	reader := bufio.NewReader(os.Stdin)
	message, _ := reader.ReadString('\n')

	// Envia a mensagem ao servidor
	fmt.Fprint(conn, message)

	// Lê a resposta do servidor
	response, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Println("Resposta do servidor:", response)
}