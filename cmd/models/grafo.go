package models

import (
	"fmt"
	"math"
	"sync"
)

type FilaRecarga struct {
	mu     sync.Mutex
	Carros []string
	notify chan struct{}
}

func (f *FilaRecarga) Adicionar(carroID string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Carros = append(f.Carros, carroID)
	return len(f.Carros)
}

func (f *FilaRecarga) Remover() (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.Carros) == 0 {
		return "", false
	}
	carroID := f.Carros[0]
	f.Carros = f.Carros[1:]
	return carroID, true
}

func (f *FilaRecarga) Tamanho() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.Carros)
}

type PontoManager struct {
	mu                sync.Mutex
	pontosDisponiveis map[string]bool
	Filas             map[string]*FilaRecarga  // Alterado de filas para Filas (tornando público)
}

func (pm *PontoManager) SetDisponivel(id string, disponivel bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.pontosDisponiveis[id] = disponivel
	if disponivel {
		if fila, ok := pm.Filas[id]; ok {  // Alterado de pm.filas para pm.Filas
			select {
			case fila.notify <- struct{}{}:
			default:
			}
		}
	}
}

func (pm *PontoManager) GetDisponivel(id string) (bool, bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	disponivel, exists := pm.pontosDisponiveis[id]
	return disponivel, exists
}

func (pm *PontoManager) Init() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.pontosDisponiveis = make(map[string]bool)
	pm.Filas = make(map[string]*FilaRecarga)  // Alterado de pm.filas para pm.Filas
	for id := range GraphInstance.Nodes {
		pm.Filas[id] = &FilaRecarga{  // Alterado de pm.filas para pm.Filas
			notify: make(chan struct{}, 1),
		}
	}
}

type Graph struct {
	Nodes map[string]map[string]float64
	mu    sync.Mutex
}

var (
	GraphInstance        = Graph{Nodes: make(map[string]map[string]float64)}
	PontoManagerInstance = &PontoManager{
		pontosDisponiveis: make(map[string]bool),
		Filas:             make(map[string]*FilaRecarga),  // Alterado de filas para Filas
	}
	Mutex sync.Mutex
)

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
	g.Nodes[to][from] = weight
}

func Dijkstra(start string) (string, []string, float64) {
    Mutex.Lock()
    defer Mutex.Unlock()

    dist := make(map[string]float64)
    prev := make(map[string]string)
    unvisited := make(map[string]bool)

    // Inicializa as distâncias e os nós não visitados
    for node := range GraphInstance.Nodes {
        dist[node] = math.Inf(1)
        unvisited[node] = true
    }
    dist[start] = 0

    // Algoritmo de Dijkstra para calcular as menores distâncias
    for len(unvisited) > 0 {
        var current string
        minDist := math.Inf(1)
        for node := range unvisited {
            if dist[node] < minDist {
                minDist = dist[node]
                current = node
            }
        }
        delete(unvisited, current)

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

	minDistDisponivel := math.Inf(1)
minDistOcupado := math.Inf(1)
pontoMaisProximoDisponivel := ""
pontoMaisProximoOcupado := ""
var filaMaisProxima []string

for ponto := range GraphInstance.Nodes {
    disponivel, exists := PontoManagerInstance.GetDisponivel(ponto)
    if !exists {
        continue
    }

    if disponivel {
        if dist[ponto] < minDistDisponivel {
            minDistDisponivel = dist[ponto]
            pontoMaisProximoDisponivel = ponto
        }
    } else {
        filaTamanho := PontoManagerInstance.Filas[ponto].Tamanho()
        if pontoMaisProximoOcupado == "" || filaTamanho < len(filaMaisProxima) || (filaTamanho == len(filaMaisProxima) && dist[ponto] < minDistOcupado) {
            minDistOcupado = dist[ponto]
            pontoMaisProximoOcupado = ponto
            filaMaisProxima = PontoManagerInstance.Filas[ponto].Carros
        }
    }
}

if pontoMaisProximoDisponivel != "" {
    return pontoMaisProximoDisponivel, nil, minDistDisponivel
}
if pontoMaisProximoOcupado != "" {
    return pontoMaisProximoOcupado, filaMaisProxima, minDistOcupado
}

return "", nil, 0
}

func InicializarGrafo() {
	GraphInstance.AddEdge("1", "2", 10)
	GraphInstance.AddEdge("2", "1", 20)

	PontoManagerInstance.Init()
	PontoManagerInstance.SetDisponivel("1", true)
	PontoManagerInstance.SetDisponivel("2", true)

	fmt.Println("Grafo inicializado com os pontos e conexões:")
	for from, neighbors := range GraphInstance.Nodes {
		for to, weight := range neighbors {
			fmt.Printf("- %s -> %s (peso: %.2f)\n", from, to, weight)
		}
	}
}

func LiberarPonto(localizacao string) {
	PontoManagerInstance.SetDisponivel(localizacao, true)
	fmt.Printf("📢 Ponto %s liberado\n", localizacao)
}

func FecharPonto(localizacao string) {
	PontoManagerInstance.SetDisponivel(localizacao, false)
	fmt.Printf("🔒 Ponto %s ocupado\n", localizacao)
}