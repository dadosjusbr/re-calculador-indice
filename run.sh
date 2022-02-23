#!/bin/bash

####################################################################
#
# Autor      : Dadosjusbr <dadosjusbr@gmail.com>
# Site       : https://dadosjusbr.org/
# Licença    : MIT
# Descrição  : Executa o re-calculador de índice de transparência
# Projeto    : https://github.com/dadosjusbr/re-calculador-indice
#
####################################################################


# Pega o nome de todos os órgãos e anos, passados nos arquivos .txt
aids="${aids:=$(cat ./aids.txt)}"
years="${years:=$(cat ./years.txt)}"

for aid in ${aids[@]}; do
  for year in ${years[@]}; do
    go run main.go --aid=$aid --year=$year
  done
done