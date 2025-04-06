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
	FilaEspera        = make(map[string]int)
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
func Dijkstra(start string) (string, int, float64) {
    dist := make(map[string]float64)
    prev := make(map[string]string)
    unvisited := make(map[string]bool)

    for node := range GraphInstance.Nodes {
        dist[node] = math.Inf(1)
        unvisited[node] = true
    }
    dist[start] = 0

    for len(unvisited) > 0 {
        minNode := ""
        minDist := math.Inf(1)
        for node := range unvisited {
            if dist[node] < minDist {
                minDist = dist[node]
                minNode = node
            }
        }
        if minNode == "" {
            break
        }
        delete(unvisited, minNode)

        for neighbor, weight := range GraphInstance.Nodes[minNode] {
            if !unvisited[neighbor] {
                continue
            }
            alt := dist[minNode] + weight
            if alt < dist[neighbor] {
                dist[neighbor] = alt
                prev[neighbor] = minNode
            }
        }
    }

    for ponto, disponivel := range PontosDisponiveis {
        if disponivel {
            fmt.Printf("Ponto disponível mais próximo: %s (distância: %.2f)\n", ponto, dist[ponto])
            return ponto, FilaEspera[ponto], dist[ponto]
        }
    }
    return "", 0, 0
}

func InicializarGrafo() {
    // Adiciona as conexões entre os pontos com pesos (distâncias)
    GraphInstance.AddEdge("1", "2", 10)
    GraphInstance.AddEdge("1", "3", 15)
    GraphInstance.AddEdge("2", "4", 12)
    GraphInstance.AddEdge("3", "4", 10)
    GraphInstance.AddEdge("4", "5", 5)
    GraphInstance.AddEdge("2", "5", 20)

    // Inicializa os pontos como disponíveis
    PontosDisponiveis["1"] = true
    PontosDisponiveis["2"] = true
    PontosDisponiveis["3"] = true
    PontosDisponiveis["4"] = true
    PontosDisponiveis["5"] = true

    // Inicializa as filas de espera para cada ponto
    FilaEspera["1"] = 0
    FilaEspera["2"] = 0
    FilaEspera["3"] = 0
    FilaEspera["4"] = 0
    FilaEspera["5"] = 0

    fmt.Println("Grafo inicializado com os seguintes pontos e conexões:")
    for from, neighbors := range GraphInstance.Nodes {
        for to, weight := range neighbors {
            fmt.Printf("- %s -> %s (peso: %.2f)\n", from, to, weight)
        }
    }
}
