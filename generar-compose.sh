#!/bin/bash

if [ $# -ne 2 ]; then
  echo "Uso: $0 <nombre_archivo_salida> <cantidad_clientes>"
  exit 1
fi

NOMBRE_ARCHIVO=$1
CANTIDAD_CLIENTES=$2

python3 mi-generador.py $NOMBRE_ARCHIVO $CANTIDAD_CLIENTES