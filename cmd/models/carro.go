package models

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)


type Carro struct {
    ID          string  `json:"id"`
    Bateria     float64 `json:"bateria"` // Nível da bateria (0-100%)
    Localizacao string  `json:"localizacao"`
    Conectado   bool    `json:"conectado"`
    Debito      float64 `json:"debito"`
    PontoNaFila string  `json:"pontoNaFila"`
    NaFila      bool    `json:"naFila"`
}

// ConsumoPorKm define o consumo de bateria por quilômetro rodado
const ConsumoPorKm = 7.5

var CarrosEstado = make(map[string]*Carro) // Mapa para armazenar o estado dos carros

func (c *Carro) Rodar(distancia float64) error {
    if c.Bateria <= 0 {
        return errors.New("bateria descarregada, recarregue no ponto de recarga")
    }

    consumo := distancia * ConsumoPorKm
    if consumo > c.Bateria {
        return errors.New("bateria insuficiente para completar a distância")
    }

    c.Bateria -= consumo
    CarrosEstado[c.ID] = c // Atualiza o estado no mapa
    fmt.Printf("Carro %s rodou %.2f km. Bateria restante: %.2f%%\n", c.ID, distancia, c.Bateria)
    return nil
}

func (c *Carro) SolicitarRecarga() {
    var conn net.Conn
    var err error

    // Tenta conectar/reconectar à nuvem
    for {
        conn, err = net.Dial("tcp", "nuvem:8080")
        if err != nil {
            fmt.Printf("Carro %s: Erro ao conectar na nuvem: %s. Tentando novamente...\n", c.ID, err.Error())
            time.Sleep(10 * time.Second)
            continue
        }
        break
    }
    defer conn.Close()

    // CORREÇÃO: Verificar se o carro já está numa fila
    if c.NaFila && c.PontoNaFila != "" {
        fmt.Printf("Carro %s: Já estou na fila do ponto %s. Continuando a esperar.\n", c.ID, c.PontoNaFila)
        // Verifica status da fila
        message := fmt.Sprintf("Verificar fila|%s|%s\n", c.PontoNaFila, c.ID)
        _, err = conn.Write([]byte(message))
    } else {
        // Solicitação inicial para um ponto de recarga
        message := fmt.Sprintf("Carro precisa recarga|%s\n", c.ID)
        _, err = conn.Write([]byte(message))
    }
    
    if err != nil {
        fmt.Printf("Carro %s: Erro ao enviar solicitação de recarga: %s\n", c.ID, err.Error())
        return
    }

    // Loop para aguardar enquanto estiver na fila
    for {
        buffer := make([]byte, 1024)
        n, err := conn.Read(buffer)
        if err != nil {
            fmt.Printf("Carro %s: Erro ao receber resposta da nuvem: %s\n", c.ID, err.Error())
            return
        }

        resposta := string(buffer[:n])
        fmt.Printf("Carro %s: Resposta da nuvem: %s\n", c.ID, resposta)

        // Se o carro pode se dirigir ao ponto de recarga
        if strings.Contains(resposta, "Dirija-se ao ponto") {
            // Extrai o ID do ponto da resposta
            partes := strings.Split(resposta, "ponto:")
            if len(partes) > 1 {
                idComDistancia := strings.TrimSpace(partes[1])
                idParte := strings.Split(idComDistancia, "(")[0]
                c.Localizacao = strings.TrimSpace(idParte)
                fmt.Printf("ID do ponto extraído: %s\n", c.Localizacao)
            }

            // Reset do status de fila
            c.NaFila = false
            c.PontoNaFila = ""

            tempoRecarga := rand.Intn(10) + 10
            fmt.Printf("Carro %s está recarregando por %d segundos no ponto %s...\n", c.ID, tempoRecarga, c.Localizacao)
            time.Sleep(time.Duration(tempoRecarga) * time.Second)

            c.Bateria = 100
            fmt.Printf("Carro %s terminou a recarga no ponto %s, e ficou com %.2f%% de carga\n", c.ID, c.Localizacao, c.Bateria)

            // Notifica a nuvem que o carro terminou de recarregar, usando nova conexão
            var connNotify net.Conn
            for i := 0; i < 3; i++ { // Tenta até 3 vezes
                connNotify, err = net.Dial("tcp", "nuvem:8080")
                if err == nil {
                    break
                }
                time.Sleep(2 * time.Second)
            }
            
            if err != nil {
                fmt.Printf("Carro %s: Erro ao conectar para notificar recarga completa: %s\n", c.ID, err.Error())
                return
            }
            
            message := fmt.Sprintf("Carro recarregado|%s\n", c.Localizacao)
            _, err = connNotify.Write([]byte(message))
            connNotify.Close()
            
            if err != nil {
                fmt.Printf("Carro %s: Erro ao enviar mensagem de recarregado: %s\n", c.ID, err.Error())
            }
            
            // Termina aqui, recarga completa
            return
            
        } else if strings.Contains(resposta, "Ponto ocupado") {
            // Extrai o ID do ponto e a posição na fila
            partes := strings.Split(resposta, "fila do ponto")
            if len(partes) > 1 {
                restante := strings.TrimSpace(partes[1])
                idParte := strings.Split(restante, ":")[0]
                pontoFila := strings.TrimSpace(idParte)

                // CORREÇÃO: Salva o ponto onde está na fila
                c.PontoNaFila = pontoFila
                c.NaFila = true
                c.Localizacao = pontoFila

                posicaoParte := strings.Split(restante, "posição")[1]
                posicao := strings.TrimSpace(posicaoParte)

                fmt.Printf("Carro %s está na fila do ponto %s na posição %s\n", c.ID, c.PontoNaFila, posicao)

                // Aguarda um tempo antes de verificar se chegou sua vez
                time.Sleep(15 * time.Second)
                
                // CORREÇÃO: Envia mensagem de verificação para o ponto específico
                var connCheck net.Conn
                connCheck, err = net.Dial("tcp", "nuvem:8080")
                if err != nil {
                    fmt.Printf("Carro %s: Erro ao conectar para verificar status: %s\n", c.ID, err.Error())
                    time.Sleep(5 * time.Second)
                    continue
                }
                
                // CORREÇÃO: Verifica especificamente o status no ponto onde está na fila
                message := fmt.Sprintf("Verificar fila|%s|%s\n", c.PontoNaFila, c.ID)
                _, err = connCheck.Write([]byte(message))
                
                if err != nil {
                    fmt.Printf("Carro %s: Erro ao enviar verificação: %s\n", c.ID, err.Error())
                    connCheck.Close()
                    time.Sleep(5 * time.Second)
                    continue
                }
                
                // Lê a resposta da verificação
                buffer := make([]byte, 1024)
                n, err := connCheck.Read(buffer)
                connCheck.Close()
                
                if err != nil {
                    fmt.Printf("Carro %s: Erro ao ler resposta de verificação: %s\n", c.ID, err.Error())
                    time.Sleep(5 * time.Second)
                    continue
                }
                
                checkResposta := string(buffer[:n])
                fmt.Printf("Carro %s: Resposta da verificação: %s\n", c.ID, checkResposta)
                
                // Se for sua vez de carregar
                if strings.Contains(checkResposta, "sua vez") {
                    fmt.Printf("Carro %s: É minha vez de carregar no ponto %s!\n", c.ID, c.PontoNaFila)
                    
                    tempoRecarga := rand.Intn(10) + 10
                    fmt.Printf("Carro %s está recarregando por %d segundos no ponto %s...\n", c.ID, tempoRecarga, c.PontoNaFila)
                    time.Sleep(time.Duration(tempoRecarga) * time.Second)
    
                    c.Bateria = 100
                    fmt.Printf("Carro %s terminou a recarga no ponto %s, e ficou com %.2f%% de carga\n", c.ID, c.PontoNaFila, c.Bateria)
    
                    // Notifica que terminou a recarga
                    var connNotify net.Conn
                    connNotify, err = net.Dial("tcp", "nuvem:8080")
                    if err != nil {
                        fmt.Printf("Carro %s: Erro ao conectar para notificar recarga: %s\n", c.ID, err.Error())
                        return
                    }
                    
                    message := fmt.Sprintf("Carro recarregado|%s\n", c.PontoNaFila)
                    connNotify.Write([]byte(message))
                    connNotify.Close()
                    
                    // Reset do status de fila
                    c.NaFila = false
                    c.PontoNaFila = ""
                    
                    return // Termina aqui, recarga completa
                } else {
                    // Continua aguardando
                    fmt.Printf("Carro %s: Ainda aguardando na fila do ponto %s\n", c.ID, c.PontoNaFila)
                    time.Sleep(10 * time.Second)
                    continue
                }
            }
        } else if strings.Contains(resposta, "é sua vez") || strings.Contains(resposta, "sua vez") {
            fmt.Printf("Carro %s: Recebi notificação que é minha vez de carregar!\n", c.ID)
        
            // Atualiza o estado do carro
            pontoDeCarga := c.PontoNaFila
            if pontoDeCarga == "" {
                pontoDeCarga = c.Localizacao
            }
        
            c.NaFila = false
            c.PontoNaFila = ""
            c.Localizacao = pontoDeCarga
        
            // Simula o tempo de recarga
            tempoRecarga := rand.Intn(10) + 10
            fmt.Printf("Carro %s está recarregando por %d segundos no ponto %s...\n", c.ID, tempoRecarga, pontoDeCarga)
            time.Sleep(time.Duration(tempoRecarga) * time.Second)
        
            // Atualiza o nível da bateria
            c.Bateria = 100
            fmt.Printf("Carro %s terminou a recarga no ponto %s, e ficou com %.2f%% de carga\n", c.ID, pontoDeCarga, c.Bateria)
        
            // Notifica a nuvem que o carro terminou de recarregar
            var connNotify net.Conn
            connNotify, err = net.Dial("tcp", "nuvem:8080")
            if err != nil {
                fmt.Printf("Carro %s: Erro ao conectar para notificar recarga completa: %s\n", c.ID, err.Error())
                return
            }
        
            message := fmt.Sprintf("Carro recarregado|%s\n", pontoDeCarga)
            _, err = connNotify.Write([]byte(message))
            connNotify.Close()
        
            if err != nil {
                fmt.Printf("Carro %s: Erro ao enviar mensagem de recarregado: %s\n", c.ID, err.Error())
            }
        
            // Reinicia o ciclo de comunicação
            fmt.Printf("Carro %s: Recarga concluída. Voltando ao estado normal.\n", c.ID)
            c.SolicitarRecarga() // Reinicia o ciclo de comunicação
            return
        } else if strings.Contains(resposta, "Você já está na fila") {
            // Se o carro já está na fila, aguarda um tempo e verifica novamente
            partes := strings.Split(resposta, "ponto")
            if len(partes) > 1 {
                pontoID := strings.TrimSpace(strings.Split(partes[1], ".")[0])
                c.PontoNaFila = pontoID
                c.NaFila = true
            }
            
            fmt.Printf("Carro %s: Ainda na fila do ponto %s, aguardando...\n", c.ID, c.PontoNaFila)
            time.Sleep(20 * time.Second)
            
            // CORREÇÃO: Verifica especificamente o status no ponto onde está na fila
            var connCheck net.Conn
            connCheck, err = net.Dial("tcp", "nuvem:8080")
            if err != nil {
                fmt.Printf("Carro %s: Erro ao conectar para verificar status: %s\n", c.ID, err.Error())
                time.Sleep(5 * time.Second)
                continue
            }
            
            message := fmt.Sprintf("Verificar fila|%s|%s\n", c.PontoNaFila, c.ID)
            connCheck.Write([]byte(message))
            connCheck.Close()
            continue
        } else {
            fmt.Printf("⚠️ Resposta não reconhecida: %s\n", resposta)
            // Aguarda um tempo antes de tentar novamente
            time.Sleep(5 * time.Second)
            continue
        }
    }
}

// AtualizarBateria simula o consumo de bateria ao longo do tempo
func (c *Carro) AtualizarBateria() {
    for {
        time.Sleep(10 * time.Second) // Atualiza o estado da bateria a cada 10 segundos

        // Se o carro estiver na fila, não consome bateria
        if c.NaFila {
            fmt.Printf("Carro %s: Na fila do ponto %s, bateria atual: %.2f%%\n", c.ID, c.PontoNaFila, c.Bateria)
            c.SolicitarRecarga()
            return
        }

        // Simula o consumo de bateria
        distancia := rand.Float64() * 10 // Simula uma distância aleatória entre 0 e 10 km
        err := c.Rodar(distancia)
        if err != nil {
            fmt.Printf("Carro %s: %s\n", c.ID, err.Error())
            if c.Bateria <= 25 {
                c.SolicitarRecarga()
            }
            return // Encerra a rotina
        }

        // Verifica se a bateria está em estágio crítico
        if c.Bateria <= 25 {
            fmt.Printf("Carro %s: ALERTA! Bateria em estado crítico (%.2f%%). Dirija-se ao ponto de recarga mais próximo.\n", c.ID, c.Bateria)
            c.SolicitarRecarga()
            return // Encerra a rotina
        }

        fmt.Printf("Carro %s: Bateria atual: %.2f%%\n", c.ID, c.Bateria)
    }
}

// PrecoPorUnidade define o custo por 1% de recarga da bateria
const PrecoPorUnidade = 0.50 // Exemplo: R$ 0,50 por 1%

// CalcularCustoRecarga calcula o custo para recarregar a bateria até 100%
func (c *Carro) CalcularCustoRecarga() float64 {
    bateriaFaltante := 100 - c.Bateria
    custo := bateriaFaltante * PrecoPorUnidade
    return custo
}

func LiberarPonto(localizacao string) {
    Mutex.Lock()
    defer Mutex.Unlock()

    if _, existe := PontosDisponiveis[localizacao]; existe {
        PontosDisponiveis[localizacao] = true
        fmt.Printf("Ponto %s liberado\n", localizacao)
    }
}