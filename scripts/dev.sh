#!/bin/bash

# Script para desarrollo local
# Este script configura el entorno de desarrollo y ejecuta la aplicación

echo "🚀 Iniciando Backend Residuos App..."

# Verificar si existe el archivo .env
if [ ! -f ".env" ]; then
    echo "⚠️  Archivo .env no encontrado. Copiando desde .env.example..."
    cp .env.example .env
    echo "✅ Archivo .env creado. Por favor configura las variables antes de continuar."
    exit 1
fi

# Verificar si existe el archivo de credenciales de Firebase
FIREBASE_CREDS=$(grep FIREBASE_CREDENTIALS_PATH .env | cut -d '=' -f2)
if [ ! -f "$FIREBASE_CREDS" ]; then
    echo "⚠️  Archivo de credenciales de Firebase no encontrado: $FIREBASE_CREDS"
    echo "   Por favor descarga el archivo desde la consola de Firebase"
    echo "   y configura FIREBASE_CREDENTIALS_PATH en tu archivo .env"
    exit 1
fi

# Instalar/actualizar dependencias
echo "📦 Instalando dependencias..."
go mod tidy

# Compilar aplicación
echo "🔨 Compilando aplicación..."
go build -o bin/app cmd/api/main.go

if [ $? -eq 0 ]; then
    echo "✅ Compilación exitosa"
    echo "🌟 Ejecutando aplicación..."
    ./bin/app
else
    echo "❌ Error de compilación"
    exit 1
fi