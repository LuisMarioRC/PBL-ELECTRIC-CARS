package models

import (
	"fmt"
	"math"
	"sync"
)

// Graph representa um grafo com pesos nas arestas
type Graph struct {
	Nodes map[string]map[string]float64 // Lista de adjacência
	mu    sync.Mutex
}

// Criando o grafo globalmente
var (
	GraphInstance = Graph{Nodes: make(map[string]map[string]float64)}
	PontosDisponiveis = make(map[string]bool)
    FilaEspera        = make(map[string][]string)
	Mutex             sync.Mutex
)

// AddEdge adiciona uma aresta ao grafo
func (g *Graph) AddEdge(from, to string, weight float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.Nodes[from] == nil {
		g.Nodes[from] = make(map[string]float64)
	}
	g.Nodes[from][to] = weight
	if g.Nodes[to] == nil {
		g.Nodes[to] = make(map[string]float64)
	}
	g.Nodes[to][from] = weight // Estradas bidirecionais
}

// dijkstra encontra o ponto de recarga disponível mais próximo
func Dijkstra(start string) (string, []string, float64) {
    dist := make(map[string]float64)
    prev := make(map[string]string)
    unvisited := make(map[string]bool)

    // Inicializa as distâncias e os nós não visitados
    for node := range GraphInstance.Nodes {
        dist[node] = math.Inf(1) // Define todas as distâncias como infinito
        unvisited[node] = true   // Marca todos os nós como não visitados
    }
    dist[start] = 0 // A distância para o nó inicial é 0

    for len(unvisited) > 0 {
        // Encontra o nó não visitado com a menor distância
        var current string
        minDist := math.Inf(1)
        for node := range unvisited {
            if dist[node] < minDist {
                minDist = dist[node]
                current = node
            }
        }

        // Remove o nó atual dos não visitados
        delete(unvisited, current)

        // Atualiza as distâncias para os vizinhos do nó atual
        for neighbor, weight := range GraphInstance.Nodes[current] {
            if _, ok := unvisited[neighbor]; ok {
                alt := dist[current] + weight
                if alt < dist[neighbor] {
                    dist[neighbor] = alt
                    prev[neighbor] = current
                }
            }
        }
    }

    // Encontra o ponto disponível mais próximo
    minDist := math.Inf(1)
    pontoMaisProximo := ""
    var filaMaisProxima []string

    for ponto, disponivel := range PontosDisponiveis {
        if disponivel && dist[ponto] < minDist {
            minDist = dist[ponto]
            pontoMaisProximo = ponto
            filaMaisProxima = FilaEspera[ponto]
        }
    }

    if pontoMaisProximo != "" {
        fmt.Printf("Ponto disponível mais próximo: %s (distância: %.2f)\n", pontoMaisProximo, minDist)
        return pontoMaisProximo, filaMaisProxima, minDist
    }
    return "", nil, 0 // Caso nenhum ponto esteja disponível
}

func InicializarGrafo() {
    // Adiciona as conexões entre os pontos com pesos (distâncias)
    GraphInstance.AddEdge("1", "2", 10)
    GraphInstance.AddEdge("2", "3", 20)
    GraphInstance.AddEdge("1", "3", 30)
    GraphInstance.AddEdge("3", "1", 40)

    // Inicializa os pontos como disponíveis
    PontosDisponiveis["1"] = true
    PontosDisponiveis["2"] = true


    // Inicializa as filas de espera para cada ponto
    FilaEspera["1"] = []string{}
    FilaEspera["2"] = []string{}

    fmt.Println("Grafo inicializado com os seguintes pontos e conexões:")
    for from, neighbors := range GraphInstance.Nodes {
        for to, weight := range neighbors {
            fmt.Printf("- %s -> %s (peso: %.2f)\n", from, to, weight)
        }
    }
}