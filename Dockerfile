FROM python:latest

WORKDIR /app

COPY . .

RUN pip install --no-cache-dir torch torchvision torchaudio accelerate transformers

CMD [ "python", "./main.py" ]
