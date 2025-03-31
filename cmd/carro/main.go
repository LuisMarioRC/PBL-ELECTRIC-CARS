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

    carro := &models.Carro{
        ID:      carroID,
        Bateria: 100.0, // Bateria inicial
    }
    models.CarrosEstado[carroID] = carro // Adiciona o carro ao mapa de estado

    for {
        carro.AtualizarBateria()
        time.Sleep(10 * time.Second)
    }
}

// Função para simular o tempo de recarga
func recarregar(carroID string, conn net.Conn, estado *string) {
    // Define um tempo aleatório de recarga entre 10 e 10 segundos
    tempoRecarga := rand.Intn(10) + 10
    fmt.Printf("Carro %s está recarregando por %d segundos...\n", carroID, tempoRecarga)

    // Aguarda o tempo de recarga
    time.Sleep(time.Duration(tempoRecarga) * time.Second)

    // Envia mensagem para a nuvem indicando que a recarga foi concluída
    message := fmt.Sprintf("Carro recarregado|%s", carroID)
    conn.Write([]byte(message))
    fmt.Printf("Carro %s terminou a recarga\n", carroID)

    // Atualiza o estado para "livre"
    *estado = "livre"
}