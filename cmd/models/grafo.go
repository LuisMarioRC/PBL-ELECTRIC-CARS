package models

import (
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
func Dijkstra(start string) (string, int) {
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
			return ponto, FilaEspera[ponto]
		}
	}
	return "", 0
}
