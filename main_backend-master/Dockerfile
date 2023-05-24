FROM python:3.8-slim

WORKDIR main_backend
COPY requirements.txt requirements.txt
RUN pip3 install --upgrade pip && pip3 install -U -r requirements.txt

COPY src src/
COPY src/main.py main.py

CMD ["python3", "-u", "main.py"]
