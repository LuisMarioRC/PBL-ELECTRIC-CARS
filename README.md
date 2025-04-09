<img width=100% src="https://capsule-render.vercel.app/api?type=waving&color=00FFFF&height=120&section=header"/>

<div align="center">  

# Projeto de Concorrência e Conectividade com Docker e Go

## Descrição do Projeto

Este projeto simula um sistema de gerenciamento de carros elétricos que precisam recarregar suas baterias em pontos de recarga distribuídos. O sistema utiliza conceitos de concorrência e conectividade para coordenar a comunicação entre carros, pontos de recarga e uma nuvem central.

## Componentes Principais

1. **Nuvem**: Servidor central que gerencia todos os pontos de recarga e carros
2. **Pontos de Recarga**: Estações que podem estar disponíveis ou ocupadas
3. **Carros**: Veículos elétricos que consomem bateria e solicitam recargas

## Tecnologias Utilizadas

- **Linguagem**: Go (Golang)
- **Contêinerização**: Docker e Docker Compose
- **Protocolo de Comunicação**: TCP/IP
- **Sincronização**: Mutexes e canais

## Como Executar o Projeto

### Pré-requisitos

- Docker instalado
- Docker Compose instalado

### Passos para Execução

1. Clone o repositório do projeto
2. Navegue até o diretório do projeto
3. Execute o seguinte comando:

```bash
docker-compose up --build
```

Isso irá construir e iniciar todos os serviços definidos no docker-compose.yml.

## Estrutura de Arquivos

- `main.go`: Ponto de entrada para cada componente (nuvem, ponto, carro)
- `carro.go`: Lógica do carro e consumo de bateria
- `grafo.go`: Implementação do grafo para gerenciamento de pontos
- `ponto.go`: Definição da estrutura do ponto de recarga
- `recarga.go`: Definição da estrutura de recarga
- `docker-compose.yml`: Configuração dos serviços Docker

## Tratamento de Concorrência

O projeto utiliza várias técnicas para lidar com concorrência:

1. **Mutexes**: Protegem estruturas de dados compartilhadas
   - `sync.Mutex` é usado extensivamente para proteger o acesso ao grafo de pontos e às filas de recarga
   - Exemplo em `grafo.go`: `mu sync.Mutex` protege o acesso ao mapa de nós

2. **Canais**: Usados para comunicação entre goroutines
   - O canal `notify` em `FilaRecarga` sinaliza quando um ponto fica disponível

3. **Goroutines**: Permitem execução concorrente
   - Cada conexão TCP é tratada em uma goroutine separada no servidor nuvem
   - Operações de longa duração (como recarga) são executadas concorrentemente

4. **Padrões de Sincronização**:
   - O padrão worker pool é implementado implicitamente através das goroutines que gerenciam as filas
   - O algoritmo de Dijkstra é executado de forma thread-safe usando mutexes

## Fluxo de Operação

1. Os pontos de recarga se registram na nuvem
2. Os carros consomem bateria enquanto se movimentam
3. Quando a bateria está baixa, os carros solicitam recarga à nuvem
4. A nuvem encontra o ponto mais próximo disponível usando Dijkstra
5. Se não houver pontos disponíveis, o carro é colocado em uma fila
6. Quando um ponto fica disponível, o próximo carro na fila é notificado

## Monitoramento e Logs

O sistema gera logs detalhados para cada componente:
- Carros registram consumo de bateria e operações de recarga
- Pontos registram status de disponibilidade
- A nuvem registra todas as operações de gerenciamento

Os dados de recarga são armazenados em arquivos JSON no volume compartilhado `dados_recarga`.

## Personalização

Você pode ajustar o número de carros e pontos modificando o arquivo `docker-compose.yml`. Para alterar parâmetros como consumo de bateria ou tempos de recarga, edite as constantes em `carro.go`.


## **Alunos(as)**

<table align='center'>
<tr> 
  <td align="center">
    <a href="https://github.com/LuisMarioRC">
      <img src="https://avatars.githubusercontent.com/u/142133059?v=4" width="100px;" alt=""/>
    </a>
    <br /><sub><b><a href="https://github.com/LuisMarioRC">Luis Mario</a></b></sub><br />👨💻
  </td>
  <td align="center">
    <a href="https://github.com/laizagordiano">
      <img src="https://avatars.githubusercontent.com/u/132793645?v=4" width="100px;" alt=""/>
    </a>
    <br /><sub><b><a href="https://github.com/laizagordiano">Laiza Gordiano</a></b></sub><br />👨💻
  </td>
  <td align="center">
    <a href="https://github.com/GHenryssg">
      <img src=https://avatars.githubusercontent.com/u/142272107?v=4" width="100px;" alt=""/>
    </a>
    <br /><sub><b><a href="https://github.com/GHenryssg">Gabriel Henry</a></b></sub><br />👨💻
  </td>
</tr>

</table>


<img width=100% src="https://capsule-render.vercel.app/api?type=waving&color=00FFFF&height=120&section=footer"/>

<div align="center"> 