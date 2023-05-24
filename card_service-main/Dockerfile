FROM python:3.8-slim

WORKDIR card_service
COPY requirements.txt requirements.txt
RUN pip3 install --upgrade pip && pip3 install -U -r requirements.txt

COPY src src/
COPY src/main.py main.py
COPY src/health_check.py health_check.py

COPY docker-entrypoint.sh docker-entrypoint.sh

RUN chmod +x docker-entrypoint.sh

CMD ["./docker-entrypoint.sh"]
