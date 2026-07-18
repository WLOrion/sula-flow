# Sula Flow - Football Transfer Network

Análise de redes de transferências de jogadores de futebol entre América do Sul e Europa, identificando rotas e clubes intermediários no mercado internacional.

## 📋 Requisitos

- Python 3.8+
- Neo4j 4.4+
- Docker (opcional)

## 🚀 Instalação

### Local
```bash
# Clone o repositório
git clone https://github.com/WLOrion/sula-flow.git
cd sula-flow

# Instale as dependências
pip install -r requirements.txt

# Configure variáveis de ambiente
cp .env.example .env
```

### Docker
```bash
docker-compose up -d
```

## 🗄️ Configuração do Grafo

### Carregar dados no Neo4j
```bash
# Importar o grafo
python scripts/load_graph.py --file data/transfers.csv

# Ou usar o dump disponível
neo4j-admin load --from=data/graph-dump.db --database=neo4j
```

### Estrutura do Grafo
- **Nodes**: `Club`, `Player`, `Country`, `Transfer`
- **Relationships**: `HAS_PLAYER`, `FROM_CLUB`, `TO_CLUB`, `CONNECTED_TO`

## 🔌 API Endpoints

### Base URL: `http://localhost:8000/api`

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET | `/clubs/{id}` | Detalhes de um clube |
| GET | `/clubs/{id}/transfers` | Transferências de um clube |
| GET | `/players/{id}/trajectory` | Trajetória de um jogador |
| GET | `/routes/south-america-europe` | Principais rotas SA → EU |
| GET | `/analytics/intermediaries` | Clubes intermediários |
| GET | `/analytics/centrality` | Métricas de centralidade |

## 📊 Queries Úteis

### Clubes mais conectados
```cypher
MATCH (c:Club)
RETURN c.name, c.country, 
       size((c)-[:CONNECTED_TO]-()) as connections
ORDER BY connections DESC
LIMIT 10
```

### Rotas Sul América → Europa
```cypher
MATCH path = (c1:Club)-[:CONNECTED_TO*1..3]->(c2:Club)
WHERE c1.continent = 'South America' 
  AND c2.continent = 'Europe'
RETURN path
LIMIT 100
```

### Jogadores com mais transferências
```cypher
MATCH (p:Player)-[:HAS_PLAYER]->(t:Transfer)
RETURN p.name, count(t) as transfers
ORDER BY transfers DESC
LIMIT 20
```

### Clubes intermediários (betweenness)
```cypher
CALL gds.betweenness.stream('transfer-graph')
YIELD nodeId, score
RETURN gds.util.asNode(nodeId).name AS club, score
ORDER BY score DESC
LIMIT 10
```

## 🛠️ Scripts Disponíveis

```bash
# Coletar dados do Transfermarkt
python scripts/scraper.py --years 2015-2025

# Processar e limpar dados
python scripts/process_data.py

# Gerar visualizações
python scripts/visualize.py --type flow-map

# Executar análises
python scripts/analyze.py --metric all
```

## 📁 Estrutura

```
sula-flow/
├── api/              # API REST
├── data/             # Datasets e dumps
├── scripts/          # Scripts de ETL e análise
├── queries/          # Queries Cypher prontas
├── notebooks/        # Jupyter notebooks
└── docker-compose.yml
```

## 🎯 Métricas Disponíveis

- **Grau**: Conexões diretas de cada clube
- **Betweenness**: Clubes que servem de ponte
- **PageRank**: Importância global na rede
- **Comunidades**: Detecção de grupos de clubes

## 🔗 Links Úteis

- [Neo4j Browser](http://localhost:7474)
- [API Docs](http://localhost:8000/docs)
- [Jupyter Lab](http://localhost:8888)

## 📝 Licença

MIT