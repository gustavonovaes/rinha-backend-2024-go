# Submissão para Rinha de Backend, Segunda Edição: 2024/Q1 - Controle de Concorrência

## Stack

- `Nginx 1.25` Load balance
- `PostgreSQL 16` Banco de Dados
- `Go 1.22` API

<br/>

<div style="display:flex; vertical-align:middle; align-itens:center;">
    <img src="https://www.vectorlogo.zone/logos/nginx/nginx-ar21.svg" alt="logo nginx" height="50" width="auto" style="padding-right: 1rem;">
    <img src="https://www.vectorlogo.zone/logos/postgresql/postgresql-ar21.svg" alt="logo postgresql" height="50" width="auto" style="padding-right: 1rem;">
    <img src="https://www.vectorlogo.zone/logos/golang/golang-ar21.svg" alt="logo go" height="50" width="auto">
</div>


## Rodando aplicação

Dev mode

```
docker-compose -f docker-compose.dev.yml up -d

curl -X POST http://localhost:9999/clientes/1/transacoes \
    --data '{"valor":42, "tipo":"c", "descricao":"Marvin"}'
```

Completo

```
docker-compose up -d --build

curl -X POST http://localhost:9999/clientes/1/transacoes \
    --data '{"valor":42, "tipo":"c", "descricao":"Marvin"}'
```



## Rodando testes

Unitário e Integração

```
docker-compose -f docker-compose.dev.yml run --rm api go test .
```


## Gatling report

```
================================================================================
---- Global Information --------------------------------------------------------
> request count                                      61503 (OK=61503  KO=0     )
> min response time                                      0 (OK=0      KO=-     )
> max response time                                    103 (OK=103    KO=-     )
> mean response time                                     2 (OK=2      KO=-     )
> std deviation                                          2 (OK=2      KO=-     )
> response time 50th percentile                          2 (OK=2      KO=-     )
> response time 75th percentile                          3 (OK=3      KO=-     )
> response time 95th percentile                          5 (OK=5      KO=-     )
> response time 99th percentile                          6 (OK=6      KO=-     )
> mean requests/sec                                251.033 (OK=251.033 KO=-     )
---- Response Time Distribution ------------------------------------------------
> t < 800 ms                                         61503 (100%)
> 800 ms <= t < 1200 ms                                  0 (  0%)
> t >= 1200 ms                                           0 (  0%)
> failed                                                 0 (  0%)
================================================================================
```