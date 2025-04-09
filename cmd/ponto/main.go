package main

import (
    "fmt"
    "log"
    "net"
    "os"
    "strings"
    "time"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Uso: ./ponto <ID do ponto>")
        os.Exit(1)
    }

    pontoID := os.Args[1]
    if pontoID == "0" || pontoID == "" {
        log.Println("❌ ID do ponto inválido")
        return
    }

    fmt.Printf("🚩 Iniciando ponto %s...\n", pontoID)

    registrado := false // Flag para verificar se o ponto já foi registrado

    for {
        if !registrado {
            conn, err := net.Dial("tcp", "nuvem:8080")
            if err != nil {
                fmt.Printf("⚠️  [%s] Falha ao conectar na nuvem: %v\n", pontoID, err)
                time.Sleep(10 * time.Second)
                continue
            }

            message := fmt.Sprintf("Registrar ponto|%s\n", pontoID)
            _, err = conn.Write([]byte(message))
            if err != nil {
                fmt.Printf("⚠️  [%s] Erro ao enviar mensagem: %v\n", pontoID, err)
                conn.Close()
                time.Sleep(10 * time.Second)
                continue
            }

            buffer := make([]byte, 1024)
            n, err := conn.Read(buffer)
            if err != nil {
                fmt.Printf("⚠️  [%s] Erro ao ler resposta da nuvem: %v\n", pontoID, err)
                conn.Close()
                time.Sleep(10 * time.Second)
                continue
            }

            resposta := string(buffer[:n])
            // Verifica a resposta real do servidor
            if strings.Contains(resposta, "registrado e conectado") {
                fmt.Printf("✅ [%s] Registrado com sucesso\n", pontoID)
                registrado = true // Marca o ponto como registrado
            } else {
                fmt.Printf("❌ [%s] Falha no registro: %s\n", pontoID, resposta)
            }

            conn.Close()
        } else {
            // Após o registro, o ponto pode realizar outras tarefas, como monitoramento
            //fmt.Printf("ℹ️  [%s] Ponto já registrado. Monitorando...\n", pontoID)
            time.Sleep(30 * time.Second) // Intervalo maior para evitar sobrecarga
        }
    }
}