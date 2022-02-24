# Re-calculador de índice de transparência do DadosJusBr

## Exemplo de rotinas

para atualização do índice de um órgão em um ano:

```
go run main.go --aid=tjpb --year=2021
```
para atualização em lote:
```
./run.sh
```
Este último comando irá considerar os órgãos listados no arquivo `aids.txt` e os anos em `years.txt`.
