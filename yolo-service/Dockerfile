# Imagem base mais completa com Python
FROM python:3.11-bullseye

# Evita buffering e define diretório
ENV PYTHONUNBUFFERED=1
WORKDIR /app

# Copia as dependências
COPY requirements.txt .

# Instala dependências do sistema necessárias pro torch, opencv e PIL
RUN apt-get update && apt-get install -y \
    libgl1 \
    libglib2.0-0 \
    build-essential \
    git \
 && rm -rf /var/lib/apt/lists/*

# Instala dependências do Python
RUN pip install --no-cache-dir -r requirements.txt
RUN pip install python-multipart
# Copia o código
COPY app/ ./app
COPY model/ ./model

EXPOSE 7676

# Comando para iniciar o servidor FastAPI
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "7676"]
