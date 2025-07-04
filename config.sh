#!/bin/bash

# DB
DB_IMAGE_NAME="spaceresearch-db"
DB_CONTAINER_NAME="spaceresearch-db-1"
DB_PWD="ChangeMe" # Don't add spaces and be careful with special symbols
DB_SCHEMA="TF"
DB_USER="user"
DB_PORT="5432"

# BE
BE_IMAGE_NAME="spaceresearch-be"
BE_CONTAINER_NAME="spaceresearch-be-1"
BE_HOST="spaceresearch.gchalakov.com"
BE_PORT="8080"
ALLOWED_WS_ORIGIN="https://spaceresearch.gchalakov.com"
OLLAMA_API="http://192.168.0.27:11434/api/chat"
OLLAMA_MODEL="qwen3:4b"
OLLAMA_CTX="32768"
OLLAMA_EMB_API="http://192.168.0.27:11434/api/embeddings"
OLLAMA_EMB_MODEL="nomic-embed-text"
