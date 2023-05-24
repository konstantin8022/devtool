FROM python:3.8-slim

WORKDIR api_gateway
COPY requirements.txt requirements.txt
RUN pip3 install --upgrade pip && pip3 install -U -r requirements.txt
RUN apt update && apt install -y libzstd1

COPY src src/
COPY src/main.py main.py
COPY docker-entrypoint.sh docker-entrypoint.sh

RUN chmod +x docker-entrypoint.sh

CMD ["./docker-entrypoint.sh"]
