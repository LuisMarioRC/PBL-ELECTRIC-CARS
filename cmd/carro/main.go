package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/LuisMarioRC/PBL-ELECTRIC-CARS/cmd/models"
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

	// Conecta à nuvem na porta 8080
	conn, err := net.Dial("tcp", "nuvem:8080") // Alterado de "localhost:8080" para "nuvem:8080"
	if err != nil {
		log.Fatalf("Erro ao conectar à nuvem: %v", err)
	}
	defer conn.Close() // Garante que a conexão será fechada ao final do programa
	fmt.Printf("Carro %s conectado à nuvem\n", carroID)

	carro := &models.Carro{
		ID:      carroID,
		Bateria: 100.0, // Bateria inicial
	}
	models.CarrosEstado[carroID] = carro // Adiciona o carro ao mapa de estado

	for {
		// Simula o consumo de bateria
		carro.AtualizarBateria()

		// Se a bateria estiver baixa, solicita recarga
		if carro.Bateria <= 20 {
			fmt.Printf("Carro %s com bateria baixa (%.2f%%). Solicitando recarga...\n", carroID, carro.Bateria)

			// Envia mensagem para a nuvem solicitando recarga
			message := fmt.Sprintf("Carro precisa recarga|%s", carroID)
			_, err := conn.Write([]byte(message))
			if err != nil {
				fmt.Println("Erro ao enviar mensagem para a nuvem:", err)
				time.Sleep(10 * time.Second) // Aguarda antes de tentar novamente
				continue
			}

			// Aguarda resposta da nuvem
			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Println("Erro ao ler resposta da nuvem:", err)
				time.Sleep(10 * time.Second) // Aguarda antes de tentar novamente
				continue
			}

			// Processa a resposta da nuvem
			resposta := string(buffer[:n])
			fmt.Printf("Resposta da nuvem para o carro %s: %s\n", carroID, resposta)

			// Se a resposta indicar que o carro pode recarregar, simula o processo de recarga
			if resposta == fmt.Sprintf("Carro %s: Dirija-se ao ponto", carroID) {
				recarregar(carroID, conn)
			}
		}

		time.Sleep(10 * time.Second) // Aguarda antes de verificar novamente
	}
}

// Função para simular o tempo de recarga
func recarregar(carroID string, conn net.Conn) {
	// Define um tempo aleatório de recarga entre 10 e 20 segundos
	tempoRecarga := rand.Intn(10) + 10
	fmt.Printf("Carro %s está recarregando por %d segundos...\n", carroID, tempoRecarga)

	// Aguarda o tempo de recarga
	time.Sleep(time.Duration(tempoRecarga) * time.Second)

	// Envia mensagem para a nuvem indicando que a recarga foi concluída
	message := fmt.Sprintf("Carro recarregado|%s", carroID)
	conn.Write([]byte(message))
	fmt.Printf("Carro %s terminou a recarga\n", carroID)
}
