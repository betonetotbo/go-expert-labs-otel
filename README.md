# go-expert-labs-otel

Repositório sobre o desafio de Labs da pós graduação **go-expert** da Full Cycle.

## Desafio

Veja mais detalhes [aqui](TODO.md).

## Como executar

Basicamente todo o ambiente é configurável pelo [Makefile](Makefile).

> **ATENÇÃO**
> 
> Antes de executar esta aplicação você deve obter uma **chave de API** no https://www.weatherapi.com/.
> 
> Crie então o arquivo `Makefile.env` na raiz desde repositório e dentro coloque a chave neste formato:
> `WEATHER_API_KEY=XXX` (aonde `XXX` é a chave de API obtida).

Para iniciar a aplicação execute:

```bash
make build
make up
```

Para parar a execução, execute:

```bash
make down
```

## O ambiente

O docker compose deste desafio irá alocar as seguintes portas:

* Interface do jaeger: http://localhost:16686
* Serviço de entrada: http://localhost:8080
* Serviço orquestrador: http://localhost:8081

## Como fazer requisições

Realizar um `POST` para o serviço de entrada passando um json com o `cep` desejado: 

```bash
curl -v http://localhost:8080 -d '{"cep":"01001000"}'
```

Depois da requisição você pode consultar o Jaeger para visualizar o tracing.