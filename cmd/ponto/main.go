package main

import (
    "fmt"
    "net"
    "os"
    "time"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Uso: ./ponto <ID do ponto>")
        os.Exit(1)
    }

    pontoID := os.Args[1]

    for {
        conn, err := net.Dial("tcp", "nuvem:8080")
        if err != nil {
            fmt.Println("Erro ao conectar na nuvem:", err)
            time.Sleep(10 * time.Second) // Aguarda antes de tentar novamente
            continue
        }

        // Envia mensagem no formato esperado pelo servidor
        message := fmt.Sprintf("Ponto de recarga disponível|%s", pontoID)
        _, err = conn.Write([]byte(message))
        if err != nil {
            fmt.Println("Erro ao enviar mensagem:", err)
            conn.Close()
            time.Sleep(10 * time.Second) // Aguarda antes de tentar novamente
            continue
        }

        buffer := make([]byte, 1024)
        n, err := conn.Read(buffer)
        if err != nil {
            fmt.Println("Erro ao ler resposta da nuvem:", err)
            conn.Close()
            time.Sleep(10 * time.Second) // Aguarda antes de tentar novamente
            continue
        }

        fmt.Println("Resposta da nuvem:", string(buffer[:n]))
        conn.Close()

        // Aguarda 10 segundos antes de enviar a próxima atualização
        time.Sleep(10 * time.Second)
    }
}