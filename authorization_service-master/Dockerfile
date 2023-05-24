FROM python:3.8-slim

WORKDIR sre_auth
COPY requirements.txt requirements.txt
RUN pip3 install --upgrade pip && pip3 install -U -r requirements.txt

COPY main.py main.py
COPY misc.py misc.py

CMD ["python3", "-u", "main.py"]
