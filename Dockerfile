# Use uma imagem base do Golang
FROM golang:1.17

# Defina o diretório de trabalho dentro do contêiner
WORKDIR /app

# Copie todos os arquivos do diretório atual para o diretório de trabalho dentro do contêiner
COPY . .

# Baixe as dependências
RUN go mod tidy

# Compile o aplicativo Go
RUN go build -o main ./cmd/server

# Comando para executar o aplicativo
CMD ["./main"]
