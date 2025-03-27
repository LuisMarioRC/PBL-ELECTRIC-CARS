package main

import (
    "fmt"
    "log"
    "net"
    "os"
    "time"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Uso: ./carro <ID do carro>")
        os.Exit(1)
    }

    carroID := os.Args[1] 
    if carroID == "0" || carroID == "" {
        log.Println("Erro: ID do carro inválido")
        return
    }

    fmt.Printf("Carro %s iniciado\n", carroID)

    for {
        conn, err := net.Dial("tcp", "nuvem:8080")
        if err != nil {
            fmt.Println("Erro ao conectar na nuvem:", err)
            os.Exit(1)
        }

      
        message := fmt.Sprintf("Carro precisa recarga|%s", carroID)
        conn.Write([]byte(message))

        buffer := make([]byte, 1024)
        n, _ := conn.Read(buffer)
        fmt.Println("Resposta da nuvem:", string(buffer[:n]))

        conn.Close()

        
        time.Sleep(10 * time.Second)
    }
}