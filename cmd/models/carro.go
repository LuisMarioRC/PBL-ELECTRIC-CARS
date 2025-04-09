package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

type Carro struct {
	ID          string
	Bateria     float64
	Localizacao string
	Conectado   bool
	Debito      float64
	PontoNaFila string
	NaFila      bool
}

const ConsumoPorKm = 7.5
var CarrosEstado = make(map[string]*Carro)

func (c *Carro) Rodar(distancia float64) error {
	if c.Bateria <= 0 {
		return errors.New("bateria descarregada")
	}

	consumo := distancia * ConsumoPorKm
	if consumo > c.Bateria {
		return errors.New("Bateria insuficiente para rodar distância solicitada")
	}

	c.Bateria -= consumo
	CarrosEstado[c.ID] = c
	fmt.Printf("Carro %s rodou %.2f km. Bateria: %.2f%%\n", c.ID, distancia, c.Bateria)
	return nil
}

func (c *Carro) SolicitarRecarga() {
    var conn net.Conn
    var err error

	for i := 0; i < 3; i++ {
		conn, err = net.Dial("tcp", "nuvem:8080")
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		fmt.Printf("Carro %s: Falha ao conectar\n", c.ID)
		return
	}
	defer conn.Close()

	if c.NaFila && c.PontoNaFila != "" {
		message := fmt.Sprintf("Verificar fila|%s|%s\n", c.PontoNaFila, c.ID)
		_, err = conn.Write([]byte(message))
	} else {
		message := fmt.Sprintf("Carro precisa recarga|%s\n", c.ID)
		_, err = conn.Write([]byte(message))
	}

	if err != nil {
		fmt.Printf("Carro %s: Erro ao enviar solicitação\n", c.ID)
		return
	}

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Carro %s: Erro ao receber resposta\n", c.ID)
			return
		}

		resposta := string(buffer[:n])
		fmt.Printf("🚗 Recebido resposta: %s\n", resposta)

		if strings.Contains(resposta, "Dirija-se ao ponto") {
			// Extrai o ID do ponto da resposta
			partes := strings.Split(resposta, "ponto")
			if len(partes) > 1 {
				idComDistancia := strings.TrimSpace(partes[1])
				idParte := strings.Split(idComDistancia, "(")[0]
				c.Localizacao = strings.TrimSpace(idParte)

				
			}

			c.NaFila = false
			c.PontoNaFila = ""
		
			// Atualiza o estado do carro na nuvem
			CarrosEstado[c.ID] = c

			// Simula tempo de recarga
			tempoRecarga := rand.Intn(10) + 10
			fmt.Printf("Carro %s recarregando por %d segundos\n", c.ID, tempoRecarga)
			time.Sleep(time.Duration(tempoRecarga) * time.Second)

			c.Bateria = 100
			custo := c.CalcularCustoRecarga()
			fmt.Printf("Carro %s recarga completa. Custo: R$%.2f.\n", c.ID, custo)

			// Notifica a nuvem
			connNotify, err := net.Dial("tcp", "nuvem:8080")
			if err == nil {
				connNotify.Write([]byte(fmt.Sprintf("Pagamento Recebido|%s\n", c.ID)))
				message := fmt.Sprintf("Carro recarregado|%s\n", c.Localizacao)
				connNotify.Write([]byte(message))
				connNotify.Close()
			}
			return

		} else if strings.Contains(resposta, "Adicionado à fila") {
			partes := strings.Split(resposta, "ponto")
			if len(partes) > 1 {
				restante := strings.TrimSpace(partes[1])
				idParte := strings.Split(restante, "(")[0]
				c.PontoNaFila = strings.TrimSpace(idParte)
				c.NaFila = true
			}
			time.Sleep(15 * time.Second)
			continue

		} else if strings.Contains(resposta, "Inicie recarga") {
			fmt.Printf("Carro %s: Recebido comando para iniciar recarga\n", c.ID)
		
			// Extrai a localização do ponto do comando
			partes := strings.Split(resposta, "|")
			if len(partes) > 1 {
				c.Localizacao = strings.TrimSpace(partes[1]) // Define a localização do ponto
			} else {
				fmt.Printf("Carro %s: Localização não encontrada no comando 'Inicie recarga'\n", c.ID)
				return
			}
		
			c.NaFila = false
			c.PontoNaFila = ""
		
			CarrosEstado[c.ID] = c
		
			// Simula tempo de recarga
			tempoRecarga := rand.Intn(10) + 10
			fmt.Printf("Carro %s recarregando por %d segundos\n", c.ID, tempoRecarga)
			time.Sleep(time.Duration(tempoRecarga) * time.Second)
		
			c.Bateria = 100
			custo := c.CalcularCustoRecarga()
			fmt.Printf("Carro %s recarga completa. Custo: R$%.2f.\n", c.ID, custo)
		
			// Notifica a nuvem
			connNotify, err := net.Dial("tcp", "nuvem:8080")
			if err == nil {
				connNotify.Write([]byte(fmt.Sprintf("Pagamento Recebido|%s\n", c.ID)))
				message := fmt.Sprintf("Carro recarregado|%s\n", c.Localizacao)
				connNotify.Write([]byte(message))
				connNotify.Close()
			} else {
				fmt.Printf("Carro %s: Falha ao notificar a nuvem sobre a recarga\n", c.ID)
			}
			return
		}

		time.Sleep(5 * time.Second)
	}
}

func (c *Carro) AtualizarBateria() {
	for {
		time.Sleep(5 * time.Second)

		if c.NaFila {
			c.SolicitarRecarga()
			continue
		}

		distancia := rand.Float64() * 10
		err := c.Rodar(distancia)
		if err != nil {
			fmt.Printf("Carro %s: %s\n", c.ID, err.Error())
			if c.Bateria <= 25 {
				c.SolicitarRecarga()
			}
			continue
		}

		if c.Bateria <= 25 {
			fmt.Printf("Carro %s: Bateria crítica (%.2f%%)\n", c.ID, c.Bateria)
			c.SolicitarRecarga()
			continue
		}
	}
}

func (c *Carro) CalcularCustoRecarga() float64 {
    fmt.Println("Iniciando cálculo de recarga para o carro:", c.ID)

    custo := 10 + rand.Float64()*(80-10)
    custoFormat := fmt.Sprintf("%.2f", custo)

	
	date := time.Now()
	dateFomat := date.Format("2006-01-02 15:04:05")

    // Novo dado a ser adicionado
    novoRegistro := map[string]interface{}{
        "CarroID": c.ID,
		"Ponto de recarga": c.Localizacao,
        "Custo":   custoFormat,
        "Data":    dateFomat,
    }

    err := os.MkdirAll("/app/json", os.ModePerm)
    if err != nil {
        fmt.Printf("Carro %s: Erro ao criar diretório JSON: %s\n", c.ID, err.Error())
        return custo
    }

    filePath := fmt.Sprintf("/app/json/custo_recarga_carro%s.json", c.ID)

    var registros []map[string]interface{}

    // Se o arquivo já existe, lê os registros existentes
    if _, err := os.Stat(filePath); err == nil {
        conteudo, err := os.ReadFile(filePath)
        if err != nil {
            fmt.Printf("Carro %s: Erro ao ler arquivo JSON existente: %s\n", c.ID, err.Error())
        } else {
            err = json.Unmarshal(conteudo, &registros)
            if err != nil {
                fmt.Printf("Carro %s: Erro ao decodificar JSON existente: %s\n", c.ID, err.Error())
            }
        }
    }

    // Adiciona o novo registro
    registros = append(registros, novoRegistro)

    // Converte tudo novamente para JSON formatado
    file, err := json.MarshalIndent(registros, "", "  ")
    if err != nil {
        fmt.Printf("Carro %s: Erro ao salvar custo da recarga: %s\n", c.ID, err.Error())
        return custo
    }

    // Salva no arquivo
    err = os.WriteFile(filePath, file, 0644)
    if err != nil {
        fmt.Printf("Carro %s: Erro ao escrever arquivo JSON: %s\n", c.ID, err.Error())
    } else {
        fmt.Printf("Custo da recarga salvo com sucesso em %s\n", filePath)
    }

    return custo
}


